package hmap

import (
	"math"

	"alvin.com/GoCarver/geom"
)

// FiftyPercentTestSampler is a constant-depth sampler that always returns 0.5.
type FiftyPercentTestSampler struct {
}

func (f *FiftyPercentTestSampler) GetNumSamplesFromP0ToP1(p0, p1 geom.Pt2) int {
	l := p1.Sub(p0)
	return int(math.Round(2.56 * l.Len()))
}

func (f *FiftyPercentTestSampler) At(q *geom.Pt2) float64 {
	return 0.5
}

// ConstantDepthTestSampler is a constant-depth sampler that returns a preset depth.
type ConstantDepthTestSampler struct {
	depth float64
}

func NewConstantDepthSampler(depth float64) ConstantDepthTestSampler {
	return ConstantDepthTestSampler{depth: depth}
}

func (s *ConstantDepthTestSampler) GetNumSamplesFromP0ToP1(p0, p1 geom.Pt2) int {
	l := p1.Sub(p0)
	return int(math.Round(1.0 * l.Len())) // i.e. 1 pix per mm.
}

func (s *ConstantDepthTestSampler) At(q *geom.Pt2) float64 {
	return s.depth
}
