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
	
	var tiles_url = "http://" + location.host + "/tiles/all/{z}/{x}/{y}.mvt";
	console.log("TILES", tiles_url);
	
	//
	
	map = new maplibregl.Map({
            container: 'map',
	    style: 'https://demotiles.maplibre.org/style.json',
	    bounds: bounds,
	});

	map.on('load', () => {

            map.addSource('all', {
		type: 'vector',
		tiles: [
		    tiles_url,
		],
            });
	    
            map.addLayer({
		'id': 'all-fill',
		'type': 'fill',
		'source': 'all',
		'source-layer': 'all',
		'paint': {
		    'fill-color': '#cc6699',
		    'fill-opacity': 0.1,
		}
            });

	    /*
            map.addLayer({
		'id': 'all-circle',
		'type': 'circle',
		'source': 'all',
		'source-layer': 'all',
		'paint': {
		    'circle-color': '#000',
		    'circle-radius': 4,
		    'circle-opacity': 0.5,
		}
            });
	     */
	    
            map.addLayer({
		'id': 'all-line',
		'type': 'line',
		'source': 'all',
		'source-layer': 'all',
		'layout': {
                    'line-join': 'round',
                    'line-cap': 'round'
		},
		'paint': {
                    'line-color': '#000',
                    'line-width': 1,
		}
            });
	    
	    var label_props = cfg.label_properties;

	    if (label_props){

		var count_props = label_props.length;
		
		if (count_props > 0) {

		    map.on('click', 'all-fill', (e) => {
			
			var label_text = [];
			
			for (var i=0; i < count_props; i++){
			    var prop = label_props[i];
			    var value = e.features[0].properties[ prop ];
			    label_text.push("<strong>" + prop + "</strong> " + value);
			}
			
			if (label_text.length > 0){			    
			    new maplibregl.Popup()
					  .setLngLat(e.lngLat)
					  .setHTML(label_text.join("<br />"))
					  .addTo(map);
			};
		    });
		}
	    }
	    
            map.on('mouseenter', 'all', () => {
		map.getCanvas().style.cursor = 'pointer';
            });

            map.on('mouseleave', 'all', () => {
		map.getCanvas().style.cursor = '';
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
