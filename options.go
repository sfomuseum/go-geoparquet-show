package show

import (
	"context"
	"database/sql"
	"flag"
	"fmt"

	"github.com/sfomuseum/go-flags/flagset"
	www_show "github.com/sfomuseum/go-www-show/v2"
)

// RunOptions defines options for configuring and starting a local web server to serve GeoParquet data as vector tiles.
type RunOptions struct {
	// A valid `sql.DB` (DuckDB) instance to use for querying data
	Database *sql.DB
	// The URI of the GeoParquet data. Specifically, the value passed to the DuckDB read_parquet() function.
	Datasource string
	// The port number to listen for requests on (on localhost). If 0 then a random port number will be chosen.
	Port int
	// Enable verbose (debug) logging.
	Verbose bool
	// A `sfomuseum/go-www-show.Browser` instance to use for opening URLs.
	Browser www_show.Browser
	// An optional list of properties to use when creating popup labels
	LabelProperties []string
	// Which vector tile renderer to use. Valid options are: leaflet, maplibre.
	Renderer string `json:"renderer"`
	// An option column name to use for a initial bounding box constraint. This columns is expected to contain the maximum X (longitude) value of the geometry it is associated with.
	MaxXColumn string
	// An option column name to use for a initial bounding box constraint. This columns is expected to contain the maximum Y (latitude) value of the geometry it is associated with.
	MaxYColumn string
}

// Derive a new `RunOptions` instance from 'fs'.
func RunOptionsFromFlagSet(ctx context.Context, fs *flag.FlagSet) (*RunOptions, error) {

	flagset.Parse(fs)

	db, err := sql.Open(db_engine, "")

	if err != nil {
		return nil, fmt.Errorf("Failed to open database, %w", err)
	}

	browser, err := www_show.NewBrowser(ctx, browser_uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to create new browser, %w", err)
	}

	opts := &RunOptions{
		Database:        db,
		Datasource:      data_source,
		Port:            port,
		Verbose:         verbose,
		Browser:         browser,
		LabelProperties: label_properties,
		Renderer:        renderer,
		MaxXColumn:      max_x_column,
		MaxYColumn:      max_y_column,
	}

	return opts, nil
}
