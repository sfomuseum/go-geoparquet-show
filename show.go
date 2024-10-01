package show

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/sfomuseum/go-geoparquet-show/static/www"
	"github.com/sfomuseum/go-http-mvt"
	www_show "github.com/sfomuseum/go-www-show/v2"
)

// Run with launch a web server and browser serving GeoParquet data as vector tiles using the default flag set.
func Run(ctx context.Context) error {
	fs := DefaultFlagSet()
	return RunWithFlagSet(ctx, fs)
}

// Run with launch a web server and browser serving GeoParquet data as vector tiles using options derived from 'fs'
func RunWithFlagSet(ctx context.Context, fs *flag.FlagSet) error {

	opts, err := RunOptionsFromFlagSet(ctx, fs)

	if err != nil {
		return err
	}

	return RunWithOptions(ctx, opts)
}

// Run with launch a web server and browser serving GeoParquet data as vector tiles using configuration details provided by 'opts'
func RunWithOptions(ctx context.Context, opts *RunOptions) error {

	if opts.Verbose {
		slog.SetLogLoggerLevel(slog.LevelDebug)
		slog.Debug("Verbose logging enabled")
	}

	map_cfg := &mapConfig{
		LabelProperties: opts.LabelProperties,
		Renderer:        opts.Renderer,
	}

	if len(opts.LabelProperties) > 0 && opts.Renderer == "leaflet" {
		slog.Warn("Rendering label properties in Leaflet maps is currently disabled.")
	}

	// START OF set up database

	setup := []string{
		"INSTALL spatial",
		"LOAD spatial",
	}

	for _, q := range setup {

		_, err := opts.Database.ExecContext(ctx, q)

		if err != nil {
			return fmt.Errorf("Database setup command (%s) failed, %w", q, err)
		}
	}

	table_cols := make([]string, 0)

	// START OF get table defs

	// Update to use https://www.markhneedham.com/blog/2024/09/22/duckdb-dynamic-column-selection/

	q := fmt.Sprintf(`DESCRIBE SELECT * FROM read_parquet("%s")`, opts.Datasource)

	rows, err := opts.Database.QueryContext(ctx, q)

	if err != nil {
		slog.Error("Failed to query database", "error", err, "query", q)
		return fmt.Errorf("Failed to query database, %w", err)
	}

	defer rows.Close()

	for rows.Next() {

		var col_name string
		var col_type string
		var col_null any
		var col_key any
		var col_default any
		var col_extra any

		err := rows.Scan(&col_name, &col_type, &col_null, &col_key, &col_default, &col_extra)

		if err != nil {
			slog.Error("Failed to scan row", "error", err)
			return fmt.Errorf("Failed to scan row, %w", err)
		}

		// slog.Debug("Column definition", "name", col_name, "type", col_type)
		table_cols = append(table_cols, col_name)
	}

	err = rows.Err()

	if err != nil {
		return fmt.Errorf("There was a problem scanning rows, %w", err)
	}

	// END OF get table defs

	// START OF feature(s) extent

	extent_q := fmt.Sprintf(`SELECT MIN(ST_XMin(ST_GeomFromWKB(geometry::WKB_BLOB))) AS minx, MIN(ST_YMin(ST_GeomFromWKB(geometry::WKB_BLOB))) AS miny, MAX(ST_Xmax(ST_GeomFromWKB(geometry::WKB_BLOB))) AS maxx, MAX(ST_YMax(ST_GeomFromWKB(geometry::WKB_BLOB))) AS maxy FROM read_parquet("%s")`, opts.Datasource)

	extent_row := opts.Database.QueryRowContext(ctx, extent_q)

	var minx float64
	var miny float64
	var maxx float64
	var maxy float64

	err = extent_row.Scan(&minx, &miny, &maxx, &maxy)

	if err != nil {
		return fmt.Errorf("Failed to derive database extent, %w", err)
	}

	map_cfg.MinX = minx
	map_cfg.MinY = miny
	map_cfg.MaxX = maxx
	map_cfg.MaxY = maxy

	// END OF feature(s) extent

	mux := http.NewServeMux()

	www_fs := http.FS(www.FS)
	mux.Handle("/", http.FileServer(www_fs))

	map_cfg_handler := mapConfigHandler(map_cfg)
	mux.Handle("/map.json", map_cfg_handler)

	// https://github.com/sfomuseum/go-http-mvt

	features_opts := &GetFeaturesForTileFuncOptions{
		Database:     opts.Database,
		Datasource:   opts.Datasource,
		TableColumns: table_cols,
		MaxXColumn:   opts.MaxXColumn,
		MaxYColumn:   opts.MaxYColumn,
	}

	features_cb := GetFeaturesForTileFunc(features_opts)

	mvt_opts := &mvt.TileHandlerOptions{
		GetFeaturesCallback: features_cb,
		Simplify:            true,
	}

	mvt_handler, err := mvt.NewTileHandler(mvt_opts)

	if err != nil {
		return err
	}

	// https://github.com/victorspringer/http-cache/
	// Initial tests suggest this still has problems
	// (Whole zoom levels getting dropped for example)

	mux.Handle("/tiles/", mvt_handler)

	// https://github.com/sfomuseum/go-www-show

	www_show_opts := &www_show.RunOptions{
		Port:    opts.Port,
		Mux:     mux,
		Browser: opts.Browser,
	}

	return www_show.RunWithOptions(ctx, www_show_opts)
}

func mapConfigHandler(cfg *mapConfig) http.Handler {

	fn := func(rsp http.ResponseWriter, req *http.Request) {

		rsp.Header().Set("Content-type", "application/json")

		enc := json.NewEncoder(rsp)
		err := enc.Encode(cfg)

		if err != nil {
			slog.Error("Failed to encode map config", "error", err)
			http.Error(rsp, "Internal server error", http.StatusInternalServerError)
		}

		return
	}

	return http.HandlerFunc(fn)
}
