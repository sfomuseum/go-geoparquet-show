window.addEventListener("load", function load(event){

    var map;

    // To do: Read this from cfg
    var tiles_styles = {
	
	all: function(properties, zoom) {
	    return {
		weight: 2,
		color: 'red',
		opacity: .5,
		fillColor: 'yellow',
		fill: true,
		radius: 6,
		fillOpacity: 0.1
	    }
	}
    };
    
    var init = function(cfg){

	var bounds = [
	    [ cfg.minx, cfg.miny ],
	    [ cfg.maxx, cfg.maxy ],
	];

	console.log("BOUNDS", bounds);
	
	var tiles_url = "/tiles/all/{z}/{x}/{y}.mvt";

	//
	
	map = new maplibregl.Map({
            container: 'map',
	    style: 'https://demotiles.maplibre.org/style.json',
	    bounds: bounds,
	});

	map.on('load', () => {

	    console.log("WOO");
	    
            map.addSource('all', {
		type: 'vector',
		url: tiles_url,
            });
            map.addLayer({
		'id': 'terrain-data',
		'type': 'line',
		'source': 'all',
		'source-layer': 'all',
		'layout': {
                    'line-join': 'round',
                    'line-cap': 'round'
		},
		'paint': {
                    'line-color': '#ff69b4',
                    'line-width': 1
		}
            });
	});

	return;
	
	var bounds = [
	    [ cfg.miny, cfg.minx ],
	    [ cfg.maxy, cfg.maxx ],
	];
		
	map = L.map('map');
	map.fitBounds(bounds);

	// To do: Read this from cfg	
	var tiles_url = "/tiles/all/{z}/{x}/{y}.mvt";
	
	var tiles_opts = {
	    rendererFactory: L.canvas.tile,
	    vectorTileLayerStyles: tiles_styles,
	    interactive: true,
	};
	
	var layer = L.vectorGrid.protobuf(tiles_url, tiles_opts);

	// https://github.com/Leaflet/Leaflet.VectorGrid/issues/148
	
	layer.on('click', function(e) {
	    console.log("CLICK", e.layer);
	});
    
	layer.addTo(map);
	
    };
    
    fetch("/map.json")
	.then((rsp) => rsp.json())
	.then((cfg) => {
	    init(cfg);
	}).catch((err) => {
	    console.error("Failed to retrieve map config", err);
	});
        
    
});
