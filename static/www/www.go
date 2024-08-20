package www

import (
	"embed"
)

//go:embed css/* javascript/* *.html
var FS embed.FS
