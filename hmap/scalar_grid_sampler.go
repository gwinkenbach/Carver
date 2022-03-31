package hmap

import "alvin.com/GoCarver/geom"

type ScalarGridSampler interface {
	GetNumSamplesFromX0ToX1(x0, x1 float64) int
	GetNumSamplesFromY0ToY1(y0, y1 float64) int
	EnableInvertImage(enable bool)
	At(p geom.Pt2) float64
}
