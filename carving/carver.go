package carving

import (
	"io"
	"math"

	g "alvin.com/GoCarver/geom"
	"alvin.com/GoCarver/hmap"
)

const (
	ToolTypeBallPoint = 1
	ToolTypeFlat      = 2

	CarveModeXOnly  = 100
	CarveModeYOnly  = 101
	CarveModeXThenY = 102

	FinishPassModeAlongFirstDirOnly = 200
	FinishPassModeAlongLastDirOnly  = 201
	FinishPassModeAlongAllDirs      = 202

	minStepSize = 0.1 // Minimum step size in mm, a.k.a. resolution.
)

// Carver provides support for generating carving code.
// Usage:
// 1. Create a carver with NewCarver.
// 2. Configure the various carving parameters with the ConfigureXXX functions.
// 3. Generate the carving code by calling the Run function.
type Carver struct {
	materialDimMm     g.Size2
	materialTopMm     float64
	carvingBottomLeft g.Pt2
	carvingDimMm      g.Size2

	carveMode int

	zWhite      float64 // Z coordinate for white samples.
	zBlack      float64 // Z coordinate for black samples.
	maxStepDown float64

	toolDiameterMm   float64
	stepOverFraction float64
	horizFeedRate    float64
	vertFeedRate     float64

	enableFinishingPass        bool
	finishingPassStepFraction  float64
	finishingPassMode          int
	finishingPassHorizFeedRate float64

	sampler hmap.ScalarGridSampler
	output  io.Writer
}

// NewCarver creates and return a new carver that outputs carving code to the
// given writer.
func NewCarver(output io.Writer) *Carver {
	return &Carver{output: output}
}

// ConfigureMaterial is used to configure the carving-material parameters.
func (c *Carver) ConfigureMaterial(
	materialDimMm g.Size2,
	carvingAreaOrigin g.Pt2,
	carvingAreaDimMm g.Size2,
	materialTopMm float64) {

	c.materialDimMm = materialDimMm
	c.carvingBottomLeft = carvingAreaOrigin
	c.carvingDimMm = carvingAreaDimMm
	c.materialTopMm = materialTopMm
}

// ConfigureTool is used to configure the carving-tool parameters.
func (c *Carver) ConfigureTool(
	toolType int,
	toolDiameterMm float64,
	horizontalFeedRateMmPerMin float64,
	verticalFeedRateMmPerMin float64) {

	c.toolDiameterMm = toolDiameterMm
	c.horizFeedRate = horizontalFeedRateMmPerMin
	c.vertFeedRate = verticalFeedRateMmPerMin
}

// ConfigureCarvingProfile is used to configure the carving-profile parameters,
// such as the height of the material, carving mode, etc.
func (c *Carver) ConfigureCarvingProfile(
	sampler hmap.ScalarGridSampler,
	topHeightMm float64,
	bottomHeightMm float64,
	stepOverFraction float64,
	maxStepDownSizeMm float64,
	carveMode int) {

	c.carveMode = carveMode
	c.sampler = sampler
	c.stepOverFraction = stepOverFraction
	c.zWhite = topHeightMm - c.materialTopMm
	c.zBlack = bottomHeightMm - c.materialTopMm
	c.maxStepDown = maxStepDownSizeMm
}

// Configure the finishing pass. When enabled, the finishing pass is the very last carving pass
// in either direction. It runs once at full depth with the step-over reduced to the given
// fraction.
func (c *Carver) ConfigureFinishingPass(
	enabled bool, stepOverFraction float64, carverFinishMode int, horizFeedRate float64) {
	if stepOverFraction >= 1.0 || stepOverFraction < 0.01 {
		c.enableFinishingPass = false
		return
	}

	c.enableFinishingPass = enabled
	c.finishingPassHorizFeedRate = horizFeedRate
	c.finishingPassMode = carverFinishMode
	c.finishingPassStepFraction = stepOverFraction
}

// Run is called to generate the carving code. It is ok to (re)configure the carver and
// call Run multiple times. However, all output go to the same writer.
func (c *Carver) Run() {
	gen := newGrblGenerator(c.horizFeedRate, c.vertFeedRate)
	gen.configure(c.output, c.materialDimMm.W, c.materialDimMm.H, 0)

	gen.startJob()
	c.carveAlongX(gen)
	c.carveAlongY(gen)
	gen.endJob()
}

// Generate carving runs along the x-direction. This will generate the main carving passes as
// well as the optional finishing pass if carving only takes place along X.
func (c *Carver) carveAlongX(gen codeGenerator) {
	if c.carveMode != CarveModeXOnly && c.carveMode != CarveModeXThenY {
		return
	}

	c.genCarvingRunsAlongX(c.stepOverFraction, false /* not full depth */, gen)
	if c.needFinishingPassAlongX() {
		oldFeedRate := gen.changeHorizontalFeedRate(c.finishingPassHorizFeedRate)
		c.genCarvingRunsAlongX(c.finishingPassStepFraction, true /* full depth */, gen)
		gen.changeHorizontalFeedRate(oldFeedRate)
	}
}

// Generating a series of carving path in the x-direction to fully cover the entire
// carving area for given the step-over fraction. Usually, each run consists of several
// passes, determined by the max step-down size. However, when <carveAtFullDepth> is true,
// a single pass at full depth  along each run. Carving at full depth should only be used
// for the very last pass.
func (c *Carver) genCarvingRunsAlongX(
	stepOverFraction float64, carveAtFullDepth bool, gen codeGenerator) {

	runs := c.setupXRuns(stepOverFraction, gen, carveAtFullDepth)
	stepDir := 1.0
	iRun := -1
	for {
		nextRun := c.findNextUnfinishedRun(iRun, runs)
		if nextRun == -1 {
			break
		}

		c.genMoveToNextRun(iRun, nextRun)

		iRun = nextRun
		run := runs[iRun]
		run.doOnePass(stepDir)

		// Flip the step direction after each run.
		if stepDir == 1.0 {
			stepDir = -1.0
		} else {
			stepDir = 1.0
		}
	}
}

// Generate carving runs along the y-direction. This will generate the main carving passes as
// well as the optional finishing pass if carving only takes place along X.
func (c *Carver) carveAlongY(gen codeGenerator) {
	if c.carveMode != CarveModeYOnly && c.carveMode != CarveModeXThenY {
		return
	}

	c.genCarvingRunsAlongY(c.stepOverFraction, c.carveMode == CarveModeXThenY, gen)
	if c.needFinishingPassAlongY() {
		oldFeedRate := gen.changeHorizontalFeedRate(c.finishingPassHorizFeedRate)
		c.genCarvingRunsAlongY(c.finishingPassStepFraction, true /* full depth */, gen)
		gen.changeHorizontalFeedRate(oldFeedRate)
	}
}

// Generating a series of carving path in the y-direction to fully cover the entire
// carving area for given the step-over fraction. Usually, each run consists of several
// passes, determined by the max step-down size. However, when <carveAtFullDepth> is true,
// a single pass at full depth  along each run. Carving at full depth should only be used
// for the very last pass.
func (c *Carver) genCarvingRunsAlongY(
	stepOverFraction float64, carveAtFullDepth bool, gen codeGenerator) {

	runs := c.setupYRuns(stepOverFraction, gen, carveAtFullDepth)
	stepDir := 1.0
	iRun := -1
	for {
		nextRun := c.findNextUnfinishedRun(iRun, runs)
		if nextRun == -1 {
			break
		}

		c.genMoveToNextRun(iRun, nextRun)

		iRun = nextRun
		run := runs[iRun]
		run.doOnePass(stepDir)

		// Flip the step direction after each run.
		if stepDir == 1.0 {
			stepDir = -1.0
		} else {
			stepDir = 1.0
		}
	}
}

// Find and return the index to the next unfinished run after fromRun. If fromRun is
// -1, it means we're starting a new set of runs. Returns -1 if no such run is found.
func (c *Carver) findNextUnfinishedRun(fromRun int, runs []oneRun) int {
	startIndex := 0
	if fromRun >= 0 {
		startIndex = (fromRun + 1) % len(runs)
	}

	i := startIndex
	for {
		if !runs[i].isDone() {
			return i
		}

		i = (i + 1) % len(runs)
		if i == startIndex {
			return -1
		}
	}
}

// Generate a tool move fromRun to toRun. If fromRun == -1, this is the initial move at
// the start of carving a series of runs.
func (c *Carver) genMoveToNextRun(fromRun, toRun int) {
	// TODO: for now grbl generator handles moving from path to path adequately.
}

// Setup the carving runs in the x-direction. Returns an array of x-carving-runs.
func (c *Carver) setupXRuns(
	stepOverFraction float64, gen codeGenerator, carveAtFulldepth bool) []oneRun {

	numRuns := c.getNumRunsNeeded(stepOverFraction, c.carvingDimMm.H)
	if numRuns == 0 {
		return nil
	}

	runs := make([]oneRun, numRuns)

	yStep := 0.0
	if numRuns > 1 {
		yStep = (c.carvingDimMm.H - c.toolDiameterMm) / float64(numRuns-1)
	}

	for i := range runs {
		y := c.carvingBottomLeft.Y + c.toolDiameterMm*0.5 + float64(i)*yStep
		if i == numRuns-1 {
			// Last run, y = Ymax.
			y = c.carvingBottomLeft.Y + c.carvingDimMm.H - c.toolDiameterMm*0.5
		}

		run := &xCarvingRun{}
		run.configure(c.sampler, gen,
			c.carvingDimMm.W-c.toolDiameterMm,
			c.carvingBottomLeft.X+0.5*c.toolDiameterMm,
			y,
			c.zWhite, c.zBlack, c.maxStepDown)

		if carveAtFulldepth {
			run.setEnableCarvingAtFulldepth(true)
		}

		runs[i] = run
	}

	return runs
}

// Setup the carving runs in the y-direction. Returns an array of y-carving-runs.
func (c *Carver) setupYRuns(
	stepOverFraction float64, gen codeGenerator, carveAtFulldepth bool) []oneRun {

	numRuns := c.getNumRunsNeeded(stepOverFraction, c.carvingDimMm.W)
	if numRuns == 0 {
		return nil
	}

	runs := make([]oneRun, numRuns)

	xStep := 0.0
	if numRuns > 1 {
		xStep = (c.carvingDimMm.W - c.toolDiameterMm) / float64(numRuns-1)
	}

	for i := range runs {
		x := c.carvingBottomLeft.X + c.toolDiameterMm*0.5 + float64(i)*xStep
		if i == numRuns-1 {
			// Last run, x = Xmax.
			x = c.carvingBottomLeft.X + c.carvingDimMm.W - c.toolDiameterMm*0.5
		}

		run := &yCarvingRun{}
		run.configure(c.sampler, gen,
			c.carvingDimMm.H-c.toolDiameterMm,
			c.carvingBottomLeft.Y+0.5*c.toolDiameterMm,
			x,
			c.zWhite, c.zBlack, c.maxStepDown)

		if carveAtFulldepth {
			run.setEnableCarvingAtFulldepth(true)
		}

		runs[i] = run
	}

	return runs
}

// Return the number of runs (carving paths) needed to cover distToCover. There is always runs
// just inside both sides of the carving area. Runs are added in the middle to cover the
// entire carving area.
func (c *Carver) getNumRunsNeeded(stepOverFraction, distToCover float64) int {
	// The first run has a fraction of cutting thickness that is nor shared with other runs.
	// The last run has a cutting path that is not shared. Together, they are exactly one tool
	// diameter in thickness. The rest of the carving distance is divided by the step size.
	cutSize := c.toolDiameterMm * stepOverFraction
	numSteps := (distToCover-c.toolDiameterMm)/cutSize - 0.001 + 1.0
	if numSteps < 0 {
		numSteps = 0
	}

	count := int(math.Ceil(numSteps))

	// fmt.Printf("H=%f, step size = %f, num steps = %f or %d\n", distToCover, stepSize, numSteps, count)

	return count
}

// Returns whether a finishing pass is needed in the x-direction.
func (c *Carver) needFinishingPassAlongX() bool {
	// Finishing must be enabled
	if !c.enableFinishingPass {
		return false
	}

	// There's no point in finishing with the same step-over fraction
	if math.Abs(c.finishingPassStepFraction-c.stepOverFraction) < 0.02 {
		return false
	}

	// We must not be carving along Y only.
	if c.carveMode == CarveModeYOnly {
		return false
	}

	// If carving along X only, then any finish mode is fine.
	if c.carveMode == CarveModeXOnly {
		return true
	}

	// If carving in both X and Y directions, then match finish mode.
	return c.carveMode == CarveModeXThenY &&
		(c.finishingPassMode == FinishPassModeAlongFirstDirOnly ||
			c.finishingPassMode == FinishPassModeAlongAllDirs)
}

// Returns whether a finishing pass is needed in the y-direction.
func (c *Carver) needFinishingPassAlongY() bool {
	// Finishing must be enabled
	if !c.enableFinishingPass {
		return false
	}

	// There's no point in finishing with the same step-over fraction
	if math.Abs(c.finishingPassStepFraction-c.stepOverFraction) < 0.02 {
		return false
	}

	// We must not be carving along X only.
	if c.carveMode == CarveModeXOnly {
		return false
	}

	// If carving along Y only, then any finish mode is fine.
	if c.carveMode == CarveModeYOnly {
		return true
	}

	// If carving in both X and Y directions, then match finish mode.
	return c.carveMode == CarveModeXThenY &&
		(c.finishingPassMode == FinishPassModeAlongLastDirOnly ||
			c.finishingPassMode == FinishPassModeAlongAllDirs)
}
