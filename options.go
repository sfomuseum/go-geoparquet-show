package show

import (
	"context"
	"database/sql"
	"flag"
	"fmt"

	_ "github.com/marcboeker/go-duckdb"

	"github.com/sfomuseum/go-flags/flagset"
	www_show "github.com/sfomuseum/go-www-show"
)

type RunOptions struct {
	Database   *sql.DB
	Datasource string
	Port       int
	Verbose    bool
	Browser    www_show.Browser
}

func RunOptionsFromFlagSet(ctx context.Context, fs *flag.FlagSet) (*RunOptions, error) {

	flagset.Parse(fs)

	db, err := sql.Open(db_engine, "")

	if err != nil {
		return nil, fmt.Errorf("Failed to open database, %w", err)
	}

	browser, err := www_show.NewBrowser(ctx, "web://")

	if err != nil {
		return nil, fmt.Errorf("Failed to create new browser, %w", err)
	}

	opts := &RunOptions{
		Database:   db,
		Datasource: data_source,
		Port:       port,
		Verbose:    verbose,
		Browser:    browser,
	}

	return opts, nil
}
