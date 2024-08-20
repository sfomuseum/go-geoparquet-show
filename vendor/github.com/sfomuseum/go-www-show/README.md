# go-www-show

Go package for starting a local	webserver and then opening its URL in a target source once it (the web server) is running.

## Usage

This package is meant to be used by other packages that have configured a [http.Mux](https://pkg.go.dev/net/http#ServeMux) instance for serving web requests. The package will start a web server on localhost listening on a randomly chose port number (unless a user-defined value is provided) and then open that URL in a target environment (like a web browser).

```
import (
	"context"
	"net/http"

	"github.com/sfomuseum/go-www-show"
)

func main() {

	ctx := context.Background()
	
	mux := http.NewServeMux()
	
	// Configure handlers for mux here

	browser, _ = show.NewBrowser(ctx, "web://")
	
        show_opts := &www_show.RunOptions{
                Browser: browser,
                Mux:     mux,
        }

        return show.RunWithOptions(ctx, www_show_opts)
}
```

This package defines a `Browser` interface for opening URLs. Currently there is only a single implementation (`web`) for opening URLs in the operating system's default web browser. In the future there may be other implementations to "open" a URL by delivering it to a remote service that handles opening and displaying that URL.

For a complete working example see ...