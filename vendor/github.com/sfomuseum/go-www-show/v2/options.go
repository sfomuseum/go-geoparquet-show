package show

import (
	"net/http"
)

// RunOptions defines options for serving and opening a local web server
type RunOptions struct {
	// The hostname for the web server to listen on. If empty then "localhost" will be assumed.
	Host string
	// The port number for the web server to listen on. If `0` then a random port number will be assigned.
	Port int
	// The `http.ServeMux` that the web server should use to route requests.
	Mux *http.ServeMux
	// The `Browser` instance to use to open the URL pointing to the web server.
	Browser Browser
}
