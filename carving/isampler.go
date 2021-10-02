package carving

import "alvin.com/GoCarver/geom"

type CarvingDepthSampler interface {
	getNumSamplesFromP0ToP1(p0, p1 geom.Pt2) int
	at(p *geom.Pt2) float64
}
