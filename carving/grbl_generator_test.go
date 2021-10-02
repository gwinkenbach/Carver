package carving

import (
	"testing"

	"alvin.com/GoCarver/geom"
	a "gotest.tools/assert"
	is "gotest.tools/assert/cmp"
)

func buildPath(g *grblGenerator, verts []geom.Pt3) {
	for _, v := range verts {
		g.addPathPoint(v)
	}
}

func TestAddPoint(t *testing.T) {
	g := newGrblGenerator(100, 100)

	g.addPathPoint(geom.NewPt3(0, 0, 0))
	if len(g.currentPath) != 1 {
		t.Errorf("addPoint: expected len == 1\n")
	}

	g.addPathPoint(geom.NewPt3(1, 1, 1))
	if len(g.currentPath) != 2 {
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
	g.simplifyPath()
	a.Assert(t, is.Len(g.currentPath, 3))
	for i, q := range g.currentPath {
		a.DeepEqual(t, q, verts[i])
	}

	// 2 vertices.
	g.reset()
	verts = []geom.Pt3{{0, 0, 0}, {1, 1, 1}}
	buildPath(g, verts)
	g.simplifyPath()
	a.Assert(t, is.Len(g.currentPath, 2))
	a.DeepEqual(t, g.currentPath[0], verts[0])
	a.DeepEqual(t, g.currentPath[1], verts[1])

	// 3 colinear vertices.
	g.reset()
	verts = []geom.Pt3{{0, 0, 0}, {1, 1, 1}, {2, 2, 2}}
	buildPath(g, verts)
	g.simplifyPath()
	a.Assert(t, is.Len(g.currentPath, 2))
	a.DeepEqual(t, g.currentPath[0], verts[0])
	a.DeepEqual(t, g.currentPath[1], verts[2])

	// 4 colinear vertices.
	g.reset()
	verts = []geom.Pt3{{0, 0, 0}, {1, 1, 1}, {2, 2, 2}, {3, 3, 3}}
	buildPath(g, verts)
	g.simplifyPath()
	a.Assert(t, is.Len(g.currentPath, 2))
	a.DeepEqual(t, g.currentPath[0], verts[0])
	a.DeepEqual(t, g.currentPath[1], verts[3])

	// 4 colinear vertices and one outlier
	g.reset()
	verts = []geom.Pt3{{0, 0, 0}, {1, 1, 1}, {2, 2, 2}, {3, 3, 3}, {4, 4, 0}}
	buildPath(g, verts)
	g.simplifyPath()
	a.Assert(t, is.Len(g.currentPath, 3))
	a.DeepEqual(t, g.currentPath[0], verts[0])
	a.DeepEqual(t, g.currentPath[1], verts[3])
	a.DeepEqual(t, g.currentPath[2], verts[4])

	// 4 colinear + 3 colinear vertices.
	g.reset()
	verts = []geom.Pt3{{0, 0, 0}, {1, 1, 1}, {2, 2, 2}, {3, 3, 3}, {4, 4, 0}, {5, 5, 0}, {6, 6, 0}}
	buildPath(g, verts)
	g.simplifyPath()
	a.Assert(t, is.Len(g.currentPath, 4))
	a.DeepEqual(t, g.currentPath[0], verts[0])
	a.DeepEqual(t, g.currentPath[1], verts[3])
	a.DeepEqual(t, g.currentPath[2], verts[4])
	a.DeepEqual(t, g.currentPath[3], verts[6])

	// 4  + 3 + 2colinear vertices.
	g.reset()
	verts = []geom.Pt3{{0, 0, 0}, {1, 1, 1}, {2, 2, 2}, {3, 3, 3}, {4, 4, 0}, {5, 5, 0}, {6, 6, 0},
		{7, 0, 0}, {8, 0, 0}}
	buildPath(g, verts)
	g.simplifyPath()
	a.Assert(t, is.Len(g.currentPath, 6))
	a.DeepEqual(t, g.currentPath[0], verts[0])
	a.DeepEqual(t, g.currentPath[1], verts[3])
	a.DeepEqual(t, g.currentPath[2], verts[4])
	a.DeepEqual(t, g.currentPath[3], verts[6])
	a.DeepEqual(t, g.currentPath[4], verts[7])
	a.DeepEqual(t, g.currentPath[5], verts[8])
}
