package carving

import (
	"io"

	"alvin.com/GoCarver/geom"
	"alvin.com/GoCarver/hmap"
)

type MaterialConfig struct {
	MaterialDim       geom.Size2
	CarvingAreaOrigin geom.Pt2
	CarvingAreaDim    geom.Size2
	MaterialThickness float64
}

type ToolConfig struct {
	ToolType      int
	ToolDiameter  float64
	HorizFeedRate float64
	VertFeedRate  float64
	MaxStepDown   float64
}

type CarvingConfig struct {
	Sampler             hmap.ScalarGridSampler
	Tool                ToolConfig
	CarvingTopZ         float64
	CarvingBottomZ      float64
	StepOverFraction    float64
	CarvingMode         int
	EnableFinishing     bool
	FinishStepFraction  float64
	FinishHorizFeedRate float64
	FinishMode          int
}

type MachiningConfig struct {
	Material MaterialConfig
	Carving  CarvingConfig
}

func configureCarver(c *Carver, mc *MachiningConfig) {
	c.ConfigureMaterial(mc.Material.MaterialDim, mc.Material.CarvingAreaOrigin,
		mc.Material.CarvingAreaDim, mc.Material.MaterialThickness)

	c.ConfigureTool(mc.Carving.Tool.ToolType, mc.Carving.Tool.ToolDiameter,
		mc.Carving.Tool.HorizFeedRate, mc.Carving.Tool.VertFeedRate)

	c.ConfigureCarvingProfile(mc.Carving.Sampler, mc.Carving.CarvingTopZ, mc.Carving.CarvingBottomZ,
		mc.Carving.StepOverFraction, mc.Carving.Tool.MaxStepDown, mc.Carving.CarvingMode)

	c.ConfigureFinishingPass(mc.Carving.EnableFinishing, mc.Carving.FinishStepFraction,
		mc.Carving.FinishMode, mc.Carving.FinishHorizFeedRate)
}

func DoMachining(config *MachiningConfig, output io.Writer) {
	gen := newGrblGenerator(config.Carving.Tool.HorizFeedRate, config.Carving.Tool.VertFeedRate)
	gen.configure(output, config.Material.MaterialDim.W, config.Material.MaterialDim.H,
		config.Material.MaterialThickness)

	carver := NewCarver(output)
	configureCarver(carver, config)

	gen.startJob()
	carver.Run(gen)
	gen.endJob()
}
