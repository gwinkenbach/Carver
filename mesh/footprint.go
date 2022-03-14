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

// Return the footprint's center point.
func (f *Footprint) GetCenterPoint() geom.Pt2 {
	return geom.NewPt2((f.PMin.X+f.PMax.X)*0.5, (f.PMin.Y+f.PMax.Y)*0.5)
}

// Return the footprint's x-range as (xMin, xMax).
func (f *Footprint) GetXRange() (float64, float64) {
	return f.PMin.X, f.PMax.X
}

// Return the footprint's y-range as (yMin, yMax).
func (f *Footprint) GetYRange() (float64, float64) {
	return f.PMin.Y, f.PMax.Y
}

// Return the foorptint width xMax - xMin
func (f *Footprint) GetWidth() float64 {
	return f.PMax.X - f.PMin.X
}

// Return the foorptint height yMax - yMin
func (f *Footprint) GetHeight() float64 {
	return f.PMax.Y - f.PMin.Y
}
