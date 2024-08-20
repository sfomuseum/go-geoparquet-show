package show

import (
	"net/http"
)

type RunOptions struct {
	Port    int
	Mux     *http.ServeMux
	Browser Browser
}
