package carving

import (
	"fmt"
	"io"

	"alvin.com/GoCarver/geom"
)

// A code generator that prints output, for debugging.
type printTestGenerator struct {
	lastX, lastY float64
}

func (g *printTestGenerator) configure(
	output io.Writer,
	matWidth, matHeight, matThickness float64) {
	fmt.Printf("Carve: width=%f, height=%f, depth=%f\n", matWidth, matHeight, matThickness)
}

func (g *printTestGenerator) startJob() {

}

func (g *printTestGenerator) endJob() {

}

func (g *printTestGenerator) startPath(x, y, depth float64) {
	fmt.Printf("start p=(%f, %f) - ", x, y)
	g.lastX, g.lastY = x, y
}

func (g *printTestGenerator) moveTo(x, y, depth float64) {
	g.lastX, g.lastY = x, y
}

func (g *printTestGenerator) endPath(discard bool) {
	fmt.Printf("end p=(%f, %f)\n", g.lastX, g.lastY)
}

// A code generator used for unit testing.
type unitTestGenerator struct {
	gotStart      bool
	firstPoint    geom.Pt2
	firstDepth    float64
	lastPoint     geom.Pt2
	lastDepth     float64
	numPoints     int
	pathCompleted bool
}

var _ codeGenerator = (*unitTestGenerator)(nil)

func (g *unitTestGenerator) configure(
	output io.Writer,
	matWidth, matHeight, matThickness float64) {
}

func (g *unitTestGenerator) changeHorizontalFeedRate(newFeedRateMmPerMin float64) float64 {
	return 400.0
}

func (g *unitTestGenerator) startJob() {

}

func (g *unitTestGenerator) endJob() {

}

func (g *unitTestGenerator) startPath(x, y, depth float64) {
	g.pathCompleted = false
	g.firstDepth = depth
	g.firstPoint = geom.NewPt2(x, y)
	g.numPoints = 1
	g.gotStart = true
}

func (g *unitTestGenerator) moveTo(x, y, depth float64) {
	if g.gotStart {
		g.lastDepth = depth
		g.lastPoint = geom.NewPt2(x, y)
		g.numPoints++
	}
}

func (g *unitTestGenerator) endPath(discard bool) {
	if g.gotStart {
		g.pathCompleted = true
		g.gotStart = false
	}
}
