package show

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"
)

// RunWithOptions start a local webserver and then open its URL in a target source once it (the web server) is running.
func RunWithOptions(ctx context.Context, opts *RunOptions) error {

	port := opts.Port
	host := opts.Host

	if host == "" {
		host = "localhost"
	}

	if port == 0 {

		addr := fmt.Sprintf("%s:0", host)
		listener, err := net.Listen("tcp", addr)

		if err != nil {
			return fmt.Errorf("Failed to determine next available port %s, %w", err)
		}

		port = listener.Addr().(*net.TCPAddr).Port
		err = listener.Close()

		if err != nil {
			return fmt.Errorf("Failed to close listener used to derive port, %w", err)
		}
	}

	//

	addr := fmt.Sprintf("%s:%d", host, port)
	url := fmt.Sprintf("http://%s", addr)

	http_server := http.Server{
		Addr: addr,
	}

	http_server.Handler = opts.Mux

	done_ch := make(chan bool)
	err_ch := make(chan error)

	go func() {

		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint

		slog.Info("Received interupt signal, shutting server down")
		err := http_server.Shutdown(ctx)

		if err != nil {
			slog.Error("HTTP server shutdown error", "error", err)
		}

		slog.Debug("Close done channel")
		close(done_ch)
	}()

	go func() {

		slog.Debug("Start server")
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

	err := opts.Browser.OpenURL(ctx, url)

	if err != nil {
		return fmt.Errorf("Failed to open URL %s, %w", url, err)
	}

	slog.Info("Server is ready and features are viewable", "url", url)
	<-done_ch

	slog.Debug("Exiting")
	return nil
}
