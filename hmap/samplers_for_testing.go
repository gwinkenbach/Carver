package hmap

import (
	"math"

	"alvin.com/GoCarver/geom"
)

// FiftyPercentTestSampler is a constant-depth sampler that always returns 0.5.
type FiftyPercentTestSampler struct {
}

func (f *FiftyPercentTestSampler) GetNumSamplesFromX0ToX1(x0, x1 float64) int {
	l := x1 - x0
	return int(math.Round(2.56 * l))
}

func (f *FiftyPercentTestSampler) GetNumSamplesFromY0ToY1(y0, y1 float64) int {
	l := y1 - y0
	return int(math.Round(2.56 * l))
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

func (s *ConstantDepthTestSampler) GetNumSamplesFromX0ToX1(x0, x1 float64) int {
	l := x1 - x0
	return int(math.Round(1.0 * l)) // i.e. 1 pix per mm.
}

func (s *ConstantDepthTestSampler) GetNumSamplesFromY0ToY1(y0, y1 float64) int {
	l := y1 - y0
	return int(math.Round(1.0 * l)) // i.e. 1 pix per mm.
}

func (s *ConstantDepthTestSampler) At(q *geom.Pt2) float64 {
	return s.depth
}
