package carving

import (
	"math"

	"alvin.com/GoCarver/geom"
	"alvin.com/GoCarver/hmap"
)

type yCarvingRun struct {
	carvingRun
}

func (r *yCarvingRun) configure(
	sampler hmap.ScalarGridSampler, // The sampler to get the image value at each point.
	generator codeGenerator, // The output code generator.
	carvingHeight float64, // The height along y of the carving area.
	yAtBottom float64, // The y-coordinate at the bottom side of each run.
	runX float64, // The x coordinate for this run.
	whiteCarvingDepth float64, // The carving depth for white image samples.
	blackCarvingDepth float64, // The carving depth for black image samples.
	depthStepDown float64, // How much to step down in depth at each pass.
) {
	if carvingHeight < 0 {
		yAtBottom = yAtBottom + carvingHeight
		carvingHeight = -carvingHeight
	}

	// Get the number of available sample along X. We need at least two samples,
	// one at each end of the run.
	p0 := geom.NewPt2(runX, yAtBottom)
	p1 := geom.NewPt2(runX, yAtBottom+carvingHeight)
	numSamples := sampler.GetNumSamplesFromP0ToP1(p0, p1)
	if numSamples <= 1 {
		numSamples = 2
	}

	deltaY := carvingHeight / float64(numSamples-1)
	if deltaY < minStepSize {
		numSamples = int(math.Ceil(carvingHeight/minStepSize)) + 1
		if numSamples <= 1 {
			numSamples = 2
		}

		deltaY = carvingHeight / float64(numSamples-1)
	}

	r.sampler = sampler
	r.generator = generator

	r.numSteps = numSamples - 1
	r.step = geom.NewVec2(0, deltaY)
	r.startingPoint = p0
	r.endPoint = p1

	r.blackCarvingDepth = blackCarvingDepth
	r.whiteCarvingDepth = whiteCarvingDepth
	r.depthStepDown = depthStepDown
	r.currentCarvingDepth = 0.0

	r.needMorePasses = true

	r.sanitize()
}
