package show

import (
	"flag"
	"fmt"
	"os"

	"github.com/sfomuseum/go-flags/flagset"
)

var data_source string
var db_engine string
var port int

var verbose bool

func DefaultFlagSet() *flag.FlagSet {

	fs := flagset.NewFlagSet("show")

	fs.IntVar(&port, "port", 0, "The port number to listen for requests on (on localhost). If 0 then a random port number will be chosen.")
	fs.StringVar(&data_source, "data-source", "", "The URI of the GeoParquet data. Specifically, the value passed to the DuckDB read_parquet() function.")
	fs.StringVar(&db_engine, "database-engine", "duckdb", "The database/sql engine (driver) to use.")
	fs.BoolVar(&verbose, "verbose", false, "Enable vebose (debug) logging.")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Command-line tool for serving GeoParquet features as vector tiles from an on-demand web server.\n")
		fmt.Fprintf(os.Stderr, "Usage:\n\t %s [options]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Valid options are:\n")
		fs.PrintDefaults()
	}

	return fs
}
