package carving

import (
	"math"

	"alvin.com/GoCarver/geom"
	"alvin.com/GoCarver/hmap"
)

type xCarvingRun struct {
	carvingRun
}

func (r *xCarvingRun) configure(
	sampler hmap.ScalarGridSampler, // The sampler to get the image value at each point.
	generator codeGenerator, // The output code generator.
	carvingWidth float64, // The width along x of the carving area.
	xAtLeft float64, // The x-coordinate at the left side of each run.
	runY float64, // The y coordinate for this run.
	whiteCarvingDepth float64, // The carving depth for white image samples.
	blackCarvingDepth float64, // The carving depth for black image samples.
	depthStepDown float64, // How much to step down in depth at each pass.
) {
	if carvingWidth < 0 {
		xAtLeft = xAtLeft + carvingWidth
		carvingWidth = -carvingWidth
	}

	// Get the number of available sample along X. We need at least two samples,
	// one at each end of the run.
	p0 := geom.NewPt2(xAtLeft, runY)
	p1 := geom.NewPt2(xAtLeft+carvingWidth, runY)
	numSamples := sampler.GetNumSamplesFromP0ToP1(p0, p1)
	if numSamples <= 1 {
		numSamples = 2
	}

	deltaX := carvingWidth / float64(numSamples-1)
	if deltaX < minStepSize {
		numSamples = int(math.Ceil(carvingWidth/minStepSize)) + 1
		if numSamples <= 1 {
			numSamples = 2
		}

		deltaX = carvingWidth / float64(numSamples-1)
	}

	r.sampler = sampler
	r.generator = generator

	r.numSteps = numSamples - 1
	r.step = geom.NewVec2(deltaX, 0)
	r.startingPoint = p0
	r.endPoint = p1

	r.blackCarvingDepth = blackCarvingDepth
	r.whiteCarvingDepth = whiteCarvingDepth
	r.depthStepDown = depthStepDown
	r.currentCarvingDepth = 0.0

	r.needMorePasses = true

	r.sanitize()
}
