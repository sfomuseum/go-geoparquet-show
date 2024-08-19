package show

import (
	"flag"

	"github.com/sfomuseum/go-flags/flagset"
)

var data_source string
var db_engine string

func DefaultFlagSet() *flag.FlagSet {

	fs := flagset.NewFlagSet("show")

	fs.StringVar(&data_source, "data-source", "", "...")
	fs.StringVar(&db_engine, "database-engine", "duckdb", "...")

	return fs
}
