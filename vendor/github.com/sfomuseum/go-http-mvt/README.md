# go-http-mvt

Go package proving http.HandlerFunc instances for serving MVT vector tiles using custom data providers.

## Example

```
import (
	"net/http"
	
	"github.com/sfomuseum/go-http-mvt"
)

func main() {

	// This is your custom code to implement the following method signature
	//  to return a dictionary of GeoJSON FeatureCollection instances:
	// func(context.Context, string, *maptile.Tile) (map[string]*geojson.FeatureCollection, error)
	
	var features_cb := mvt.GetFeaturesCallbackFunc

	mvt_opts := &mvt.TileHandlerOptions{
		GetFeaturesCallback: features_cb,
		Simplify:            true,
	}

	mvt_handler, _ := mvt.NewTileHandler(mvt_opts)

	mux := http.NewServeMux()
	mux.Handle("/tiles/", mvt_handler)

	http.ListenAndServe(":8080", mux)
}
```

This package does not provide any tile caching or CORS support. That is left to separate middleware handlers like [victorspringer/http-cache](https://github.com/victorspringer/http-cache) or [rs/cors](https://github.com/rs/cors) (or equivalents) respectively.