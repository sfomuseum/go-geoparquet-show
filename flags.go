package show

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/sfomuseum/go-flags/flagset"
	"github.com/sfomuseum/go-flags/multi"
	"github.com/sfomuseum/go-www-show/v2"
)

var data_source string
var db_engine string
var port int

var browser_uri string

var renderer string

var label_properties multi.MultiString

var max_x_column string
var max_y_column string

var verbose bool

func DefaultFlagSet() *flag.FlagSet {

	fs := flagset.NewFlagSet("show")

	browser_schemes := show.BrowserSchemes()
	str_schemes := strings.Join(browser_schemes, ",")

	browser_desc := fmt.Sprintf("A valid sfomuseum/go-www-show/v2.Browser URI. Valid options are: %s", str_schemes)

	fs.StringVar(&browser_uri, "browser-uri", "web://", browser_desc)

	fs.IntVar(&port, "port", 0, "The port number to listen for requests on (on localhost). If 0 then a random port number will be chosen.")
	fs.StringVar(&data_source, "data-source", "", "The URI of the GeoParquet data. Specifically, the value passed to the DuckDB read_parquet() function.")
	fs.StringVar(&db_engine, "database-engine", "duckdb", "The database/sql engine (driver) to use.")

	fs.StringVar(&renderer, "renderer", "leaflet", "Which rendering library to use to draw vector tiles. Valid options are: leaflet, maplibre.")
	fs.Var(&label_properties, "label", "Zero or more (GeoJSON Feature) properties to use to construct a label for a feature's popup menu when it is clicked on.")

	fs.StringVar(&max_x_column, "max-x-column", "", "An option column name to use for a initial bounding box constraint. This columns is expected to contain the maximum X (longitude) value of the geometry it is associated with. This will only work if the -max-y-column flag is also set.")
	fs.StringVar(&max_y_column, "max-y-column", "", "An option column name to use for a initial bounding box constraint. This columns is expected to contain the maximum Y (latitude) value of the geometry it is associated with. This will only work if the -max-x-column flag is also set.")

	fs.BoolVar(&verbose, "verbose", false, "Enable vebose (debug) logging.")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Command-line tool for serving GeoParquet features as vector tiles from an on-demand web server.\n")
		fmt.Fprintf(os.Stderr, "Usage:\n\t %s [options]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Valid options are:\n")
		fs.PrintDefaults()
	}

	return fs
}
