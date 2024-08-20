# go-geoparquet-show

Command-line tool for serving GeoParquet data from an on-demand web server.

## Important

This is an early-stage project. There may still be bugs (or just bad decisions).

It works reasonably well for small GeoParquet files. It is _very slow_ for large GeoParquet files. Maybe I am just "doing it wrong"?

## Tools

```
$> make cli
go build -mod vendor -ldflags="-s -w" -o bin/show cmd/show/main.go
```

### show

```
$> ./bin/show -h
Command-line tool for serving GeoParquet features as vector tiles from an on-demand web server.
Usage:
	 ./bin/show [options]
Valid options are:
  -data-source string
    	The URI of the GeoParquet data. Specifically, the value passed to the DuckDB read_parquet() function.
  -database-engine string
    	The database/sql engine (driver) to use. (default "duckdb")
  -port int
    	The port number to listen for requests on (on localhost). If 0 then a random port number will be chosen.
  -verbose
    	Enable vebose (debug) logging.
```

## See also

* https://github.com/sfomuseum/go-http-mvt
* https://github.com/sfomuseum/go-www-show
* https://github.com/marcboeker/go-duckdb