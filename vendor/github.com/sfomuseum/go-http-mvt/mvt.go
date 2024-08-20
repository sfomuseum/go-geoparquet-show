package mvt

// https://github.com/victorspringer/http-cache

import (
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"strconv"
	"time"

	orb_mvt "github.com/paulmach/orb/encoding/mvt"
	"github.com/paulmach/orb/geojson"
	"github.com/paulmach/orb/maptile"
	"github.com/paulmach/orb/simplify"
)

var re_path = regexp.MustCompile(`/.*/(.*)/(\d+)/(\d+)/(\d+).(\w+)$`)

type GetFeaturesCallbackFunc func(*http.Request, string, *maptile.Tile) (map[string]*geojson.FeatureCollection, error)

type TileHandlerOptions struct {
	GetFeaturesCallback GetFeaturesCallbackFunc
	Simplify            bool
	Timings             bool
}

func NewTileHandler(opts *TileHandlerOptions) (http.Handler, error) {

	fn := func(rsp http.ResponseWriter, req *http.Request) {

		logger := slog.Default()
		logger = logger.With("path", req.URL.Path)

		if opts.Timings {

			t1 := time.Now()

			defer func() {
				logger.Debug("Time to process tile", "time", time.Since(t1))
			}()
		}

		layer, t, err := getTileForRequest(req)

		if err != nil {
			logger.Error("Failed to get tile for request", "error", err)
			http.Error(rsp, "Internal server error", http.StatusInternalServerError)
			return
		}

		logger = logger.With("layer", layer)

		collections, err := opts.GetFeaturesCallback(req, layer, t)

		if err != nil {
			logger.Error("Failed to get data for tile", "error", err)
			http.Error(rsp, "Internal server error", http.StatusInternalServerError)
			return
		}

		if opts.Timings {

			t2 := time.Now()

			defer func() {
				logger.Debug("Time to yield MVT", "time", time.Since(t2))
			}()
		}

		layers := orb_mvt.NewLayers(collections)
		layers.ProjectToTile(*t)

		layers.Clip(orb_mvt.MapboxGLDefaultExtentBound)

		if opts.Simplify {
			layers.Simplify(simplify.DouglasPeucker(1.0))
		}

		layers.RemoveEmpty(1.0, 1.0)

		data, err := orb_mvt.Marshal(layers)

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

func getTileForRequest(req *http.Request) (string, *maptile.Tile, error) {

	path := req.URL.Path

	if !re_path.MatchString(path) {
		return "", nil, fmt.Errorf("Invalid path")
	}

	m := re_path.FindStringSubmatch(path)

	layer := m[1]

	z, err := strconv.Atoi(m[2])

	if err != nil {
		return "", nil, fmt.Errorf("Invalid {z} parameter, %w", err)
	}

	x, err := strconv.Atoi(m[3])

	if err != nil {
		return "", nil, fmt.Errorf("Invalid {x} parameter, %w", err)
	}

	y, err := strconv.Atoi(m[4])

	if err != nil {
		return "", nil, fmt.Errorf("Invalid {y} parameter, %w", err)
	}

	zm := maptile.Zoom(uint32(z))

	t := &maptile.Tile{
		Z: zm,
		X: uint32(x),
		Y: uint32(y),
	}

	return layer, t, nil
}
