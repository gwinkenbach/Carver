package hmap

import "alvin.com/GoCarver/geom"

type ScalarGridSampler interface {
	GetNumSamplesFromP0ToP1(p0, p1 geom.Pt2) int
	At(p *geom.Pt2) float64
}
