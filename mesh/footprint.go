package mesh

import (
	"math"

	"alvin.com/GoCarver/geom"
)

// Footprint is the XY-bounding rect for an area in the mesh.
type Footprint struct {
	PMin geom.Pt2
	PMax geom.Pt2
}

// NewFootprint creates and returns a new footprint object initialized from point p1 and p2.
func NewFootprint(p1, p2 geom.Pt2) Footprint {
	return Footprint{
		PMin: geom.NewPt2(math.Min(p1.X, p2.X), math.Min(p1.Y, p2.Y)),
		PMax: geom.NewPt2(math.Max(p1.X, p2.X), math.Max(p1.Y, p2.Y)),
	}
}
