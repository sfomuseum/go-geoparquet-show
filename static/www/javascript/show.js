window.addEventListener("load", function load(event){

    var init_leaflet = function(cfg){

	var bounds = [
	    [ cfg.miny, cfg.minx ],
	    [ cfg.maxy, cfg.maxx ],
	];

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
	
	var map = L.map('map');
	map.fitBounds(bounds);

	// To do: Read this from cfg	
	var tiles_url = "/tiles/all/{z}/{x}/{y}.mvt";
	
	var tiles_opts = {
	    rendererFactory: L.canvas.tile,
	    vectorTileLayerStyles: tiles_styles,
	    interactive: true,
	};
	
	var layer = L.vectorGrid.protobuf(tiles_url, tiles_opts);

	// onclick events trigger mysterious "L.DomEvent._fakeStop is not a function" errors
	// https://github.com/Leaflet/Leaflet.VectorGrid/issues/148
	// layer.on('click', function(e) { ... })
	
	layer.addTo(map);
	
    };

    var init_maplibre = function(cfg){

	var bounds = [
	    [ cfg.minx, cfg.miny ],
	    [ cfg.maxx, cfg.maxy ],
	];

	var tiles_url = "http://" + location.host + "/tiles/all/{z}/{x}/{y}.mvt";
	
	var map = new maplibregl.Map({
            container: 'map',
	    bounds: bounds,
	    // style: 'https://demotiles.maplibre.org/style.json',	    
	    style: {
		"id": "go-geoparquet-show",
		"name": "go-geoparquet-show",
		"layers": [
		    {
			"id": "background",
			"type": "background",
			"paint": {
			    "background-color": "#D8F2FF"
			},
			"filter": [
			    "all"
			],
			"layout": {
			    "visibility": "visible"
			},
			"maxzoom": 24
		    },
		],
		"sources": {},
		"version": 8
	    }
	});
	
	map.on('load', () => {

	    try {
		
		map.addSource('all', {
		    type: 'vector',
		    tiles: [
			tiles_url,
		    ],
		});

		var l = map.addLayer({
		    'id': 'all-points',
		    'type': 'circle',
		    'source': 'all',
		    'source-layer': 'all',
		    'paint': {
			'circle-color': 'red',
			// 'circle-radius': 6,
			// Not really sure I understand what's happening here
			'circle-radius': [
			    "interpolate", ["linear"], ["zoom"],
			    0, 0,
			    20, ['*', 2, ['get', 'amount']]],
			'circle-opacity': 0.5,
			'circle-stroke-color': '#fff',
			'circle-stroke-width': 1,
		    }
		});

		// START OF this is important
		// Without this filter then the all-points layer renders layers for all the
		// points AND all the centroids of all the other features because... computers?
		// https://maplibre.org/maplibre-style-spec/expressions/#geometry-type
		
		map.setFilter('all-points', ["any",
					     ["==", ["geometry-type"], "Point"],
					     ["==", ["geometry-type"], "MultiPoint"]
		]);
		
		// END OF this is important	    

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
		}, 'all-points');
		
		
		map.addLayer({
		    'id': 'all-fill',
		    'type': 'fill',
		    'source': 'all',
		    'source-layer': 'all',
		    'paint': {
			// To do: change colour onclick
			// https://gis.stackexchange.com/questions/349827/change-polygon-color-on-click-with-mapbox
			'fill-color': '#cc6699',
			'fill-opacity': 0.1,
		    }
		}, 'all-line');
		
			
		var popup_layers = [
		    'all-fill',
		    'all-points',
		];
		
		var label_props = cfg.label_properties;
		
		if (label_props){
		    
		    var count_props = label_props.length;
		    
		    if (count_props > 0) {
			
			var show_popup = function(e){
			    
			    var label_text = [];
			    
			    for (var i=0; i < count_props; i++){
				var prop = label_props[i];
				var value = e.features[0].properties[ prop ];
				label_text.push("<strong>" + prop + "</strong> " + value);
			    }
			    
			    if (label_text.length == 0){
				return;
			    }
			    
			    new maplibregl.Popup()
					  .setLngLat(e.lngLat)
					  .setHTML(label_text.join("<br />"))
					  .addTo(map);
			};
			
			for (i in popup_layers){
			    
			    var layer_id = popup_layers[i];
			    
			    map.on('click', layer_id, (e) => {
				show_popup(e);
			    });
			}
		    }
		}
		
		for (i in popup_layers){
		    
		    var layer_id = popup_layers[i];
		    
		    map.on('mouseenter', layer_id, () => {
			map.getCanvas().style.cursor = 'pointer';
		    });
		    
		    map.on('mouseleave', layer_id, () => {
			map.getCanvas().style.cursor = '';
		    });
		}
		
	    } catch(err) {
		console.error("Failed to complete initialization on map load", err);
	    }
	});

	return;
	
    };
    
    var init = function(cfg){

	try {
	    
	    switch (cfg.renderer){
		case "maplibre":
		    init_maplibre(cfg);
		    break;
		default:
		    if (cfg.renderer != "leaflet"){
			log.warn("Unknown renderer, defaulting to leaflet", cfg.renderer)
		    }
		    init_leaflet(cfg);
		    break;
	    }
	    
	} catch(err) {
	    console.error("Failed to initialize map", err);
	}
    };
    
    fetch("/map.json")
	.then((rsp) => rsp.json())
	.then((cfg) => {
	    init(cfg);
	}).catch((err) => {
	    console.error("Failed to retrieve map config", err);
	});
        
    
});
