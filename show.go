package show

import (
	"context"
	"flag"
	"fmt"
	_ "io"
	"log"
	"log/slog"
	"net"
	"net/http"
	_ "net/url"
	"os"
	"os/signal"
	_ "strings"
	"time"

	"github.com/sfomuseum/go-geoparquet-show/static/www"
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

	tile_opts := &TileHandlerOptions{
		Database:   opts.Database,
		Datasource: opts.Datasource,
	}

	tile_handler, err := NewTileHandler(tile_opts)

	if err != nil {
		return err
	}

	mux.Handle("/tiles/", tile_handler)

	// START OF merge with go-geojson-show and put in a package or something

	// funcName(ctx, port, FS, url)

	www_fs := http.FS(www.FS)
	mux.Handle("/", http.FileServer(www_fs))

	port := opts.Port

	if port == 0 {

		listener, err := net.Listen("tcp", "localhost:0")

		if err != nil {
			log.Fatalf("Failed to determine next available port, %v", err)
		}

		port = listener.Addr().(*net.TCPAddr).Port
		err = listener.Close()

		if err != nil {
			log.Fatalf("Failed to close listener used to derive port, %v", err)
		}
	}

	//

	addr := fmt.Sprintf("localhost:%d", port)
	url := fmt.Sprintf("http://%s", addr)

	http_server := http.Server{
		Addr: addr,
	}

	http_server.Handler = mux

	done_ch := make(chan bool)
	err_ch := make(chan error)

	go func() {

		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint

		slog.Info("Shutting server down")
		err := http_server.Shutdown(ctx)

		if err != nil {
			slog.Error("HTTP server shutdown error", "error", err)
		}

		close(done_ch)
	}()

	go func() {

		err := http_server.ListenAndServe()

		if err != nil {
			err_ch <- fmt.Errorf("Failed to start server, %w", err)
		}
	}()

	server_ready := false

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case err := <-err_ch:
			log.Fatalf("Received error starting server, %v", err)
		case <-ticker.C:

			rsp, err := http.Head(url)

			if err != nil {
				slog.Warn("HEAD request failed", "url", url, "error", err)
			} else {

				defer rsp.Body.Close()

				if rsp.StatusCode != 200 {
					slog.Warn("HEAD request did not return expected status code", "url", url, "code", rsp.StatusCode)
				} else {
					slog.Debug("HEAD request succeeded", "url", url)
					server_ready = true
				}
			}
		}

		if server_ready {
			break
		}
	}

	/*
		err := opts.Browser.OpenURL(ctx, url)

		if err != nil {
			log.Fatalf("Failed to open URL %s, %v", url, err)
		}
	*/

	log.Printf("Features are viewable at %s\n", url)
	<-done_ch

	// END OF merge with go-geojson-show and put in a package or something

	return nil

}
