package mesh

import (
	"fmt"
	"math"
	"testing"

	"alvin.com/GoCarver/geom"
	"alvin.com/GoCarver/hmap"

	a "gotest.tools/assert"
	is "gotest.tools/assert/cmp"
)

const (
	xMin       = 0
	xMax       = 100
	yMin       = 0
	yMax       = 100
	zBlack     = 0
	zWhite     = 10
	numSamples = 5
)

var xVal = [numSamples]float64{0, 0.25, 0.5, 0.75, 1.0}
var yVal = [numSamples]float64{0, 0.25, 0.5, 0.75, 1.0}

type fourByFourSampler struct {
	xWeight float64
	yWeight float64
}

var _ hmap.ScalarGridSampler = (*fourByFourSampler)(nil)

func new4x4Sampler(xWeight, yWeight float64) *fourByFourSampler {
	s := &fourByFourSampler{
		xWeight: xWeight,
		yWeight: yWeight,
	}
	return s
}

func (s *fourByFourSampler) At(p *geom.Pt2) float64 {
	p.X = math.Max(xMin, math.Min(xMax, p.X))
	p.Y = math.Max(yMin, math.Min(yMax, p.Y))
	i := int((numSamples - 1) * (p.X - xMin) / (xMax - xMin))
	j := int((numSamples - 1) * (p.Y - yMin) / (yMax - yMin))
	return ((1.0 - s.xWeight) + s.xWeight*xVal[i]) * ((1.0 - s.yWeight) + s.yWeight*yVal[j])
}

func (s *fourByFourSampler) GetNumSamplesFromP0ToP1(p0, p1 geom.Pt2) int {
	if p0.X == p1.X {
		p0.Y = math.Max(yMin, p0.Y)
		p1.Y = math.Min(yMax, p1.Y)
		return int(numSamples * (p1.Y - p0.Y) / (yMax - yMin))
	}

	if p0.Y == p1.Y {
		p0.X = math.Max(xMin, p0.X)
		p1.X = math.Min(xMax, p1.X)
		return int(numSamples * (p1.X - p0.X) / (xMax - xMin))
	}

	return 0
}

func TestFlatTriangleMesh(t *testing.T) {
	// 4x4 sampler with both weights = 0 produces a flat mesh at z = zWhite.
	s := new4x4Sampler(0, 0)
	p1 := geom.NewPt2(xMin, yMin)
	p2 := geom.NewPt2(xMax, yMax)
	m := NewTriangleMesh(p1, p2, zBlack, zWhite, s)
	a.Assert(t, m != nil)

	nX, nY := m.GetNumTriangles()
	a.Assert(t, is.Equal(nX, 8))
	a.Assert(t, is.Equal(nY, 4))

	// printMeshTriangles(m, t)

	t00 := m.GetTriangle(0, 0)
	a.Assert(t, t00.Vertex(0).EqXyz(0, 0, 10))
	visitAllTriangles(m, t, func(iX, iY int, trg Triangle, t *testing.T) {
		a.Assert(t, trg.UnitNormal().EqXyz(0, 0, 1))
		a.Assert(t, trg.Vertex(0).Z == zWhite)
		a.Assert(t, trg.Vertex(1).Z == zWhite)
		a.Assert(t, trg.Vertex(2).Z == zWhite)
	})

	fp := m.GetFootprintForTriangle(0, 0)
	a.Assert(t, fp.PMax.Eq(25, 25))
	a.Assert(t, fp.PMin.Eq(0, 0))

	fp = m.GetFootprintForTriangle(1, 0)
	a.Assert(t, fp.PMax.Eq(25, 25))
	a.Assert(t, fp.PMin.Eq(0, 0))

	fp = m.GetFootprintForTriangle(6, 3)
	a.Assert(t, fp.PMax.Eq(100, 100))
	a.Assert(t, fp.PMin.Eq(75, 75))

	fp = m.GetFootprintForTriangle(7, 3)
	a.Assert(t, fp.PMax.Eq(100, 100))
	a.Assert(t, fp.PMin.Eq(75, 75))

	fp = NewFootprint(geom.NewPt2(1, 1), geom.NewPt2(24, 24))
	trg := m.GetTrianglesUnderFootprint(fp)
	a.Assert(t, trg.GetTriangleCount() == 2)
	t1 := trg.Next()
	a.Assert(t, t1.Vertex(0).EqXyz(0, 0, 10))
	a.Assert(t, t1.Vertex(1).EqXyz(0, 25, 10))
	a.Assert(t, t1.Vertex(2).EqXyz(25, 25, 10))
	t2 := trg.Next()
	a.Assert(t, t2.Vertex(0).EqXyz(0, 0, 10))
	a.Assert(t, t2.Vertex(1).EqXyz(25, 25, 10))
	a.Assert(t, t2.Vertex(2).EqXyz(25, 0, 10))

	fp = NewFootprint(geom.NewPt2(1, 1), geom.NewPt2(25, 24))
	trg = m.GetTrianglesUnderFootprint(fp)
	a.Assert(t, trg.GetTriangleCount() == 4)

	fp = NewFootprint(geom.NewPt2(1, 1), geom.NewPt2(25, 25))
	trg = m.GetTrianglesUnderFootprint(fp)
	a.Assert(t, trg.GetTriangleCount() == 8)

	fp = NewFootprint(geom.NewPt2(51, 51), geom.NewPt2(52, 52))
	trg = m.GetTrianglesUnderFootprint(fp)
	a.Assert(t, trg.GetTriangleCount() == 2)
	t1 = trg.Next()
	a.Assert(t, t1.Vertex(0).EqXyz(50, 50, 10))
	a.Assert(t, t1.Vertex(1).EqXyz(50, 75, 10))
	a.Assert(t, t1.Vertex(2).EqXyz(75, 75, 10))

	fp = NewFootprint(geom.NewPt2(0, 0), geom.NewPt2(100, 100))
	trg = m.GetTrianglesUnderFootprint(fp)
	a.Assert(t, trg.GetTriangleCount() == 32)
}

func visitAllTriangles(m *TriangleMesh, t *testing.T, visitor func(iX, iY int, trg Triangle, t *testing.T)) {
	nX, nY := m.GetNumTriangles()
	for y := 0; y < nY; y++ {
		for x := 0; x < nX; x++ {
			trg := m.GetTriangle(x, y)
			visitor(x, y, trg, t)
		}
	}
}

func printMeshTriangles(m *TriangleMesh, t *testing.T) {
	visitor := func(iX, iY int, trg Triangle, t *testing.T) {
		fmt.Printf("T%d%d: %v\n", iY, iX, trg)
	}
	visitAllTriangles(m, t, visitor)
}
