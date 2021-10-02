package carving

import (
	"math"

	"alvin.com/GoCarver/geom"
)

// FiftyPercentTestSampler is a constant-depth sampler that always returns 0.5.
type FiftyPercentTestSampler struct {
}

func (f *FiftyPercentTestSampler) getNumSamplesFromP0ToP1(p0, p1 geom.Pt2) int {
	l := p1.Sub(p0)
	return int(math.Round(2.56 * l.Len()))
}

func (f *FiftyPercentTestSampler) at(q *geom.Pt2) float64 {
	return 0.5
}

// ConstantDepthTestSampler is a constant-depth sampler that returns a preset depth.
type ConstantDepthTestSampler struct {
	depth float64
}

func newConstantDepthSampler(depth float64) ConstantDepthTestSampler {
	return ConstantDepthTestSampler{depth: depth}
}

func (s *ConstantDepthTestSampler) getNumSamplesFromP0ToP1(p0, p1 geom.Pt2) int {
	l := p1.Sub(p0)
	return int(math.Round(1.0 * l.Len())) // i.e. 1 pix per mm.
}

func (s *ConstantDepthTestSampler) at(q *geom.Pt2) float64 {
	return s.depth
}
