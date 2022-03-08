package mesh

import (
	"fmt"
	"math"
	"testing"

	"alvin.com/GoCarver/geom"
	a "gotest.tools/assert"
	"gotest.tools/assert/cmp"
)

// Make a mesh triangle out of three points.
func makeMeshTriangle(p [3]geom.Pt3) meshTriangle {
	v := p[2].Sub(p[1])
	w := p[0].Sub(p[1])
	n := v.Cross(w).Norm()
	if n.Z < 0 {
		n = n.Scale(-1)
	}
	//fmt.Printf("Make plane: p=%v n=%v\n", p, n)
	return meshTriangle{n, p}
}

// Return the point t * p2 + (1 - t) * p1.
func affPt(t float64, p1, p2 geom.Pt3) geom.Pt3 {
	oneMinusT := 1.0 - t
	return geom.NewPt3(
		oneMinusT*p1.X+t*p2.X,
		oneMinusT*p1.Y+t*p2.Y,
		oneMinusT*p1.Z+t*p2.Z)
}

// Compare Pt3 values for test assert.
func cmpPt3(value geom.Pt3, expect geom.Pt3, epsilon float64) cmp.Comparison {
	return func() cmp.Result {
		if epsilon <= 0 {
			if value.Eq(expect) {
				return cmp.ResultSuccess
			}
		} else {
			v := expect.Sub(value)
			if v.Dot(v) <= epsilon*epsilon {
				return cmp.ResultSuccess
			}
		}
		return cmp.ResultFailure(
			fmt.Sprintf("%+v is not equal to %+v with epsilon=%f", value, expect, epsilon))
	}
}

// Strick equality of Pt3 for test assert.
func eqPt3(value geom.Pt3, expect geom.Pt3) cmp.Comparison {
	return cmpPt3(value, expect, 0)
}

// Compare floats to within epsilon.
func epsEq(value float64, expect float64, epsilon float64) cmp.Comparison {
	return func() cmp.Result {
		if epsilon <= 0 {
			if value == expect {
				return cmp.ResultSuccess
			}
		} else {
			v := expect - value
			if math.Abs(v) <= epsilon {
				return cmp.ResultSuccess
			}
		}
		return cmp.ResultFailure(
			fmt.Sprintf("%+v is not equal to %+v with epsilon=%f", value, expect, epsilon))
	}
}

func TestDropPointTowardPlane(t *testing.T) {
	p := geom.NewPt3(0, 0, 10)
	r := 1.0

	// Horizontal plane at z = 0.
	trg := makeMeshTriangle([3]geom.Pt3{{X: 0, Y: 0, Z: 0}, {X: 1, Y: 0, Z: 0}, {X: 0, Y: 1, Z: 0}})
	check, z := dropPointPToDistanceRFromTrianglePlane(p, r, &trg)
	a.Assert(t, check == true)
	a.Assert(t, z == 1.0)

	// Vertical plane: should fail.
	trg = makeMeshTriangle([3]geom.Pt3{{X: 0, Y: 0, Z: 0}, {X: 1, Y: 0, Z: 0}, {X: 0, Y: 0, Z: 1}})
	check, _ = dropPointPToDistanceRFromTrianglePlane(p, r, &trg)
	a.Assert(t, check == false)

	// 45-degree plane on YZ.
	trg = makeMeshTriangle([3]geom.Pt3{{X: 0, Y: 0, Z: 0}, {X: 0, Y: 1, Z: 1}, {X: 1, Y: 0, Z: 0}})
	check, z = dropPointPToDistanceRFromTrianglePlane(p, r, &trg)
	a.Assert(t, check == true)
	a.Assert(t, epsEq(z, math.Sqrt(2), 1e-6))

	// 45-degree plane on XZ.
	trg = makeMeshTriangle([3]geom.Pt3{{X: 0, Y: 0, Z: 0}, {X: 1, Y: 0, Z: 1}, {X: 0, Y: 1, Z: 0}})
	check, z = dropPointPToDistanceRFromTrianglePlane(p, r, &trg)
	a.Assert(t, check == true)
	a.Equal(t, z, math.Sqrt(2))
}

func TestProjectPointToPlane(t *testing.T) {
	p := geom.NewPt3(1, 2, 10)

	// Horizontal plane at z = 0.
	trg := makeMeshTriangle([3]geom.Pt3{{X: 0, Y: 0, Z: 0}, {X: 1, Y: 0, Z: 0}, {X: 0, Y: 1, Z: 0}})
	q := projectPointToTrianglePlane(p, &trg)
	a.Assert(t, q.EqXyz(1, 2, 0))

	// Plane embeds point p.
	trg = makeMeshTriangle([3]geom.Pt3{{X: 1, Y: 2, Z: 10}, {X: 1, Y: 1, Z: 0}, {X: 0, Y: 0, Z: 1}})
	q = projectPointToTrianglePlane(p, &trg)
	a.Assert(t, q.EqXyz(1, 2, 10))

	// 45-degree plane on YZ.
	p = geom.NewPt3(0, 2, 0)
	trg = makeMeshTriangle([3]geom.Pt3{{X: 0, Y: 0, Z: 0}, {X: 0, Y: 1, Z: 1}, {X: 1, Y: 0, Z: 0}})
	q = projectPointToTrianglePlane(p, &trg)
	a.Assert(t, cmpPt3(q, geom.NewPt3(0, 1, 1), 1e-6))
}

func TestIsPointWithinTriangle(t *testing.T) {
	// Triangle in z=0 plane.
	trg := makeMeshTriangle([3]geom.Pt3{{X: 0, Y: 0, Z: 0}, {X: 0, Y: 1, Z: 0}, {X: 1, Y: 1, Z: 0}})
	a.Assert(t, isPlanePointWithinTriangle(geom.NewPt3(0, 0, 0), &trg))
	a.Assert(t, isPlanePointWithinTriangle(geom.NewPt3(0, 1, 0), &trg))
	a.Assert(t, isPlanePointWithinTriangle(geom.NewPt3(1, 1, 0), &trg))
	a.Assert(t, !isPlanePointWithinTriangle(geom.NewPt3(-0.01, 0, 0), &trg))
	a.Assert(t, !isPlanePointWithinTriangle(geom.NewPt3(1, 0, 0), &trg))
	a.Assert(t, !isPlanePointWithinTriangle(geom.NewPt3(1.001, 1.001, 0), &trg))

	// 45-degree plane on YZ.
	trg = makeMeshTriangle([3]geom.Pt3{{X: 0, Y: 0, Z: 0}, {X: 0, Y: 1, Z: 1}, {X: 1, Y: 1, Z: 0}})
	a.Assert(t, isPlanePointWithinTriangle(geom.NewPt3(0, 0, 0), &trg))
	a.Assert(t, isPlanePointWithinTriangle(geom.NewPt3(0, 1, 1), &trg))
	a.Assert(t, isPlanePointWithinTriangle(geom.NewPt3(1, 1, 0), &trg))
	a.Assert(t, isPlanePointWithinTriangle(geom.NewPt3(0, 0.999, 0.999), &trg))
	a.Assert(t, !isPlanePointWithinTriangle(geom.NewPt3(0, 1.0001, 1), &trg))
	a.Assert(t, isPlanePointWithinTriangle(geom.NewPt3(0, 1, 1.0001), &trg))
}

func TestProjectPointToLine(t *testing.T) {
	p := geom.NewPt3(5, 5, 5)
	q1 := geom.NewPt3(0, 0, 0)
	q2 := geom.NewPt3(1, 0, 0)
	q := projectPointToLine(p, q1, q2)
	a.Assert(t, eqPt3(q, geom.NewPt3(5, 0, 0)))

	q2 = geom.NewPt3(0, 1, 0)
	q = projectPointToLine(p, q1, q2)
	a.Assert(t, eqPt3(q, geom.NewPt3(0, 5, 0)))

	q2 = geom.NewPt3(0, 0, 1)
	q = projectPointToLine(p, q1, q2)
	a.Assert(t, eqPt3(q, geom.NewPt3(0, 0, 5)))

	q2 = geom.NewPt3(1, 1, 1)
	q = projectPointToLine(p, q1, q2)
	a.Assert(t, eqPt3(q, geom.NewPt3(5, 5, 5)))

	p = geom.NewPt3(0, 5, 0)
	q2 = geom.NewPt3(1, 1, 0)
	q = projectPointToLine(p, q1, q2)
	a.Assert(t, eqPt3(q, geom.NewPt3(2.5, 2.5, 0)))
}

func TestIsLinePointOnSegment(t *testing.T) {
	q1 := geom.NewPt3(1, 2, 3)
	q2 := geom.NewPt3(56, 12, -33)
	a.Assert(t, isLinePointOnSegment(q1, q1, q2))
	a.Assert(t, isLinePointOnSegment(q2, q1, q2))
	a.Assert(t, isLinePointOnSegment(affPt(0.001, q1, q2), q1, q2))
	a.Assert(t, isLinePointOnSegment(affPt(0.999, q1, q2), q1, q2))
	a.Assert(t, !isLinePointOnSegment(affPt(-0.001, q1, q2), q1, q2))
	a.Assert(t, !isLinePointOnSegment(affPt(1.001, q1, q2), q1, q2))
}

func TestDropPointTowardLine(t *testing.T) {
	// Drop line is x-axis.
	q := geom.NewPt3(0, 0, 0)
	w := geom.NewVec3(1, 0, 0)

	// Drop point is right above line.
	p := geom.NewPt3(1, 0, 10)
	ok, z := dropPointPToDistanceRFromLine(p, 1, q, w)
	a.Assert(t, ok)
	a.Equal(t, z, 1.0)

	// Drop point is offset from line by r.
	p = geom.NewPt3(0, 1, 10)
	ok, z = dropPointPToDistanceRFromLine(p, 1, q, w)
	a.Assert(t, ok)
	a.Equal(t, z, 0.0)

	// Drop point is offset from line by 0.5 * r.
	p = geom.NewPt3(0, 0.5, 10)
	ok, z = dropPointPToDistanceRFromLine(p, 1, q, w)
	a.Assert(t, ok)
	a.Assert(t, epsEq(z, 0.8660254, 1e-6))

	// Drop point is too far.
	p = geom.NewPt3(0, 2.0, 10)
	ok, _ = dropPointPToDistanceRFromLine(p, 1, q, w)
	a.Assert(t, ok == false)

	// Drop line is y-axis.
	w = geom.NewVec3(0, 1, 0)

	// Drop point is right above line.
	p = geom.NewPt3(0, 1, 10)
	ok, z = dropPointPToDistanceRFromLine(p, 1, q, w)
	a.Assert(t, ok)
	a.Equal(t, z, 1.0)

	// Drop point is offset from line by r.
	p = geom.NewPt3(1, 0, 10)
	ok, z = dropPointPToDistanceRFromLine(p, 1, q, w)
	a.Assert(t, ok)
	a.Equal(t, z, 0.0)

	// Drop point is offset from line by 0.5 * r.
	p = geom.NewPt3(0.5, 0, 10)
	ok, z = dropPointPToDistanceRFromLine(p, 1, q, w)
	a.Assert(t, ok)
	a.Assert(t, epsEq(z, 0.8660254, 1e-6))

	// Drop point is too far.
	p = geom.NewPt3(2.0, 0.0, 10)
	ok, _ = dropPointPToDistanceRFromLine(p, 1, q, w)
	a.Assert(t, ok == false)

	// Drop line is xy-diagonal.
	w = geom.NewVec3(1, 1, 0)

	// Drop point is above origin.
	p = geom.NewPt3(0, 0, 10)
	ok, z = dropPointPToDistanceRFromLine(p, 1, q, w)
	a.Assert(t, ok)
	a.Equal(t, z, 1.0)

	// Drop line is YZ-diagonal.
	w = geom.NewVec3(0, 1, 1)

	// Drop point is above origin.
	p = geom.NewPt3(0, 0, 10)
	ok, z = dropPointPToDistanceRFromLine(p, 1, q, w)
	a.Assert(t, ok)
	a.Assert(t, epsEq(z, math.Sqrt(2.0), 1e-6))

	// Drop line is XZ-diagonal.
	w = geom.NewVec3(1, 0, 1)

	// Drop point is above origin.
	p = geom.NewPt3(0, 0, 10)
	ok, z = dropPointPToDistanceRFromLine(p, 1, q, w)
	a.Assert(t, ok)
	a.Assert(t, epsEq(z, math.Sqrt(2.0), 1e-6))

	// Same line but with shifted origin and direction pointing backward.
	q = geom.NewPt3(55, 0, 55)
	w = geom.NewVec3(-6, 0, -6)

	// Drop point is above origin.
	p = geom.NewPt3(0, 0, 10)
	ok, z = dropPointPToDistanceRFromLine(p, 1, q, w)
	a.Assert(t, ok)
	a.Assert(t, epsEq(z, math.Sqrt(2.0), 1e-6))
}
