package model

import (
	"log"

	"alvin.com/GoCarver/geom"
	"alvin.com/GoCarver/qtui"
	"github.com/therecipe/qt/gui"
)

type material struct {
	MaterialWidth      float32 `json:"material_width"`
	MaterialHeight     float32 `json:"material_height"`
	MaterialThickness  float32 `json:"material_thickness"`
	CarvingAreaWidth   float32 `json:"carving_area_width"`
	CarvingAreaHeight  float32 `json:"carving_area_height"`
	CarvingAreaOffsetX float32 `json:"carving_area_offset_x"`
	CarvingAreaOffsetY float32 `json:"carving_area_offset_y"`
	BlackCarvingDepth  float32 `json:"black_carving_depth"`
	WhiteCarvingDepth  float32 `json:"white_carving_depth"`
}

type carving struct {
	ToolDiameter       float32 `json:"tool_diameter"`
	ToolType           int     `json:"tool_type"`
	StepOverPercent    float32 `json:"step_over_percent"`
	MaxStepDownSize    float32 `json:"max_step_down_size"`
	HorizontalFeedRate float32 `json:"horizontal_feed_rate"`
	VerticalFeedRate   float32 `json:"vertical_feed_rate"`
	CarvingMode        int     `json:"carving_mode"`
}

type heightMap struct {
	Image         *gui.QImage `json:"_"` // Ignored in JSON
	ImageFileName string      `json:"imageFileName"`
	ImageMode     int         `json:"imageMode"`
	MirrorY       bool        `json:"mirrorY"`
	MirrorX       bool        `json:"mirrorX"`
}

type modelRoot struct {
	Material  material  `json:"material"`
	Carving   carving   `json:"carving"`
	HeightMap heightMap `json:"height_map"`
}

const (
	ToolTypeBallNose = 0
	ToolTypeStraight = 1

	CarvingModeAlongX      = 0
	CarvingModeAlongY      = 1
	CarvingModeAlongXThenY = 2

	ImageModeFill = geom.ImageModeFill // Stretch image to fill viewport
	ImageModeFit  = geom.ImageModeFit  // Whole image fits in viewport, keep aspect ratio
	ImageModeCrop = geom.ImageModeCrop // Stretch image to fill viewport, keep aspect ratio
)

type Model struct {
	root modelRoot

	fromFilePath string
	dirty        bool
}

func NewModel() *Model {
	return &Model{
		root: modelRoot{
			Material: material{
				MaterialWidth:      100.0,
				MaterialHeight:     100.0,
				MaterialThickness:  15.0,
				BlackCarvingDepth:  -5.0,
				WhiteCarvingDepth:  0.0,
				CarvingAreaWidth:   90.0,
				CarvingAreaHeight:  90.0,
				CarvingAreaOffsetX: 5.0,
				CarvingAreaOffsetY: 5.0,
			},

			Carving: carving{
				ToolDiameter:       3.175, // millimeters
				StepOverPercent:    40,    // Percent of tool diameter
				MaxStepDownSize:    0.5,
				HorizontalFeedRate: 500.0, // millimeters per minute
				VerticalFeedRate:   300.0, // millimeters per minutes
				CarvingMode:        CarvingModeAlongX,
			},
		},
	}
}

func (m *Model) SetDirty(dirty bool) {
	m.dirty = dirty
}

func (m *Model) GetFloat32(tag string) float32 {
	switch tag {
	case qtui.ItemMaterialWidth:
		return m.root.Material.MaterialWidth
	case qtui.ItemMaterialHeight:
		return m.root.Material.MaterialHeight
	case qtui.ItemMaterialThickness:
		return m.root.Material.MaterialThickness
	case qtui.ItemCarvingAreaWidth:
		return m.root.Material.CarvingAreaWidth
	case qtui.ItemCarvingAreaHeight:
		return m.root.Material.CarvingAreaHeight
	case qtui.ItemCarvingAreaOffsetX:
		return m.root.Material.CarvingAreaOffsetX
	case qtui.ItemCarvingAreaOffsetY:
		return m.root.Material.CarvingAreaOffsetY
	case qtui.ItemBlackCarvingDepth:
		return m.root.Material.BlackCarvingDepth
	case qtui.ItemWhiteCarvingDepth:
		return m.root.Material.WhiteCarvingDepth
	case qtui.ItemToolDiameter:
		return m.root.Carving.ToolDiameter
	case qtui.ItemToolStepOver:
		return m.root.Carving.StepOverPercent
	case qtui.ItemMaxStepDownSize:
		return m.root.Carving.MaxStepDownSize
	case qtui.ItemHorizontalFeedRate:
		return m.root.Carving.HorizontalFeedRate
	case qtui.ItemVerticalFeedRate:
		return m.root.Carving.VerticalFeedRate
	}

	log.Fatalf("Model: GetFloat32: Invalid tag = %s", tag)
	return 0
}

func (m *Model) GetChoice(tag string) int {
	switch tag {
	case qtui.ItemCarvingMode:
		return m.root.Carving.CarvingMode
	case qtui.ItemImageMode:
		return m.root.HeightMap.ImageMode
	case qtui.ItemToolType:
		return m.root.Carving.ToolType
	}

	log.Fatalf("Model: GetChoice: Invalid tag = %s", tag)
	return 0
}

func (m *Model) GetBool(tag string) bool {
	switch tag {
	case qtui.ItemMirrorImageX:
		return m.root.HeightMap.MirrorX
	case qtui.ItemMirrorImageY:
		return m.root.HeightMap.MirrorY
	}

	log.Fatalf("Model: GetBool: Invalid tag = %s", tag)
	return false
}

func (m *Model) GetHeightMap() *gui.QImage {
	return m.root.HeightMap.Image
}