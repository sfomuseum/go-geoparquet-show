package show

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"strconv"
	// "strings"
	"time"

	"github.com/paulmach/orb/encoding/mvt"
	"github.com/paulmach/orb/encoding/wkb"
	"github.com/paulmach/orb/encoding/wkt"
	"github.com/paulmach/orb/geojson"
	"github.com/paulmach/orb/maptile"
	"github.com/paulmach/orb/simplify"
)

var re_path = regexp.MustCompile(`/.*/(.*)/(\d+)/(\d+)/(\d+).(\w+)$`)

type TileHandlerOptions struct {
	Database   *sql.DB
	Datasource string
}

func NewTileHandler(opts *TileHandlerOptions) (http.Handler, error) {

	fn := func(rsp http.ResponseWriter, req *http.Request) {

		ctx := req.Context()
		ctx, cancel := context.WithCancel(ctx)

		defer cancel()

		logger := slog.Default()
		logger = logger.With("path", req.URL.Path)

		t1 := time.Now()

		defer func() {
			logger.Debug("Time to process tile", "time", time.Since(t1))
		}()

		layer, t, err := getTileForRequest(req)

		if err != nil {
			logger.Error("Failed to get tile for request", "error", err)
			http.Error(rsp, "Internal server error", http.StatusInternalServerError)
			return
		}

		logger = logger.With("layer", layer)

		fc, err := getFeaturesForTile(req, opts, layer, t)

		if err != nil {
			logger.Error("Failed to get data for tile", "error", err)
			http.Error(rsp, "Internal server error", http.StatusInternalServerError)
			return
		}

		t2 := time.Now()

		defer func() {
			logger.Debug("Time to yield MVT", "time", time.Since(t2))
		}()

		collections := map[string]*geojson.FeatureCollection{
			layer: fc,
		}

		layers := mvt.NewLayers(collections)
		layers.ProjectToTile(*t)

		layers.Clip(mvt.MapboxGLDefaultExtentBound)

		layers.Simplify(simplify.DouglasPeucker(1.0))
		layers.RemoveEmpty(1.0, 1.0)

		data, err := mvt.Marshal(layers)

		if err != nil {
			logger.Error("Failed to marshal layers", "error", err)
			http.Error(rsp, "Internal server error", http.StatusInternalServerError)
			return
		}

		rsp.Header().Set("Content-Type", "application/vnd.mapbox-vector-tile")
		rsp.Write(data)
		return
	}

	return http.HandlerFunc(fn), nil
}

func getFeaturesForTile(req *http.Request, opts *TileHandlerOptions, layer string, t *maptile.Tile) (*geojson.FeatureCollection, error) {

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
		opts.Datasource)

	rows, err := opts.Database.QueryContext(ctx, q, string(enc_poly))

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

		/*
			if strings.HasPrefix(wkt_geom, "MULTIPOINT ("){
				wkt_geom = strings.Replace(wkt_geom, "MULTIPOINT (", "MULTIPOINT(", 1)
			}
		*/

		orb_geom, err := wkt.Unmarshal(wkt_geom)

		if err != nil {
			logger.Error("Failed to unmarshal geometry", "id", id, "geom", wkt_geom, "error", err)
			continue

			// return nil, fmt.Errorf("Failed to unmarshal geometry, %w", err)
		}

		f := geojson.NewFeature(orb_geom)
		f.Properties["wof:id"] = id
		f.Properties["wof:name"] = name

		fc.Append(f)
		feature_count += 1
	}

	err = rows.Err()

	if err != nil {
		// return nil fmt.Errorf("There was a problem scanning rows, %w", err)
		return nil, err
	}

	return fc, nil
}

func getTileForRequest(req *http.Request) (string, *maptile.Tile, error) {

	path := req.URL.Path

	if !re_path.MatchString(path) {
		return "", nil, fmt.Errorf("Invalid path")
	}

	m := re_path.FindStringSubmatch(path)

	layer := m[1]

	z, err := strconv.Atoi(m[2])

	if err != nil {
		return "", nil, err
	}

	x, err := strconv.Atoi(m[3])

	if err != nil {
		return "", nil, err
	}

	y, err := strconv.Atoi(m[4])

	if err != nil {
		return "", nil, err
	}

	zm := maptile.Zoom(uint32(z))

	t := &maptile.Tile{
		Z: zm,
		X: uint32(x),
		Y: uint32(y),
	}

	return layer, t, nil
}
