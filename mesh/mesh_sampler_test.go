package mesh

import (
	"math"
	"testing"

	"alvin.com/GoCarver/geom"
	a "gotest.tools/assert"
)

func TestSampleTriangleWithBallpoint(t *testing.T) {
	trg := makeMeshTriangle([3]geom.Pt3{{X: 0, Y: 0, Z: 1}, {X: 0, Y: 1, Z: 1}, {X: 1, Y: 1, Z: 0}})

	// Drop tool in the middle of the triangle.
	fp := NewFootprint(geom.NewPt2(0, 0), geom.NewPt2(1, 1))
	ok, h := sampleTriangleWithBallpointTool(fp, &trg)
	a.Equal(t, ok, true)
	a.Assert(t, epsEq(h, 0.5*math.Sqrt(2), 1e-6))

	// Drop tool in middle of top edge.
	fp = NewFootprint(geom.NewPt2(-0.5, 0.0), geom.NewPt2(0.5, 1.0))
	ok, h = sampleTriangleWithBallpointTool(fp, &trg)
	a.Equal(t, ok, true)
	a.Equal(t, h, 1.0)

	// Barely grazing top edge
	fp = NewFootprint(geom.NewPt2(-0.9999999, 0.0), geom.NewPt2(0.0000001, 1.0))
	ok, h = sampleTriangleWithBallpointTool(fp, &trg)
	a.Equal(t, ok, true)
	a.Assert(t, epsEq(h, 0.5, 0.00032))
}
