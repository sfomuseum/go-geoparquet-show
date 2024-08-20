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

    var zm = 12;
    
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
		fillOpacity: 0.7
	    }
	}
    };
    
    var tiles_url = "/tiles/all/{z}/{x}/{y}.mvt";

    var tiles_opts = {
	rendererFactory: L.canvas.tile,
	vectorTileLayerStyles: tiles_styles,
    };
    
    var layer = L.vectorGrid.protobuf(tiles_url, tiles_opts);
    layer.addTo(map);

        
    
});
