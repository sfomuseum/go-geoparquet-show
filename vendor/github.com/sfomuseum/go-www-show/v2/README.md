# go-www-show

Go package for starting a local	webserver and then opening its URL in a target environment once it (the web server) is running.

## Usage

This package is meant to be used by other packages that have configured a [http.Mux](https://pkg.go.dev/net/http#ServeMux) instance for serving web requests. The package will start a web server on localhost listening on a randomly chose port number (unless a user-defined value is provided) and then open that URL in a target environment (like a web browser).

```
import (
	"context"
	"net/http"

	"github.com/sfomuseum/go-www-show/v2"
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

_For a complete working example see the [sfomuseum/go-geojson-show](https://github.com/sfomuseum/go-geojson-show/blob/main/show.go) package._

## Browsers (targets)

This package defines a `Browser` interface for opening URLs in or more target environments.

```
// Browser is an interface for rendering URLs.
type Browser interface {
	// OpenURL opens a given in a URL specific to the browser's implementation context. The method expects
	// a URL to open and a channel that can be used (optionally) to signal when the specific `Browser`
	// implementation exits or otherwise triggers a "completed" state.
	OpenURL(context.Context, string, chan bool) error
}
```

The following implementations are available when using this package:

### WebBrowser (web://)

Open URLs in the operating system's default web browser. The `WebBrowser` implementation can be instantiated like this:

```
browser, _ := show.NewBrowser(ctx, "web://")
```

### WebViewBrowser (webview://)

Open URLs in a dedicated application window using the [webview/webview_go](https://github.com/webview/webview_go) package. The `WebViewBrowsr` implementation can be instantiated like this:

```
browser, _ = show.NewBrowser(ctx, "webview://?width=500&height=700")
```

#### Parameters

| Name | Value | Required | Notes |
| --- | --- | --- | --- |
| width | int | no | The initial width of the webview window. Default is 1024. |
| height | int | no | The initial height of the webview window. Default is 1024. |

Applications requiring the `WebViewBrowser` implementation will need to be built with the `webview` tag. For example:

```
$> go build -mod vendor -tags webview -ldflags="-s -w" -o bin/show-url cmd/show-url/main.go
```

## Tools

```
$> make cli
go build -mod vendor -ldflags="-s -w" -o bin/show-url cmd/show-url/main.go
```

_Note that the default `cli` Makefile target does not build tools with the `webview` tag enabled._

### show-url

Open a URL in a sfomuseum/go-www-show/v2.Browser environment.

```
$> ./bin/show-url -h
Open a URL in a sfomuseum/go-www-show/v2.Browser environment.
Usage:
	 ./bin/show-url [options]
Valid options are:
  -browser-uri string
    	A valid sfomuseum/go-www-show/v2.Browser URI. Valid options are: web:// (default "web://")
  -url string
    	The URL to open
```

## See also

* https://github.com/pkg/browser
* https://github.com/webview/webview_go
