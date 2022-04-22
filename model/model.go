package model

import (
	"image"
	"log"

	"alvin.com/GoCarver/geom"
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

	EnableFinishPass           bool    `json:"enable_finish_pass"`
	FinishPassReductionPercent float32 `json:"finish_step_reduction_percent"`
	FinishMode                 int     `json:"finish_mode"`
	FinishHorizFeedRate        float32 `json:"finish_horiz__feed_rate"`
}

type heightMap struct {
	Image         image.Image `json:"-"` // Ignored in JSON
	ImageFileName string      `json:"imageFileName"`
	ImageMode     int         `json:"imageMode"`
	MirrorY       bool        `json:"mirrorY"`
	MirrorX       bool        `json:"mirrorX"`
}

type contourMachining struct {
	Enable             bool    `json:"enable_contour_machining"`
	ToolType           int     `json:"contour_tool_type"`
	ToolDiameter       float32 `json:"contour_tool_diameter"`
	MaxStepDownSize    float32 `json:"contour_max_step_down_size"`
	HorizontalFeedRate float32 `json:"contour_horizontal_feed_rate"`
	VerticalFeedRate   float32 `json:"contour_vertical_feed_rate"`
	CornerRadius       float32 `json:"contour_corner_radius"`
	NumTabsPerSize     int     `json:"contour_num_tabs_per_side"`
	TabWidth           float32 `json:"contour_tab_width"`
	TabHeight          float32 `json:"contour_tab_height"`
}

type modelRoot struct {
	Material  material         `json:"material"`
	Carving   carving          `json:"carving"`
	HeightMap heightMap        `json:"height_map"`
	Contour   contourMachining `json:"contour_machining"`
}

const (
	ToolTypeBallNose = 0
	ToolTypeStraight = 1

	CarvingModeAlongX      = 0
	CarvingModeAlongY      = 1
	CarvingModeAlongXThenY = 2

	FinishModeFirstDirectionOnly = 0
	FinishModeLastDirectionOnly  = 1
	FinishModeInAllDirections    = 2

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
				ToolDiameter:               3.175, // millimeters
				StepOverPercent:            40,    // Percent of tool diameter
				MaxStepDownSize:            0.5,
				HorizontalFeedRate:         500.0, // millimeters per minute
				VerticalFeedRate:           300.0, // millimeters per minutes
				CarvingMode:                CarvingModeAlongX,
				EnableFinishPass:           false,
				FinishPassReductionPercent: 50.0,
				FinishMode:                 FinishModeFirstDirectionOnly,
				FinishHorizFeedRate:        750.0, // millimeters per minute,
			},

			Contour: contourMachining{
				ToolDiameter:       3.175, // millimeters
				MaxStepDownSize:    0.5,
				HorizontalFeedRate: 500.0, // millimeters per minute
				VerticalFeedRate:   300.0, // millimeters per minutes
				CornerRadius:       5.0,   // millimeters
				NumTabsPerSize:     2,
				TabWidth:           4.0, // millimeters
				TabHeight:          0.5, // millimeters
			},
		},
	}
}

func (m *Model) SetDirty(dirty bool) {
	m.dirty = dirty
}

func (m *Model) GetFloat32Value(tag string) float32 {
	switch tag {
	case MatWidthTag:
		return m.root.Material.MaterialWidth
	case MatHeightTag:
		return m.root.Material.MaterialHeight
	case MatThicknessTag:
		return m.root.Material.MaterialThickness
	case CarvWidthTag:
		return m.root.Material.CarvingAreaWidth
	case CarvHeightTag:
		return m.root.Material.CarvingAreaHeight
	case CarvOffsetXTag:
		return m.root.Material.CarvingAreaOffsetX
	case CarvOffsetYTag:
		return m.root.Material.CarvingAreaOffsetY
	case CarvBlackDepthTag:
		return m.root.Material.BlackCarvingDepth
	case CarvWhiteDepthTag:
		return m.root.Material.WhiteCarvingDepth
	case ToolDiamTag:
		return m.root.Carving.ToolDiameter
	case StepOverTag:
		return m.root.Carving.StepOverPercent
	case MaxStepDownTag:
		return m.root.Carving.MaxStepDownSize
	case HorizFeedRateTag:
		return m.root.Carving.HorizontalFeedRate
	case VertFeedRateTag:
		return m.root.Carving.VerticalFeedRate
	case FinishPassReductionTag:
		return m.root.Carving.FinishPassReductionPercent
	case FinishPassHorizFeedRateTag:
		return m.root.Carving.FinishHorizFeedRate
	case ContourCornerRadiusTag:
		return m.root.Contour.CornerRadius
	case ContourHorizFeedRateTag:
		return m.root.Contour.HorizontalFeedRate
	case ContourVertFeedRateTag:
		return m.root.Contour.VerticalFeedRate
	case ContourToolDiameterTag:
		return m.root.Contour.ToolDiameter
	case ContourTabWidthTag:
		return m.root.Contour.TabWidth
	case ContourTabHeightTag:
		return m.root.Contour.TabHeight
	case ContourMaxStepDownTag:
		return m.root.Contour.MaxStepDownSize
	}

	log.Fatalf("Model: GetFloat32: Invalid tag = %s", tag)
	return 0
}

func (m *Model) GetIntValue(tag string) int {
	switch tag {
	case CarvDirectionTag:
		return m.root.Carving.CarvingMode
	case ImgFillModeTag:
		return m.root.HeightMap.ImageMode
	case ToolTypeTag:
		return m.root.Carving.ToolType
	case FinishPassModeTag:
		return m.root.Carving.FinishMode
	case ContourToolTypeTag:
		return m.root.Contour.ToolType
	case ContourNubTabsPerSideTag:
		return m.root.Contour.NumTabsPerSize
	}

	log.Fatalf("Model: GetChoice: Invalid tag = %s", tag)
	return 0
}

func (m *Model) GetBoolValue(tag string) bool {
	switch tag {
	case ImgMirrorXTag:
		return m.root.HeightMap.MirrorX
	case ImgMirrorYTag:
		return m.root.HeightMap.MirrorY
	case UseFinishPassTag:
		return m.root.Carving.EnableFinishPass
	case EnableContourTag:
		return m.root.Contour.Enable
	}

	log.Fatalf("Model: GetBool: Invalid tag = %s", tag)
	return false
}

func (m *Model) GetHeightMap() image.Image {
	return m.root.HeightMap.Image
}

func (m *Model) SetFloat32Value(tag string, val float32) {
	switch tag {
	case MatWidthTag:
		m.root.Material.MaterialWidth = val
	case MatHeightTag:
		m.root.Material.MaterialHeight = val
	case MatThicknessTag:
		m.root.Material.MaterialThickness = val
	case CarvWidthTag:
		m.root.Material.CarvingAreaWidth = val
	case CarvHeightTag:
		m.root.Material.CarvingAreaHeight = val
	case CarvOffsetXTag:
		m.root.Material.CarvingAreaOffsetX = val
	case CarvOffsetYTag:
		m.root.Material.CarvingAreaOffsetY = val
	case CarvBlackDepthTag:
		m.root.Material.BlackCarvingDepth = val
	case CarvWhiteDepthTag:
		m.root.Material.WhiteCarvingDepth = val
	case ToolDiamTag:
		m.root.Carving.ToolDiameter = val
	case StepOverTag:
		m.root.Carving.StepOverPercent = val
	case MaxStepDownTag:
		m.root.Carving.MaxStepDownSize = val
	case HorizFeedRateTag:
		m.root.Carving.HorizontalFeedRate = val
	case VertFeedRateTag:
		m.root.Carving.VerticalFeedRate = val
	case FinishPassReductionTag:
		m.root.Carving.FinishPassReductionPercent = val
	case FinishPassHorizFeedRateTag:
		m.root.Carving.FinishHorizFeedRate = val
	case ContourCornerRadiusTag:
		m.root.Contour.CornerRadius = val
	case ContourHorizFeedRateTag:
		m.root.Contour.HorizontalFeedRate = val
	case ContourVertFeedRateTag:
		m.root.Contour.VerticalFeedRate = val
	case ContourToolDiameterTag:
		m.root.Contour.ToolDiameter = val
	case ContourTabWidthTag:
		m.root.Contour.TabWidth = val
	case ContourTabHeightTag:
		m.root.Contour.TabHeight = val
	case ContourMaxStepDownTag:
		m.root.Contour.MaxStepDownSize = val
	default:
		log.Fatalf("Model: SetFloat32: Invalid tag = %s", tag)
	}
}

func (m *Model) SetIntValue(tag string, val int) {
	switch tag {
	case CarvDirectionTag:
		m.root.Carving.CarvingMode = val
	case ImgFillModeTag:
		m.root.HeightMap.ImageMode = val
	case ToolTypeTag:
		m.root.Carving.ToolType = val
	case FinishPassModeTag:
		m.root.Carving.FinishMode = val
	case ContourToolTypeTag:
		m.root.Contour.ToolType = val
	case ContourNubTabsPerSideTag:
		m.root.Contour.NumTabsPerSize = val
	default:
		log.Fatalf("Model: SetChoice: Invalid tag = %s", tag)
	}
}

func (m *Model) SetBoolValue(tag string, val bool) {
	switch tag {
	case ImgMirrorXTag:
		m.root.HeightMap.MirrorX = val
	case ImgMirrorYTag:
		m.root.HeightMap.MirrorY = val
	case UseFinishPassTag:
		m.root.Carving.EnableFinishPass = val
	case EnableContourTag:
		m.root.Contour.Enable = val
	default:
		log.Fatalf("Model: SetBool: Invalid tag = %s", tag)
	}
}

func GetModelValueByTag[T any](m *Model, tag string) T {
	var ret T
	switch p := any(&ret).(type) {
	case *int:
		*p = m.GetIntValue(tag)
	case *float32:
		*p = m.GetFloat32Value(tag)
	case *float64:
		*p = float64(m.GetFloat32Value(tag))
	case *bool:
		*p = m.GetBoolValue(tag)
	default:
		log.Fatalf("GetModelValueByTag: Unsupported type: %T\n", ret)
	}

	return ret
}

func SetModelValueByTag[T any](m *Model, tag string, val T) {
	switch p := any(&val).(type) {
	case *int:
		m.SetIntValue(tag, *p)
	case *float32:
		m.SetFloat32Value(tag, *p)
	case *float64:
		m.SetFloat32Value(tag, float32(*p))
	case *bool:
		m.SetBoolValue(tag, *p)
	default:
		log.Fatalf("SetModelValueByTag: Unsupported type: %T\n", val)
	}
}
