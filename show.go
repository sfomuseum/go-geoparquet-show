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

func Run(ctx context.Context) error {
	fs := DefaultFlagSet()
	return RunWithFlagSet(ctx, fs)
}

func RunWithFlagSet(ctx context.Context, fs *flag.FlagSet) error {

	opts, err := RunOptionsFromFlagSet(ctx, fs)

	if err != nil {
		return err
	}

	return RunWithOptions(ctx, opts)
}

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

	mux := http.NewServeMux()

	www_fs := http.FS(www.FS)
	mux.Handle("/", http.FileServer(www_fs))

	features_cb := GetFeaturesForTileFunc(opts.Database, opts.Datasource)

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
