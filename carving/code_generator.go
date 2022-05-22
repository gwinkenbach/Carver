package carving

import "io"

// codeGenerator defines an interface through which the output code is emitted to a writer.
// An example of code generator is the GRBL generator (grblGenerator).
type codeGenerator interface {
	// Configure the generator with the outpout writer and material dimensions.
	configure(codeWriter io.Writer, matWidth, matHeight, matThickness float64)

	// startJob is called once at the beginning of the carving job.
	startJob()
	// endJob is called once at the end of the carving job.
	endJob()

	// Change the horizontal fee rate to <newFeedRateMmPerMin> and return the old feed rate.
	changeHorizontalFeedRate(newFeedRateMmPerMin float64) float64
	// Change the vertical fee rate to <newFeedRateMmPerMin> and return the old feed rate.
	changeVerticalFeedRate(newFeedRateMmPerMin float64) float64

	// Each carving path constitutes of a series of 3D linear segments. The starting point is
	// set with startPath. The subsequent points along the path are set with moveTo. Finally,
	// the path is terminmated with a call to endPath. If discardPath is true, the generator
	// should discard the entire path instead of emitting code for it.
	startPath(x, y, depth float64)
	moveTo(x, y, depth float64)
	clockwiseArcTo(x, y, depth, radius float64)
	counterclockwiseArcTo(x, y, depth, radius float64)
	endPath(discard bool)
}
