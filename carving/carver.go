package carving

import (
	"io"
	"math"

	g "alvin.com/GoCarver/geom"
	"alvin.com/GoCarver/hmap"
)

const (
	ToolTypeBall = 1
	ToolTypeFlat = 2

	CarveModeXOnly  = 100
	CarveModeYOnly  = 101
	CarveModeXThenY = 102

	minStepSize = 0.1 // Minimum step size in mm, a.k.a. resolution.
)

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

	sampler hmap.ScalarGridSampler
	output  io.Writer
}

func NewCarver(output io.Writer) *Carver {
	return &Carver{output: output}
}

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

func (c *Carver) ConfigureTool(
	toolType int,
	toolDiameterMm float64,
	horizontalFeedRateMmPerMin float64,
	verticalFeedRateMmPerMin float64) {

	c.toolDiameterMm = toolDiameterMm
	c.horizFeedRate = horizontalFeedRateMmPerMin
	c.vertFeedRate = verticalFeedRateMmPerMin
}

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

func (c *Carver) Run() {
	gen := newGrblGenerator(c.horizFeedRate, c.vertFeedRate)
	gen.configure(c.output, c.materialDimMm.W, c.materialDimMm.H, 0)

	gen.startJob()
	c.carveAlongX(gen)
	c.carveAlongY(gen)
	gen.endJob()
}

func (c *Carver) carveAlongX(gen codeGenerator) {
	if c.carveMode != CarveModeXOnly && c.carveMode != CarveModeXThenY {
		return
	}

	runs := c.setupXRuns(gen)
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

func (c *Carver) carveAlongY(gen codeGenerator) {
	if c.carveMode != CarveModeYOnly && c.carveMode != CarveModeXThenY {
		return
	}

	runs := c.setupYRuns(gen, c.carveMode == CarveModeXThenY)
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

func (c *Carver) setupXRuns(gen codeGenerator) []oneRun {
	numRuns := c.getNumRunsNeeded(c.carvingDimMm.H)
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
		runs[i] = run
	}

	return runs
}

func (c *Carver) setupYRuns(gen codeGenerator, carveAtFulldepth bool) []oneRun {
	numRuns := c.getNumRunsNeeded(c.carvingDimMm.W)
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
func (c *Carver) getNumRunsNeeded(distToCover float64) int {
	// The first run has a fraction of cutting thickness that is nor shared with other runs.
	// The last run has a cutting path that is not shared. Together, they are exactly one tool
	// diameter in thickness. The rest of the carving distance is divided by the step size.
	cutSize := c.toolDiameterMm * c.stepOverFraction
	numSteps := (distToCover-c.toolDiameterMm)/cutSize - 0.001 + 1.0
	if numSteps < 0 {
		numSteps = 0
	}

	count := int(math.Ceil(numSteps))

	// fmt.Printf("H=%f, step size = %f, num steps = %f or %d\n", distToCover, stepSize, numSteps, count)

	return count
}
