package hmap

import (
	"image"
	"image/color"
	"math"

	"alvin.com/GoCarver/geom"
)

const (
	uint16Weight = 1.0 / math.MaxUint16
)

type pixelDepthSampler struct {
	img       *image.Gray16
	imgWidth  int
	imgHeight int

	carvingAreaOrigin geom.Pt2
	carvingAreaDim    geom.Size2

	matToPixelXform geom.Matrix33 // Material to pixel coordinate transform.
}

func NewPixelDepthSampler(
	mcToNicXform *geom.Matrix33,
	carvingAreaOrigin geom.Pt2,
	carvingAreaDim geom.Size2,
	img *image.Gray16) ScalarGridSampler {

	sampler := &pixelDepthSampler{
		carvingAreaOrigin: carvingAreaOrigin,
		carvingAreaDim:    carvingAreaDim,
	}

	sampler.imgWidth = img.Bounds().Dx()
	sampler.imgHeight = img.Bounds().Dy()
	sampler.img = img

	t := geom.NewTranslateMatrix33(0, float64(sampler.imgHeight-1)+0.5)
	s := geom.NewScaleMatrix33(float64(sampler.imgWidth-1)+0.5, -float64(sampler.imgHeight-1)-0.5)
	sampler.matToPixelXform = *mcToNicXform.Mul(s.Mul(&t))

	return sampler
}

func (p *pixelDepthSampler) GetNumSamplesFromP0ToP1(p0, p1 geom.Pt2) int {
	q0 := p0.Xform(&p.matToPixelXform)
	q1 := p1.Xform(&p.matToPixelXform)
	l := q1.Sub(q0)
	return int(l.Len())
}

func (p *pixelDepthSampler) At(q *geom.Pt2) float64 {
	q1 := q.Xform(&p.matToPixelXform)

	y := int(math.Max(0, q1.Y))
	if y > p.imgHeight-1 {
		y = p.imgHeight - 1
	}

	x := int(math.Max(0, q1.X))
	if x > p.imgWidth-1 {
		x = p.imgWidth - 1
	}

	pixVal := p.img.At(x, y)
	grayVal := pixVal.(color.Gray16)
	return float64(grayVal.Y) * uint16Weight
}
