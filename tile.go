package show

// https://github.com/victorspringer/http-cache

import (
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	// "reflect"
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

// GetFeaturesForTileFunc returns a `mvt.GetFeaturesCallbackFunc` callback function using 'db' and 'database' to yield
// a dictionary of GeoJSON FeatureCollections instances.
func GetFeaturesForTileFunc(db *sql.DB, datasource string, table_cols []string) mvt.GetFeaturesCallbackFunc {

	// quoted_cols wraps each column name in double-quotes
	quoted_cols := make([]string, 0)

	// pointer_cols is a list of column names we use to construct an array of pointers
	// to indices to an array of values (below) that database column values will be written
	// in to – this is a bit of unfortunate hoop-jumping that is necessary
	// to account for the way that database/sql "scans" column data
	// in to variables.
	pointer_cols := make([]string, 0)

	for _, c := range table_cols {
		switch c {
		case "geometry":
			// Note: We are treatin the geometry columns as a special case
			// in the SQL query below.
		default:
			quoted_cols = append(quoted_cols, fmt.Sprintf(`"%s"`, c))
			pointer_cols = append(pointer_cols, c)
		}
	}

	// But wait, there's more! Append the "geometry" column back in to pointer_cols
	// so that it is included in the pointers/values.
	pointer_cols = append(pointer_cols, "geometry")

	str_cols := strings.Join(quoted_cols, ",")

	fn := func(req *http.Request, layer string, t *maptile.Tile) (map[string]*geojson.FeatureCollection, error) {

		logger := slog.Default()
		logger = logger.With("path", req.URL.Path)

		feature_count := 0
		t1 := time.Now()

		defer func() {
			logger.Debug("Time to get features", "count", feature_count, "time", time.Since(t1))
		}()

		ctx := req.Context()

		bound := t.Bound()
		poly := bound.ToPolygon()

		enc_poly, err := wkb.MarshalToHex(poly, wkb.DefaultByteOrder)

		if err != nil {
			return nil, err
		}

		// Note: Do not change the order of columns here (geometry at the end) without adjusting
		// pointer_cols above.

		q := fmt.Sprintf(`SELECT
				%s, ST_AsText(ST_GeomFromWkb(geometry)) AS geometry
			  FROM
				read_parquet("%s")
			  WHERE
				ST_Intersects(ST_GeomFromWkb(geometry), ST_GeomFromHEXWKB(?))`,
			str_cols, datasource)

		// slog.Debug(q)

		rows, err := db.QueryContext(ctx, q, string(enc_poly))

		if err != nil {
			slog.Error("Failed to query database", "error", err, "query", q, "geom", enc_poly)
			return nil, fmt.Errorf("Failed to query database, %w", err)
		}

		defer rows.Close()

		fc := geojson.NewFeatureCollection()

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
				slog.Error("Failed to scan row", "error", err)
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
			feature_count += 1
		}

		err = rows.Err()

		if err != nil {
			return nil, fmt.Errorf("There was a problem scanning rows, %w", err)
		}

		collections := map[string]*geojson.FeatureCollection{
			layer: fc,
		}

		return collections, nil
	}

	return fn
}
