package show

import (
	"flag"

	"github.com/sfomuseum/go-flags/flagset"
)

var data_source string
var db_engine string
var port int

var verbose bool

func DefaultFlagSet() *flag.FlagSet {

	fs := flagset.NewFlagSet("show")

	fs.IntVar(&port, "port", 0, "...")
	fs.StringVar(&data_source, "data-source", "", "...")
	fs.StringVar(&db_engine, "database-engine", "duckdb", "...")
	fs.BoolVar(&verbose, "verbose", false, "Enable vebose (debug) logging.")

	return fs
}
