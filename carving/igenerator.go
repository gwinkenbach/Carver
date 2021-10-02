package carving

import "io"

type codeGenerator interface {
	configure(output io.Writer, matWidth, matHeight, matThickness float64)

	startJob()
	endJob()

	startPath(x, y, depth float64)
	moveTo(x, y, depth float64)
	endPath(discard bool)
}
