package show

// https://github.com/victorspringer/http-cache

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"time"

	"github.com/paulmach/orb/encoding/wkb"
	"github.com/paulmach/orb/encoding/wkt"
	"github.com/paulmach/orb/geojson"
	"github.com/paulmach/orb/maptile"
	"github.com/sfomuseum/go-http-mvt"
)

// START OF The DuckDB spatial extension returns WKT-formatted MultiPoint strings
//  without enclosing bracketsfor individual points which makes Orb sad. See also:
// https://libgeos.org/specifications/wkt/

var re_wkt_point = regexp.MustCompile(`\-?\d+(?:\.\d+)? \-?\d+(?:\.\d+)?`)

func fixMultiPoint(wkt_geom string) string {
	return re_wkt_point.ReplaceAllStringFunc(wkt_geom, replaceMultiPoint)
}

func replaceMultiPoint(s string) string {
	return fmt.Sprintf("(%s)", s)
}

// END OF The DuckDB spatial extension returns WKT-formatted MultiPoint strings

// GetFeaturesForTileFuncOptions defines configuration details to pass the `GetFeaturesForTileFunc` method.
type GetFeaturesForTileFuncOptions struct {
	// A valid `sql.DB` instance (assumed for the time being to be using the "duckdb" engine).
	Database *sql.DB
	// A valid URI to a GeoParquet file to pass to the DuckDB `read_parquet` method.
	Datasource string
	// The list of table columns to query for and assign as GeoJSON properties.
	TableColumns []string
	// An option column name to use for a initial bounding box constraint. This columns is expected to contain the maximum X (longitude) value of the geometry it is associated with.
	MaxXColumn string
	// An option column name to use for a initial bounding box constraint. This columns is expected to contain the maximum Y (latitude) value of the geometry it is associated with.
	MaxYColumn string
}

// GetFeaturesForTileFunc returns a `mvt.GetFeaturesCallbackFunc` callback function using details specified in 'opts' to yield
// a dictionary of GeoJSON FeatureCollections instances.
func GetFeaturesForTileFunc(opts *GetFeaturesForTileFuncOptions) mvt.GetFeaturesCallbackFunc { // db *sql.DB, datasource string, table_cols []string) mvt.GetFeaturesCallbackFunc {

	// quoted_cols wraps each column name in double-quotes
	quoted_cols := make([]string, 0)

	// pointer_cols is a list of column names we use to construct an array of pointers
	// to indices to an array of values (below) that database column values will be written
	// in to – this is a bit of unfortunate hoop-jumping that is necessary
	// to account for the way that database/sql "scans" column data
	// in to variables.
	pointer_cols := make([]string, 0)

	for _, c := range opts.TableColumns {
		switch c {
		case "geometry":
			// Note: We are treating the geometry column as a special case
			// in the SQL query below.
		default:
			quoted_cols = append(quoted_cols, fmt.Sprintf(`"%s"`, c))
			pointer_cols = append(pointer_cols, c)
		}
	}

	// But wait, there's more! Append the "geometry" column back in to pointer_cols
	// so that it is included in the pointers/values.
	pointer_cols = append(pointer_cols, "geometry")

	// Generate a CSV string of quoted_cols for use with SQL queries below.
	str_cols := strings.Join(quoted_cols, ",")

	fn := func(ctx context.Context, layer string, t *maptile.Tile) (map[string]*geojson.FeatureCollection, error) {

		tile_key := fmt.Sprintf("%d/%d/%d", t.Z, t.X, t.Y)

		logger := slog.Default()
		logger = logger.With("layer", layer)
		logger = logger.With("tile", tile_key)

		fc := geojson.NewFeatureCollection()

		t1 := time.Now()

		defer func() {
			logger.Debug("Time to get features", "count", len(fc.Features), "time", time.Since(t1))
		}()

		bound := t.Bound()
		poly := bound.ToPolygon()

		enc_poly, err := wkb.MarshalToHex(poly, wkb.DefaultByteOrder)

		if err != nil {
			return nil, fmt.Errorf("Failed to marshal tile boundary to WKBHEX, %w", err)
		}

		// Note: Do not change the order of columns here (geometry at the end) without adjusting
		// pointer_cols above.

		where := make([]string, 0)
		args := make([]interface{}, 0)

		// START OF bbox constraint
		// It is not clear to me whether this has any meaningful impact on query times.

		if opts.MaxXColumn != "" && opts.MaxYColumn != "" {

			max_lon := opts.MaxXColumn
			max_lat := opts.MaxYColumn

			minx := bound.Min[0]
			miny := bound.Min[1]
			maxx := bound.Max[0]
			maxy := bound.Max[1]

			where_bbox := fmt.Sprintf(`(("%s" > ? AND "%s" < ?) OR ("%s" > ? AND "%s" < ?))`, max_lon, max_lon, max_lat, max_lat)
			where = append(where, where_bbox)

			args = append(args, minx)
			args = append(args, maxy)
			args = append(args, miny)
			args = append(args, maxx)
		}

		// END OF bbox constraint

		where = append(where, `ST_Intersects(ST_GeomFromWkb(geometry::WKB_BLOB), ST_GeomFromHEXWKB(?))`)
		args = append(args, string(enc_poly))

		str_where := strings.Join(where, " AND ")

		q := fmt.Sprintf(`SELECT %s, ST_AsText(ST_GeomFromWkb(geometry::WKB_BLOB)) AS geometry FROM read_parquet("%s") WHERE %s`,
			str_cols, opts.Datasource, str_where)

		/*
			logger.Debug(q)

			for i, a := range args {
				logger.Debug("arg", "offset", i, "value", a)
			}
		*/

		rows, err := opts.Database.QueryContext(ctx, q, args...)

		if err != nil {

			if errors.Is(err, context.Canceled) {
				return nil, nil
			}

			logger.Error("Failed to query database", "error", err, "query", q, "geom", enc_poly)
			return nil, fmt.Errorf("Failed to query database, %w", err)
		}

		defer rows.Close()

		for rows.Next() {

			select {
			case <-ctx.Done():
				break
			default:
				// pass
			}

			// START OF indirect all the things to satify db.Scan
			// See notes wrt/ pointer_cols above

			values := make([]any, len(pointer_cols))
			pointers := make([]any, len(pointer_cols))

			for idx, _ := range pointer_cols {
				pointers[idx] = &values[idx]
			}

			err := rows.Scan(pointers...)

			// END OF indirect all the things to satify db.Scan
			// Well not quite, there's a bit more below...

			if err != nil {

				if errors.Is(err, context.Canceled) {
					break
				}

				logger.Error("Failed to scan row", "error", err)
				return nil, fmt.Errorf("Failed to scan row, %w", err)
			}

			var wkt_geom string
			props := make(map[string]any)

			for idx, k := range pointer_cols {

				// Note: See the way we're reading from the values array even though
				// the DB layer "wrote" those values to the pointers array? That's
				// because we indirected all the things (above). Good times.

				switch k {
				case "geometry":
					wkt_geom = values[idx].(string)
				default:
					props[k] = values[idx]
				}
			}

			// See notes above
			if strings.HasPrefix(wkt_geom, "MULTIPOINT (") {
				wkt_geom = fixMultiPoint(wkt_geom)
			}

			// To do:
			// GEOMETRYCOLLECTION (

			orb_geom, err := wkt.Unmarshal(wkt_geom)

			if err != nil {
				logger.Error("Failed to unmarshal geometry", "geom", wkt_geom, "error", err)
				continue
			}

			f := geojson.NewFeature(orb_geom)
			f.Properties = props

			fc.Append(f)
		}

		err = rows.Err()

		if err != nil {

			if !errors.Is(err, context.Canceled) {
				return nil, fmt.Errorf("There was a problem scanning rows, %w", err)
			}
		}

		collections := map[string]*geojson.FeatureCollection{
			layer: fc,
		}

		return collections, nil
	}

	return fn
}
