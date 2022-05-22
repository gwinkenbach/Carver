package carving

import (
	"fmt"
	"io"
	"log"
	"math"

	"alvin.com/GoCarver/geom"
)

type pt3 = geom.Pt3      // Shorthand for Pt3
type componentFlavor int // Enum type for path component flavors.

// Path components are of two flavors:
// - Line segments, defined as an array of points. The first point of each line-segment component
//   is always equal to the last point of the preceding component.
// - A single arc. In this case, the component consists of exactly two points p0, p1
//   as follows:
//    x0 = +1 for clockwise arc, -1 for counterclockwise arc.
//    y0 = radius.
//    z0 = z at start of arc.
//    (x1, y1) = end point of arc.
//    z1 = z at end of arc.
type pathComponent struct {
	flavor componentFlavor
	points []pt3
}

const (
	initialPathBufferSize = 40

	epsilon     = 1e-5
	epsilonSqrd = epsilon * epsilon

	flatnessTolerance     = 0.04
	flatnessToleranceSqrd = flatnessTolerance * flatnessTolerance

	proximityTolerance     = 0.15
	proximityToleranceSqrd = proximityTolerance * proximityTolerance

	lineSegmentsComponent componentFlavor = iota
	arcComponent

	clockwiseArc        = 1.0
	counterclockwiseArc = -1.0

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
	grblClockwiseArcTo        = "G2 X%.2f Y%.2f Z%.2f R%.2f F%.2f"
	grblCounterclockwiseArcTo = "G2 X%.2f Y%.2f Z%.2f R%.2f F%.2f"
	grblGotoZFormat           = "Z%.2f"
)

// grblGenerator implements the codeGenerator interface to generator GRBL code.
type grblGenerator struct {
	horizFeedRate float64
	vertFeedRate  float64

	// A path consists of a series of successive components.
	path          []pathComponent
	startingPoint pt3

	grblCurrentLoc pt3
	grblOut        io.Writer
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

func (g *grblGenerator) changeVerticalFeedRate(newFeedRateMmPerMin float64) float64 {
	retVal := g.vertFeedRate
	g.vertFeedRate = newFeedRateMmPerMin
	return retVal
}

func (g *grblGenerator) startJob() {
	g.reset()
	g.path = g.path[:0] // Empty
	g.genGrblPreamble()
}

func (g *grblGenerator) endJob() {
	g.genGrblEpilogue()
}

func (g *grblGenerator) startPath(x, y, depth float64) {
	if g.path == nil {
		g.path = make([]pathComponent, 0, 8)
	} else {
		g.path = g.path[:0]
	}

	// For now just record the path starting point.
	g.startingPoint = geom.NewPt3(x, y, depth)
}

func (g *grblGenerator) moveTo(x, y, depth float64) {
	g.appendPointToPath(geom.NewPt3(x, y, depth))
}

func (g *grblGenerator) clockwiseArcTo(x, y, depth, radius float64) {
	comp := pathComponent{
		flavor: arcComponent,
		points: make([]geom.Pt3, 2),
	}
	comp.points[0] = geom.NewPt3(clockwiseArc, radius, depth)
	comp.points[1] = geom.NewPt3(x, y, depth)
	g.path = append(g.path, comp)
}

func (g *grblGenerator) counterclockwiseArcTo(x, y, depth, radius float64) {
	comp := pathComponent{
		flavor: arcComponent,
		points: make([]geom.Pt3, 2),
	}
	comp.points[0] = geom.NewPt3(counterclockwiseArc, radius, depth)
	comp.points[1] = geom.NewPt3(x, y, depth)
	g.path = append(g.path, comp)
}

func (g *grblGenerator) endPath(discard bool) {
	if discard {
		if g.path != nil {
			g.path = g.path[:0]
		}
		return
	}

	g.simplifyCompoundPath()
	if len(g.path) > 0 {
		g.emitGrblForCompoundPath()
	}

	g.path = g.path[:0]
}

// Reset the GRBL generator.
func (g *grblGenerator) reset() {
	g.path = g.path[:0] // Empty
}

// Append point q at the end of the path. Individual points are always appended to line-segments
// components, creating one if necessary.
func (g *grblGenerator) appendPointToPath(q pt3) {
	section := g.getPathComponentToAppendPointTo()
	if section.shouldUsePoint(q) {
		section.points = append(section.points, q)
	}
}

// Return the line-segment component to which a new point can be appended. That is either the
// last component in the path, if it is a line-segment component or a new line-segment
// component otherwise.
func (g *grblGenerator) getPathComponentToAppendPointTo() *pathComponent {
	numComponents := len(g.path)

	// If the compound path is empty, create a line-segment component and return it.
	if numComponents == 0 {
		// Create a new line-segment component.
		comp := pathComponent{
			flavor: lineSegmentsComponent,
			points: make([]pt3, 0, initialPathBufferSize),
		}

		// Since it's the very first component in the path, it should have the path starting point.
		comp.points = append(comp.points, g.startingPoint)
		// Add it to the compound path.
		g.path = append(g.path, comp)

		return &g.path[0]
	}

	// If the component at the end of the compound path is a line-segment component, return it.
	if g.path[numComponents-1].flavor == lineSegmentsComponent {
		return &g.path[numComponents-1]
	}

	// Add a new line-segment component at the end of the path and return it.
	comp := pathComponent{
		flavor: lineSegmentsComponent,
		points: make([]geom.Pt3, 0, 16),
	}

	// Add the last point of the previous component as the starting point for this new component.
	comp.points = append(comp.points, g.path[numComponents-1].getComponentEndPoint())

	g.path = append(g.path, comp)
	return &g.path[numComponents]
}

func (g *grblGenerator) simplifyCompoundPath() {
	for i := range g.path {
		g.path[i].simplifyComponent()
	}
}

// Emit the GRBL code to cut a path. That includes safely repositioning the tool to the first
// point in the path.
func (g *grblGenerator) emitGrblForCompoundPath() {
	for i, section := range g.path {
		if section.isLineSegmentComponent() {
			// For the very first section, we must reposition to the first point. For subsequent
			// sections, we must skip the first point as it is a duplicate of the last point
			// from the previous section.
			if i == 0 {
				g.genRepositionToPoint(section.points[0])
			}

			numPoints := len(section.points)
			for j := 1; j < numPoints; j++ {
				g.genLinearMoveToXyz(section.points[j])
			}
		} else {
			// For the very first section, we must reposition to the starting point.
			if i == 0 {
				g.genRepositionToPoint(g.startingPoint)
			}

			p1 := section.points[0]
			p2 := section.points[1]
			if p1.X > 0 {
				g.genClockwiseArcTo(p1.Y, p2)
			} else {
				g.genCounterclockwiseArcTo(p1.Y, p2)
			}
		}
	}
}

func (g *grblGenerator) genRepositionToPoint(p geom.Pt3) {
	// Check whether we need to move at all.
	if g.grblCurrentLoc.X == p.X && g.grblCurrentLoc.Y == p.Y {
		g.genLinearMoveToZ(p.Z)
		return
	}

	// For distances > 50mm, use a rapid move otherwise a linear move.
	if distP0ToP1Sqrd(g.grblCurrentLoc, p) > 2500 {
		g.genLinearMoveToZ(5.0)

		var q pt3
		q.X, q.Y, q.Z = p.X, p.Y, 5.0
		g.genRapidMoveToXyz(q)
		g.genLinearMoveToZ(p.Z)
	} else {
		g.genLinearMoveToZ(1.0)

		var q pt3
		q.X, q.Y, q.Z = p.X, p.Y, 1.0
		g.genLinearMoveToXyz(q)
		g.genLinearMoveToZ(p.Z)
	}
}

func (g *grblGenerator) genGrblPreamble() {
	g.writeStrLn(grblAbsolutePositioning)
	g.writeStrLn(grblSelectPlaneXy)
	g.writeStrLn(grblSetUnitMm)
	g.writeStrLn(grblHome)
	g.writeStrLn(grblAbsolutePositioning)
}

func (g *grblGenerator) genGrblEpilogue() {
	g.genRapidMoveToZ(25)
	g.writeStrLn(grblHome)
	g.writeStrLn(grblEndCode)
}

func (g *grblGenerator) genLinearMoveToZ(z float64) {
	if g.grblCurrentLoc.Z != z {
		fmt.Fprintf(g.grblOut, grblLinearMoveToZFormat+"\n", z, g.vertFeedRate)
		g.grblCurrentLoc.Z = z
	}
}

func (g *grblGenerator) genLinearMoveToXyz(q geom.Pt3) {
	if !g.grblCurrentLoc.Eq(q) {
		fmt.Fprintf(g.grblOut, grblLinearMoveToXyzFormat+"\n", q.X, q.Y, q.Z, g.horizFeedRate)
		g.grblCurrentLoc = q
	}
}

func (g *grblGenerator) genRapidMoveToXyz(q geom.Pt3) {
	if !g.grblCurrentLoc.Eq(q) {
		fmt.Fprintf(g.grblOut, grblRapidMoveToXyzFormat+"\n", q.X, q.Y, q.Z)
		g.grblCurrentLoc = q
	}
}

func (g *grblGenerator) genRapidMoveToZ(z float64) {
	if g.grblCurrentLoc.Z != z {
		fmt.Fprintf(g.grblOut, grblRapidMoveToZFormat+"\n", z)
		g.grblCurrentLoc.Z = z
	}
}

func (g *grblGenerator) genClockwiseArcTo(radius float64, q pt3) {
	if radius > 0 {
		fmt.Fprintf(g.grblOut, grblClockwiseArcTo+"\n", q.X, q.Y, q.Z, radius, g.horizFeedRate)
		g.grblCurrentLoc = q
	}
}

func (g *grblGenerator) genCounterclockwiseArcTo(radius float64, q pt3) {
	if radius > 0 {
		fmt.Fprintf(g.grblOut, grblCounterclockwiseArcTo+"\n", q.X, q.Y, q.Z, radius, g.horizFeedRate)
		g.grblCurrentLoc = q
	}
}

func (g *grblGenerator) writeStrLn(s string) {
	fmt.Fprintf(g.grblOut, "%s\n", s)
}

// Returns whether the component has line-segment flavor.
func (s *pathComponent) isLineSegmentComponent() bool {
	return s.flavor == lineSegmentsComponent
}

// Returns whether the component has arc flavor.
func (s *pathComponent) isArcComponent() bool {
	return s.flavor == arcComponent
}

// Returns whether point q should be added to the line-segment component. This is used to
// eliminate points that can be trivially discarded, such as successive, co-located points.
func (s *pathComponent) shouldUsePoint(q pt3) bool {
	numPoints := len(s.points)
	if numPoints == 0 {
		return true
	}

	lastPathPoint := s.points[numPoints-1]
	return !lastPathPoint.Eq(q)
}

// Simplify the component in-place. Only line-segment component are simplified.
func (s *pathComponent) simplifyComponent() {
	if s.isLineSegmentComponent() {
		s.simplifyPathByFlatness()
		s.simplifyPathByProximity()
	}
}

// Return the endpoint for this component.
func (s *pathComponent) getComponentEndPoint() pt3 {
	if s.flavor == lineSegmentsComponent {
		n := len(s.points)
		if n == 0 {
			log.Fatalln("pathComponent: empty linear-section")
		}
		return s.points[n-1]
	}

	if len(s.points) != 2 {
		log.Fatalln("pathComponent: an arc section should have exactly two points")
	}
	return s.points[1]
}

// Simplify the path using a flatness criterion. Points that are almost colinear are coalesced
// into line segments.
func (s *pathComponent) simplifyPathByFlatness() {
	fmt.Printf("simplifyPathByFlatness: In points = %v\n", s.points)
	if len(s.points) < 3 {
		return
	}

	// Function keepPoint and related keepIndex are used to keep track of points we do not discard,
	// in-place within g.currentPath.
	keepIndex := 0
	keepPoint := func(i int) {
		if i != keepIndex {
			s.points[keepIndex] = s.points[i]
		}
		keepIndex++
	}

	// Start at the begining of the path and accumulate vertices that are colinear.
	p0 := 0
	p1 := 2
	keepPoint(p0)
	for p1 < len(s.points) {
		// For all the points q between p0 and p1, measure the max distance from q to line p0-p1.
		q := p0 + 1
		distSqrd := 0.0
		for q < p1 {
			ptQ := s.points[q]
			ptP0 := s.points[p0]
			ptP1 := s.points[p1]

			d := distQtoP0P1Sqrd(ptQ, ptP0, ptP1)
			distSqrd = math.Max(distSqrd, d)
			q++
		}

		if distSqrd > flatnessToleranceSqrd {
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
	if p0 < len(s.points) {
		keepPoint(len(s.points) - 1)
	}

	s.points = s.points[:keepIndex]
	fmt.Printf("simplifyPathByFlatness: Out points = %v\n", s.points)
}

// Simplify the path using proximity criterion. Points that are too close to eachother,
// along X or Y are coalesced.
func (s *pathComponent) simplifyPathByProximity() {
	if len(s.points) < 3 {
		return
	}

	// Function keepPoint and related keepIndex are used to keep track of points we do not discard,
	// in-place within g.currentPath.
	keepIndex := 0
	keepPoint := func(q geom.Pt3) {
		if keepIndex > 0 {
			if s.points[keepIndex-1].Eq(q) {
				return
			}
		}
		s.points[keepIndex] = q
		keepIndex++
	}

	updateLastPoint := func(q geom.Pt3) {
		if keepIndex > 0 {
			s.points[keepIndex-1] = q
		}
	}

	p0 := 0
	p1 := 1
	n := len(s.points)
	q0 := s.points[p0]

	keepPoint(q0)
	for p1 < n {
		q1 := s.points[p1]

		xy1 := geom.NewPt2(q1.X, q1.Y)
		xy0 := geom.NewPt2(q0.X, q0.Y)
		d := xy1.Sub(xy0).LenSq()
		if d < proximityToleranceSqrd {
			maxZ := math.Max(q0.Z, q1.Z)
			if p1 == n-1 {
				// p1 is the last point along the path; keep it.
				q1.Z = maxZ
				updateLastPoint(q1)
			} else {
				q0.Z = maxZ
				updateLastPoint(q0)
			}
		} else {
			keepPoint(q1)
			p0 = p1
			q0 = q1
		}

		p1++
	}

	s.points = s.points[:keepIndex]
}

// Return the distance from p0 to p1 squared.
func distP0ToP1Sqrd(p0, p1 geom.Pt3) float64 {
	return p0.Sub(p1).LenSq()
}

// Return the distance from q to line p0-p1, squared. If p0 and p1 are coincident, returns
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
