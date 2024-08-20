package show

// https://github.com/victorspringer/http-cache

import (
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
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

func GetFeaturesForTileFunc(db *sql.DB, datasource string) mvt.GetFeaturesCallbackFunc {

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

		q := fmt.Sprintf(`SELECT
				"wof:id","wof:name",
				ST_AsText(ST_GeomFromWkb(geometry)) AS geometry
			  FROM
				read_parquet("%s")
			  WHERE
				ST_Intersects(ST_GeomFromWkb(geometry), ST_GeomFromHEXWKB(?))`,
			datasource)

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

			var id float64
			var name string
			var wkt_geom string

			err := rows.Scan(&id, &name, &wkt_geom)

			if err != nil {
				slog.Error("Failed to scan row", "error", err)
				return nil, fmt.Errorf("Failed to scan row, %w", err)
			}

			// See notes above
			if strings.HasPrefix(wkt_geom, "MULTIPOINT (") {
				wkt_geom = fixMultiPoint(wkt_geom)
			}

			// To do:
			// GEOMETRYCOLLECTION (

			orb_geom, err := wkt.Unmarshal(wkt_geom)

			if err != nil {
				logger.Error("Failed to unmarshal geometry", "id", id, "geom", wkt_geom, "error", err)
				continue
			}

			f := geojson.NewFeature(orb_geom)
			f.Properties["wof:id"] = id
			f.Properties["wof:name"] = name

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
