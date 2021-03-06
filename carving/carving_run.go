package carving

import (
	"log"
	"math"

	"alvin.com/GoCarver/geom"
	g "alvin.com/GoCarver/geom"
	"alvin.com/GoCarver/hmap"
)

type oneRun interface {
	isDone() bool
	setEnableCarvingAtFulldepth(enable bool)
	doOnePass(delta float64)
}

var maxDepth = 0.0

// carvingRun represent a single rectilinear run of the carving tool. It is used to manage
// the multiple carving passes that are required to carve the material to the maximum depth
// given the maximum step-down size.
type carvingRun struct {
	numSteps      int    // Number of steps along the run.
	step          g.Vec2 // Increment vector for each step.
	startingPoint g.Pt2  // Starting point for this run.
	endPoint      g.Pt2  // End point for this run.

	whiteCarvingDepth   float64 // The carving depth for white samples.
	blackCarvingDepth   float64 // The carving depth for black samples.
	currentCarvingDepth float64 // The current carving depth, always starting at 0.
	depthStepDown       float64 // How much to step down for each new pass.

	enableCarveAtFullDepth bool

	needMorePasses bool // Whether more passes are need to finish this run.

	sampler   hmap.ScalarGridSampler
	generator codeGenerator
}

var _ oneRun = (*carvingRun)(nil)

// isDone returns whether the maximum carving depth has been achieved and no more carving
// passes are needed.
func (r *carvingRun) isDone() bool {
	return !r.needMorePasses
}

// setEnableCarvingAtFulldepth is used to enable at full depth, ignoring the maximum
// step-down size.
func (r *carvingRun) setEnableCarvingAtFulldepth(enable bool) {
	r.enableCarveAtFullDepth = enable
}

// doOnePass is called to generate one carving pass along the run. Parameter delta must be
// either +1 or -1. It determines wether the run goes forward or backward along the run.
func (r *carvingRun) doOnePass(delta float64) {
	if !r.needMorePasses {
		return
	}

	if math.Abs(delta) != 1.0 {
		log.Fatalln("Invalid delta value, should be 1.0 or -1.0")
	}

	// Check wether the carving depth reaches below the old carving depth. If it doesn't we
	// can discard the path. This is mostly useful on the very fisrt pass.
	oldCarvingDepth := r.currentCarvingDepth
	discardPath := true

	// If the carving depth doesn't go as deep as the deepest sampled carving depth,
	// we'll need more passes.
	r.needMorePasses = false
	r.currentCarvingDepth = r.currentCarvingDepth - r.depthStepDown

	// fmt.Printf("*** Run y=%f, carving depth = %f, delta = %2.0f\n", r.startingPoint.Y, r.currentCarvingDepth, delta)
	// fmt.Printf("        black=%5.2f, white=%5.2f\n", r.blackCarvingDepth, r.whiteCarvingDepth)

	var origin geom.Pt2
	for s := 0; s < r.numSteps; s++ {
		var depth = 0.0
		var clipped = false

		if s == 0 {
			// First step: starting point depends on run direction.
			pt := r.startingPoint
			if delta < 0 {
				pt = r.endPoint
			}

			origin = pt

			depth, clipped = r.getCarvingDepthAt(pt)
			r.needMorePasses = r.needMorePasses || clipped
			if depth < oldCarvingDepth {
				discardPath = false
			}

			r.generator.startPath(pt.X, pt.Y, depth)
			// fmt.Printf("  Start: %4.1f, %4.1f, %4.1f\n", pt.X, pt.Y, depth)
		} else if s == r.numSteps-1 {
			// Last step: end point depends on direction.
			pt := r.startingPoint
			if delta > 0 {
				pt = r.endPoint
			}

			depth, clipped = r.getCarvingDepthAt(pt)
			r.needMorePasses = r.needMorePasses || clipped
			if depth < oldCarvingDepth {
				discardPath = false
			}

			r.generator.moveTo(pt.X, pt.Y, depth)
			r.generator.endPath(discardPath)

			// fmt.Printf("  End: %4.1f, %4.1f, depth = %4.1f, discard = %v, more = %v\n", pt.X, pt.Y, depth, discardPath, r.needMorePasses)
		} else {
			stepVec := r.step.Scale(float64(s) * delta)
			pt := origin.Add(stepVec)
			depth, clipped = r.getCarvingDepthAt(pt)
			r.needMorePasses = r.needMorePasses || clipped
			if depth < oldCarvingDepth {
				discardPath = false
			}

			r.generator.moveTo(pt.X, pt.Y, depth)
		}
	}
}

// getCarvingDepthAt samples and returns the carving depth at the given location. This
// function takes into account whether carving-at-full-depth is enabled.
func (r *carvingRun) getCarvingDepthAt(q geom.Pt2) (depth float64, clipped bool) {
	s := r.sampler.At(q)
	d := (1-s)*r.blackCarvingDepth + s*r.whiteCarvingDepth

	if d < maxDepth {
		maxDepth = d
	}

	depth = d

	if r.enableCarveAtFullDepth {
		clipped = false
	} else {
		clipped = d < r.currentCarvingDepth-0.05
		// fmt.Printf("     target depth = %f5.2f, clip = %v\n", depth, clipped)
		if clipped {
			depth = r.currentCarvingDepth
		}
	}

	return
}

// Sanitize the carving run, ensuring that it is in a valid configuration.
func (r *carvingRun) sanitize() {
	if r.numSteps <= 0 {
		r.numSteps = 1
	}

	if r.whiteCarvingDepth > 0 {
		r.whiteCarvingDepth = 0
	}

	if r.blackCarvingDepth > 0 {
		r.blackCarvingDepth = 0
	}

	if r.depthStepDown < 0 {
		r.depthStepDown = -r.depthStepDown
	}
}
