package show

// mapConfig defines common configuration details for maps.
type mapConfig struct {
	// MinX is the minimum longitude of the database's extent
	MinX float64 `json:"minx"`
	// MinY is the minimum latitude of the database's extent	
	MinY float64 `json:"miny"`
	// MaxX is the maximum longitude of the database's extent	
	MaxX float64 `json:"maxx"`
	// MaxY is the maximum latitude of the database's extent		
	MaxY float64 `json:"maxy"`
}
