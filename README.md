# go-geoparquet-show

Command-line tool for serving GeoParquet data as vector tiles from an on-demand web server.

## Motivation

It's like [go-geojson-show](https://github.com/sfomuseum/go-geojson-show) (which is itself meant to be a simpler and dumber version of [geojson.io](https://geojson.io)) but for GeoParquet files. Specifically, a simple binary application for serving GeoParquet data as vector tiles from an on-demand web server.

## Important

* This is an early-stage project. There may still be bugs (or just bad decisions).

* It works reasonably well for small GeoParquet files. It is _very slow_ for large GeoParquet files. Under the hood it is using [DuckDB](https://www.duckdb.org/), and more specifically the [go-duckdb](https://github.com/marcboeker/go-duckdb) package, to query GeoParquet files. Maybe I am just "doing it wrong"? 

* There are no interactive features yet. The code is using the [Leaflet/Leaflet.VectorGrid](https://github.com/Leaflet/Leaflet.VectorGrid) package to render tiles but all the map `onclick` events trigger "L.DomEvent._fakeStop is not a function" which I haven't figured out yet. Any help or pointers would be appreciated.

* It is not possible to define custom styles yet. There is a single global style applied to all features.

* It is not possible to define different layers for features. Currently all features are assigned to a layer named "all".

* It is not possible to filter the features returned for any given layer. Currently all the feature contained by a (map) tile's extent are returned.

## Tools

```
$> make cli
go build -mod vendor -ldflags="-s -w" -o bin/show cmd/show/main.go
```

_If you encounter problems building the tools it might have something to do with the way `go-duckdb` is vendored. The best place to start debugging things is [this section in the go-duckdb documentation](https://github.com/marcboeker/go-duckdb?tab=readme-ov-file#vendoring)._

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

#### Examples

##### Serve a GeoParquet file derived from all the records in the [sfomuseum-data-architecture](https://github.com/sfomuseum-data/sfomuseum-data-architecture) repository:

![](docs/images/go-geoparquet-show-extent.png)

```
$> ./bin/show \
	-data-source /usr/local/data/arch.geoparquet \
	-verbose
	
2024/08/20 17:37:53 DEBUG Verbose logging enabled
2024/08/20 17:37:53 DEBUG Start server
2024/08/20 17:37:53 DEBUG HEAD request succeeded url=http://localhost:58296
2024/08/20 17:37:53 INFO Server is ready and features are viewable url=http://localhost:63744
2024/08/20 17:37:54 ERROR Failed to unmarshal geometry layer=all geom="GEOMETRYCOLLECTION (MULTIPOLYGON (((-122.388006 37.614539, -122.387968 37.614548, -122.387967 37.614547, -122.388005 37.614538, -122.388006 37.614539)), ((-122.387948 37.614553, -122.387913 37.614561, -122.387913 37.614561, -122.387948 37.614552, -122.387948 37.614553)), ((-122.387997 37.614517, -122.387961 37.614526, -122.387961 37.614527, -122.387997 37.614518, -122.387997 37.614517)), ((-122.387938 37.614532, -122.387904 37.61454, -122.387905 37.614541, -122.387938 37.614532, -122.387938 37.614532))), MULTIPOLYGON (((-122.387993 37.614604, -122.388082 37.614582, -122.388082 37.614581, -122.387993 37.614603, -122.387993 37.614604)), ((-122.388029 37.614445, -122.387951 37.614464, -122.387951 37.614465, -122.388029 37.614446, -122.388029 37.614445)), ((-122.387924 37.614534, -122.387904 37.614539, -122.387905 37.61454, -122.387925 37.614535, -122.387924 37.614534))))" error="wkt: unsupported geometry"
2024/08/20 17:37:54 DEBUG Time to get features layer=all count=609 time=390.414167ms
2024/08/20 17:37:55 DEBUG Time to get features layer=all count=1 time=947.971541ms
2024/08/20 17:37:55 DEBUG Time to get features layer=all count=196 time=1.141258459s
2024/08/20 17:37:55 DEBUG Time to get features layer=all count=298 time=1.142821084s
2024/08/20 17:37:55 DEBUG Time to get features layer=all count=9 time=1.233271s
2024/08/20 17:37:55 DEBUG Time to get features layer=all count=803 time=1.281584375s
... and so on
```

The map view is initialized to fit the extent of all the features in the GeoParquet database. Here's another screenshot zoomed in to a smaller section:

![](docs/images/go-geoparquet-show-extent.png)

## See also

* https://geoparquet.org/
* https://www.duckdb.org/
* https://github.com/marcboeker/go-duckdb
* https://github.com/Leaflet/Leaflet.VectorGrid
* https://github.com/sfomuseum/go-http-mvt
* https://github.com/sfomuseum/go-www-show