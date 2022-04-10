package carving

import (
	"fmt"
	"io"
	"math"

	"alvin.com/GoCarver/geom"
)

const (
	initialPathBufferSize = 200

	epsilon     = 1e-5
	epsilonSqrd = epsilon * epsilon

	tolerance     = 0.06
	toleranceSqrd = tolerance * tolerance

	grblAbsolutePositioning = "G90"
	grblSelectPlaneXy       = "G17"
	grblSetUnitMm           = "G21"
	grblHome                = "G28 G91 Z0"
	grblSelectTool1         = "T1 M6"
	grblEndCode             = "M30"

	grblRapidMoveToXyzFormat  = "G0 X%.2f Y%.2f Z%.2f"
	grblRapidMoveToXyFormat   = "G0 X%.2f Y%.2f"
	grblRapidMoveToZFormat    = "G0 Z%.2f"
	grblLinearMoveToZFormat   = "G1 Z%.2f F%.2f"
	grblLinearMoveToXyFormat  = "G1 X%.2f Y%.2f F%.2f"
	grblLinearMoveToXyzFormat = "G1 X%.2f Y%.2f Z%.2f F%.2f"
	grblGotoZFormat           = "Z%.2f"
)

type pt3 = geom.Pt3

// grblGenerator implements the codeGenerator interface to generator GRBL code.
type grblGenerator struct {
	horizFeedRate float64
	vertFeedRate  float64

	currentPath []pt3
	currentLoc  pt3

	grblOut io.Writer
}

var _ codeGenerator = (*grblGenerator)(nil)

func newGrblGenerator(horizFeedRate, vertFeedRate float64) *grblGenerator {
	return &grblGenerator{
		horizFeedRate: horizFeedRate,
		vertFeedRate:  vertFeedRate,
	}
}

func (g *grblGenerator) configure(
	output io.Writer,
	matWidth, matHeight, matThickness float64) {

	g.grblOut = output
}

func (g *grblGenerator) changeHorizontalFeedRate(newFeedRateMmPerMin float64) float64 {
	retVal := g.horizFeedRate
	g.horizFeedRate = newFeedRateMmPerMin
	return retVal
}

func (g *grblGenerator) startJob() {
	g.reset()
	g.currentPath = g.currentPath[:0] // Empty
	g.writeGrblPreamble()
}

func (g *grblGenerator) endJob() {
	g.writeGrblEpilogue()
}

func (g *grblGenerator) startPath(x, y, depth float64) {
	if g.currentPath == nil {
		g.currentPath = make([]pt3, 0, initialPathBufferSize)
	} else {
		g.currentPath = g.currentPath[:0]
	}

	g.addPathPoint(geom.NewPt3(x, y, depth))
}

func (g *grblGenerator) moveTo(x, y, depth float64) {
	g.addPathPoint(geom.NewPt3(x, y, depth))
}

func (g *grblGenerator) endPath(discard bool) {
	if len(g.currentPath) < 2 || discard {
		if g.currentPath != nil {
			g.currentPath = g.currentPath[:0]
		}
		return
	}

	g.simplifyPath()
	if len(g.currentPath) > 1 {
		g.emitGrblForCurrentPath()
	}

	g.currentPath = g.currentPath[:0]
}

// Reset the GRBL generator.
func (g *grblGenerator) reset() {
	g.currentPath = g.currentPath[:0] // Empty
}

// Add point q to path unless it can be trivially discarded.
func (g *grblGenerator) addPathPoint(q pt3) {
	if g.shouldUsePoint(q) {
		g.currentPath = append(g.currentPath, q)
	}
}

// Returns whether point q should be added to the path. This is used to eliminate points
// that can be trivially discarded, such as successive, co-located points.
func (g *grblGenerator) shouldUsePoint(q pt3) bool {
	numPoints := len(g.currentPath)
	if numPoints == 0 {
		return true
	}

	lastPathPoint := g.currentPath[numPoints-1]
	if lastPathPoint.Eq(q) {
		return false
	}

	return true
}

// Simplify the path in g.currentPath. Points that are almost colinear are coalesced into
// line segments.
func (g *grblGenerator) simplifyPath() {
	if len(g.currentPath) < 3 {
		return
	}

	// Function keepPoint and related keepIndex are used to keep track of points we do not discrad,
	// in-place within g.currentPath.
	keepIndex := 0
	keepPoint := func(i int) {
		if i != keepIndex {
			g.currentPath[keepIndex] = g.currentPath[i]
		}
		keepIndex++
	}

	// Start at the begining of the path and accumulate vertices that are colinear.
	p0 := 0
	p1 := 2
	keepPoint(p0)
	for p1 < len(g.currentPath) {
		// For all the points q between p0 and p1, measure the max distance from q to line p0-p1.
		q := p0 + 1
		distSqrd := 0.0
		for q < p1 {
			d := distQtoP0P1Sqrd(g.currentPath[q], g.currentPath[p0], g.currentPath[p1])
			distSqrd = math.Max(distSqrd, d)
			q++
		}

		if distSqrd > toleranceSqrd {
			// Some point q between p0 and p1 is out of tolerance. We need to keep p0, p1-1 and discard
			// all the points between p0 and p1-1. Then we start a new colinear-point accumulation at
			// p1-1.
			// We keep p0 at the begining of each new colinear accumulation. No need to do it again.
			p0 = p1 - 1
			keepPoint(p0)
			p1 = p0 + 1
		}

		p1++
	}

	// All remaining points since last p0 are colinear. Only keep the last one.
	p0++
	if p0 < len(g.currentPath) {
		keepPoint(len(g.currentPath) - 1)
	}

	g.currentPath = g.currentPath[:keepIndex]
}

// Emit the GRBL code to cut a path. That includes safely repositioning the tool to the first
// point in the path.
func (g *grblGenerator) emitGrblForCurrentPath() {
	if len(g.currentPath) <= 1 {
		return
	}

	var i int
	var v geom.Pt3
	for i, v = range g.currentPath {
		if i == 0 {
			// First point in path: reposition the tool.
			g.repositionToPoint(v)
		} else {
			g.linearMoveToXyz(v)
		}
	}
}

func (g *grblGenerator) repositionToPoint(p geom.Pt3) {
	// Check whether we need to move at all.
	if g.currentLoc.X == p.X && g.currentLoc.Y == p.Y {
		g.linearMoveToZ(p.Z)
		return
	}

	// For distances > 50mm, use a rapid move otherwise a linear move.
	if distP0ToP1Sqrd(g.currentLoc, p) > 2500 {
		g.linearMoveToZ(5)

		q := p
		q.Z = 5
		g.rapidMoveToXyz(q)

		g.linearMoveToZ(p.Z)
	} else {
		g.linearMoveToZ(1)

		q := p
		q.Z = 1
		g.linearMoveToXyz(q)

		g.linearMoveToZ(p.Z)
	}
}

func (g *grblGenerator) writeGrblPreamble() {
	g.writeStrLn(grblAbsolutePositioning)
	g.writeStrLn(grblSelectPlaneXy)
	g.writeStrLn(grblSetUnitMm)
	g.writeStrLn(grblHome)
	g.writeStrLn(grblAbsolutePositioning)
}

func (g *grblGenerator) writeGrblEpilogue() {
	g.rapidMoveToZ(25)
	g.writeStrLn(grblHome)
	g.writeStrLn(grblEndCode)
}

func (g *grblGenerator) linearMoveToZ(z float64) {
	if g.currentLoc.Z != z {
		fmt.Fprintf(g.grblOut, grblLinearMoveToZFormat+"\n", z, g.vertFeedRate)
		g.currentLoc.Z = z
	}
}

func (g *grblGenerator) linearMoveToXyz(q geom.Pt3) {
	if !g.currentLoc.Eq(q) {
		fmt.Fprintf(g.grblOut, grblLinearMoveToXyzFormat+"\n", q.X, q.Y, q.Z, g.horizFeedRate)
		g.currentLoc = q
	}
}

func (g *grblGenerator) rapidMoveToXyz(q geom.Pt3) {
	if !g.currentLoc.Eq(q) {
		fmt.Fprintf(g.grblOut, grblRapidMoveToXyzFormat+"\n", q.X, q.Y, q.Z)
		g.currentLoc = q
	}
}

func (g *grblGenerator) rapidMoveToZ(z float64) {
	if g.currentLoc.Z != z {
		fmt.Fprintf(g.grblOut, grblRapidMoveToZFormat+"\n", z)
		g.currentLoc.Z = z
	}
}

func (g *grblGenerator) writeStrLn(s string) {
	fmt.Fprintf(g.grblOut, "%s\n", s)
}

// Return the distance from p0 to p1 squared.
func distP0ToP1Sqrd(p0, p1 geom.Pt3) float64 {
	v := p0.Sub(p1)
	return v.Dot(v)
}

// Return the distance from q to line p0-p1 squared. If p0 and p1 are coincident, returns
// the distance from q to p0 squared.
func distQtoP0P1Sqrd(q, p0, p1 pt3) float64 {
	v0 := q.Sub(p0)
	w := p1.Sub(p0)
	lSqrd := w.Dot(w)

	if lSqrd < epsilonSqrd {
		return v0.Dot(v0)
	}

	s := v0.Dot(w) / lSqrd
	d := v0.Sub(w.Scale(s))
	return d.Dot(d)
}
