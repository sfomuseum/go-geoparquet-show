package show

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/sfomuseum/go-geoparquet-show/static/www"
	"github.com/sfomuseum/go-http-mvt"
	www_show "github.com/sfomuseum/go-www-show"
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

		slog.Debug("Column definition", "name", col_name, "type", col_type)
		table_cols = append(table_cols, col_name)
	}

	err = rows.Err()

	if err != nil {
		return fmt.Errorf("There was a problem scanning rows, %w", err)
	}

	// END OF get table defs

	mux := http.NewServeMux()

	www_fs := http.FS(www.FS)
	mux.Handle("/", http.FileServer(www_fs))

	features_cb := GetFeaturesForTileFunc(opts.Database, opts.Datasource, table_cols)

	mvt_opts := &mvt.TileHandlerOptions{
		GetFeaturesCallback: features_cb,
		Simplify:            true,
	}

	mvt_handler, err := mvt.NewTileHandler(mvt_opts)

	if err != nil {
		return err
	}

	mux.Handle("/tiles/", mvt_handler)

	www_show_opts := &www_show.RunOptions{
		Port:    opts.Port,
		Mux:     mux,
		Browser: opts.Browser,
	}

	return www_show.RunWithOptions(ctx, www_show_opts)
}
