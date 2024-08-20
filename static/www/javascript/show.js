window.addEventListener("load", function load(event){

    // Manhattan
    var lat = 40.777718;
    var lon = -73.96693;

    // SF
    lat = 37.778008;
    lon = -122.431272;

    // SFO
    lat = 37.621131;
    lon = -122.384292;

    var zm = 6;
    
    var map = L.map('map').setView([lat, lon], zm);

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
    
    var tiles_url = "/tiles/all/{z}/{x}/{y}.mvt";

    var tiles_opts = {
	rendererFactory: L.canvas.tile,
	vectorTileLayerStyles: tiles_styles,
	interactive: true,
    };
    
    var layer = L.vectorGrid.protobuf(tiles_url, tiles_opts);

    layer.on('click', function(e) {

	console.log("CLICK", e.layer);
	L.DomEvent.stop(e);
	
	/*
	L.popup()
	// .setContent(e.layer.properties.name || e.layer.properties.type)
	 .setContent(JSON.stringify(e.layer))
	 .setLatLng(e.latlng)
	 .openOn(map);
	
	clearHighlight();
	highlight = e.layer.properties.osm_id;
	
	pbfLayer.setFeatureStyle(highlight, {
	    weight: 2,
	    color: 'red',
	    opacity: 1,
	    fillColor: 'red',
	    fill: true,
	    radius: 6,
	    fillOpacity: 1
	})
	
	L.DomEvent.stop(e);
	*/
	
    });
    
    layer.addTo(map);

        
    
});
