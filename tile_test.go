package show

import (
	"testing"

	"github.com/paulmach/orb/encoding/wkt"
)

func TestOrbMultiPoints(t *testing.T) {

	tests := []string{
		"MULTIPOINT (-122.388749 37.620113)",
	}

	for _, wkt_geom := range tests {

		orb_geom, err := wkt.Unmarshal(wkt_geom)

		if err != nil {
			t.Fatalf("Failed to unmarshal '%s', %v", wkt_geom, err)
		}

		t.Log(wkt_geom, orb_geom)
	}
}
