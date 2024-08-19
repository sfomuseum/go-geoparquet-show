package show

import (
	"context"
	"database/sql"
	"flag"
	"fmt"

	_ "github.com/marcboeker/go-duckdb"

	"github.com/sfomuseum/go-flags/flagset"
)

type RunOptions struct {
	Database   *sql.DB
	Datasource string
	Port       int
}

func RunOptionsFromFlagSet(ctx context.Context, fs *flag.FlagSet) (*RunOptions, error) {

	flagset.Parse(fs)

	db, err := sql.Open(db_engine, "")

	if err != nil {
		return nil, fmt.Errorf("Failed to open database, %w", err)
	}

	opts := &RunOptions{
		Database:   db,
		Datasource: data_source,
	}

	return opts, nil
}
