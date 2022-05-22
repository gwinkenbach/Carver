package carving

import (
	"testing"

	"alvin.com/GoCarver/geom"
	a "gotest.tools/assert"
	is "gotest.tools/assert/cmp"
)

func buildPath(g *grblGenerator, verts []geom.Pt3) {
	for _, v := range verts {
		g.appendPointToPath(v)
	}
}

func TestAddPoint(t *testing.T) {
	g := newGrblGenerator(100, 100)

	g.appendPointToPath(geom.NewPt3(0, 0, 0))
	if g.getNumPathPointsForTest() != 1 {
		t.Errorf("addPoint: expected len == 1\n")
	}

	g.appendPointToPath(geom.NewPt3(1, 1, 1))
	if g.getNumPathPointsForTest() != 2 {
		t.Errorf("addPoint: expected len == 2\n")
	}
}

func TestDistQtoP0P1Sqrd(t *testing.T) {
	p0 := geom.NewPt3(0, 0, 0)
	p1 := geom.NewPt3(10, 10, 10)
	q := geom.NewPt3(5, 5, 5)
	d := distQtoP0P1Sqrd(q, p0, p1)
	a.Assert(t, d < 1e-5)
}

func TestSimplifyPath(t *testing.T) {
	g := newGrblGenerator(100, 100)

	// Three non-colinear vertices.
	verts := []geom.Pt3{{0, 0, 0}, {1, 1, 1}, {0, 1, 1}}
	buildPath(g, verts)
	g.simplifyCompoundPath()
	a.Assert(t, is.Equal(g.getNumPathPointsForTest(), 3))
	for i, q := range g.getAllPathPointsForTest() {
		a.DeepEqual(t, q, verts[i])
	}

	// 2 vertices.
	g.reset()
	verts = []geom.Pt3{{0, 0, 0}, {1, 1, 1}}
	buildPath(g, verts)
	g.simplifyCompoundPath()
	a.Assert(t, is.Equal(g.getNumPathPointsForTest(), 2))

	pts := g.getAllPathPointsForTest()
	a.DeepEqual(t, pts[0], verts[0])
	a.DeepEqual(t, pts[1], verts[1])

	// 3 colinear vertices.
	g.reset()
	verts = []geom.Pt3{{0, 0, 0}, {1, 1, 1}, {2, 2, 2}}
	buildPath(g, verts)
	g.simplifyCompoundPath()
	a.Assert(t, is.Equal(g.getNumPathPointsForTest(), 2))

	pts = g.getAllPathPointsForTest()
	a.DeepEqual(t, pts[0], verts[0])
	a.DeepEqual(t, pts[1], verts[2])

	// 4 colinear vertices.
	g.reset()
	verts = []geom.Pt3{{0, 0, 0}, {1, 1, 1}, {2, 2, 2}, {3, 3, 3}}
	buildPath(g, verts)
	g.simplifyCompoundPath()
	a.Assert(t, is.Equal(g.getNumPathPointsForTest(), 2))

	pts = g.getAllPathPointsForTest()
	a.DeepEqual(t, pts[0], verts[0])
	a.DeepEqual(t, pts[1], verts[3])

	// 4 colinear vertices and one outlier
	g.reset()
	verts = []geom.Pt3{{0, 0, 0}, {1, 1, 1}, {2, 2, 2}, {3, 3, 3}, {4, 4, 0}}
	buildPath(g, verts)
	g.simplifyCompoundPath()
	a.Assert(t, is.Equal(g.getNumPathPointsForTest(), 3))

	pts = g.getAllPathPointsForTest()
	a.DeepEqual(t, pts[0], verts[0])
	a.DeepEqual(t, pts[1], verts[3])
	a.DeepEqual(t, pts[2], verts[4])

	// 4 colinear + 3 colinear vertices.
	g.reset()
	verts = []geom.Pt3{{0, 0, 0}, {1, 1, 1}, {2, 2, 2}, {3, 3, 3}, {4, 4, 0}, {5, 5, 0}, {6, 6, 0}}
	buildPath(g, verts)
	g.simplifyCompoundPath()
	a.Assert(t, is.Equal(g.getNumPathPointsForTest(), 4))

	pts = g.getAllPathPointsForTest()
	a.DeepEqual(t, pts[0], verts[0])
	a.DeepEqual(t, pts[1], verts[3])
	a.DeepEqual(t, pts[2], verts[4])
	a.DeepEqual(t, pts[3], verts[6])

	// 4  + 3 + 2colinear vertices.
	g.reset()
	verts = []geom.Pt3{{0, 0, 0}, {1, 1, 1}, {2, 2, 2}, {3, 3, 3}, {4, 4, 0}, {5, 5, 0}, {6, 6, 0},
		{7, 0, 0}, {8, 0, 0}}
	buildPath(g, verts)
	g.simplifyCompoundPath()
	a.Assert(t, is.Equal(g.getNumPathPointsForTest(), 6))

	pts = g.getAllPathPointsForTest()
	a.DeepEqual(t, pts[0], verts[0])
	a.DeepEqual(t, pts[1], verts[3])
	a.DeepEqual(t, pts[2], verts[4])
	a.DeepEqual(t, pts[3], verts[6])
	a.DeepEqual(t, pts[4], verts[7])
	a.DeepEqual(t, pts[5], verts[8])
}

// Return the total number of points in the current path. Each arc counts for a single point.
// Useful for unit testing.
func (g *grblGenerator) getNumPathPointsForTest() int {
	count := 0
	for i, s := range g.path {
		if s.isArcComponent() {
			count++
		} else {
			if i == 0 {
				count = len(s.points)
			} else {
				count = count + len(s.points) - 1
			}
		}
	}

	return count
}

// Return all the points in the current path as a single array. For arc, only the endpoint
// is produced in that array. Useful for unit testing.
func (g *grblGenerator) getAllPathPointsForTest() []pt3 {
	if len(g.path) == 1 && g.path[0].isLineSegmentComponent() {
		return g.path[0].points
	}

	allPoints := make([]pt3, 0, 200)
	for i, s := range g.path {
		if s.isLineSegmentComponent() {
			if i == 0 {
				allPoints = append(allPoints, s.points...)
			} else {
				allPoints = append(allPoints, s.points[1:]...)
			}
		} else {
			allPoints = append(allPoints, s.points[1])
		}
	}

	return allPoints
}
