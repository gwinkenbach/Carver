package model

import (
	"image"
	"time"

	"alvin.com/GoCarver/fui"
	"fyne.io/fyne/v2"
)

const (
	NumberRegex       = `[ ]*([0-9]*\.?[0-9]*)`
	SignedNumberRegex = `[ ]*(-?[0-9]*\.?[0-9]*)`

	PanelMaterialTag  = "material"
	PanelCarvingTag   = "carving"
	PanelHeightMapTag = "height_map"

	MatWidthTag       = "mat_width"
	MatHeightTag      = "mat_height"
	MatThicknessTag   = "mat_tick"
	CarvWidthTag      = "carv_width"
	CarvHeightTag     = "carv_height"
	CarvOffsetXTag    = "carv_offset_X"
	CarvOffsetYTag    = "carv_offset_Y"
	CarvBlackDepthTag = "carv_black_depth_tag"
	CarvWhiteDepthTag = "carv_white_depth_tag"

	ToolDiamTag      = "tool_diam"
	StepOverTag      = "step_over"
	ToolTypeTag      = "tool_type"
	MaxStepDownTag   = "max_step_down"
	HorizFeedRateTag = "horiz_feed_rate"
	VertFeedRateTag  = "vert_feed_rate"
	CarvDirectionTag = "carv_direction"

	ImgFillModeTag = "img_fill_mode"
	ImgMirrorXTag  = "img_mirror_x"
	ImgMirrorYTag  = "img_mirror_y"

	MenuNewModelTag    = "menu_new"
	MenuOpenModelTag   = "menu_open"
	MenuSaveModelTag   = "menu_save"
	MenuSaveModelAsTag = "menu_save_as"
	MenuOpenImageTag   = "menu_open_img"

	MenuGenGrblTag = "menu_gen_grbl"
)

// Choice strings.
var toolTypeChoices = []string{"Ball nose", "Straight"}
var carvingDirectionChoices = []string{"Along X", "Along Y", "First along X then along Y"}
var imageFillModeChoices = []string{"Fill", "Fit", "Crop"}

// Map image mode index from UI item to string mode used by Image Panel.
var imgModeIndexToStrMode = []string{fui.ImgModeFill, fui.ImgModeFit, fui.ImgModeCrop}

// A table with all the UI-item tags.
var allUIItemTags = [...]string{
	MatWidthTag, MatHeightTag, MatThicknessTag,
	CarvWidthTag, CarvHeightTag, CarvOffsetXTag, CarvOffsetYTag, CarvBlackDepthTag, CarvWhiteDepthTag,
	ToolDiamTag, StepOverTag, ToolTypeTag, MaxStepDownTag, HorizFeedRateTag, VertFeedRateTag,
	CarvDirectionTag,
	ImgFillModeTag, ImgMirrorXTag, ImgMirrorYTag}

type UIManager struct {
	uiRoot *fui.MainLayout
	menu   *fui.MainMenu

	onUIChangeListener     func(uiItemTag string)
	onMenuSelectedListener func(menuTag string)
	disableListeners       bool

	updateImageTimer *time.Timer
}

func NewUIManager() *UIManager {
	return &UIManager{}
}

func (ui *UIManager) BuildUI(w fyne.Window) {
	ui.uiRoot = fui.NewTopLayout()
	ui.buildControlPanel()
	w.SetContent(ui.uiRoot.GetRootContainer())

	ui.buildMainMenu(w)
}

func (ui *UIManager) FinalizeUI() {
	ui.uiRoot.GetControlPanel().SetChangeListener(func(tag string) {
		ui.onUIItemChange(tag)
	})
}

func (ui *UIManager) DisableListeners() {
	ui.disableListeners = true
}

func (ui *UIManager) EnableListeners() {
	ui.disableListeners = false
}

func (ui *UIManager) SetUIChangeListener(l func(string)) {
	ui.onUIChangeListener = l
}

func (ui *UIManager) SetMenuSelectedListener(l func(string)) {
	ui.onMenuSelectedListener = l
}

func (ui *UIManager) GetAllUIItemTags() []string {
	return allUIItemTags[:]
}

func (ui *UIManager) SetMenuItemEnabledState(menuTag string, enabled bool) {
	ui.menu.SetMenuItemEnabled(menuTag, enabled)
}

func (ui *UIManager) SetImage(img image.Image) {
	ui.uiRoot.GetImagePanel().SetImage(img)
	ui.delayedUpdateImagePanel()
}

func (ui *UIManager) SetUIItemFloatValue(tag string, val float32) {
	cp := ui.uiRoot.GetControlPanel()
	cp.SetWidgetFloatValue(tag, float64(val))
}

func (ui *UIManager) SetUIItemIntValue(tag string, val int) {
	cp := ui.uiRoot.GetControlPanel()
	cp.SetChoiceByIndex(tag, val)
}

func (ui *UIManager) SetUIItemBoolValue(tag string, val bool) {
	cp := ui.uiRoot.GetControlPanel()
	cp.SetCheckboxState(tag, val)
}

func (ui *UIManager) GetUIItemFloatValue(tag string) float32 {
	cp := ui.uiRoot.GetControlPanel()
	val, _ := cp.GetWidgetFloatValue(tag)
	return float32(val)
}

func (ui *UIManager) GetUIItemIntValue(tag string) int {
	cp := ui.uiRoot.GetControlPanel()
	val, _ := cp.GetWidgetIntValue(tag)
	return val
}

func (ui *UIManager) GetUIItemBoolValue(tag string) bool {
	cp := ui.uiRoot.GetControlPanel()
	val, _ := cp.GetCheckBoxState(tag)
	return val
}

func (ui *UIManager) buildControlPanel() {
	ui.buildMaterialPanel()
	ui.buildCarvingPanel()
	ui.buildHeightMapPanel()
	ui.uiRoot.GetControlPanel().Finalize()
}

func (ui *UIManager) buildMaterialPanel() {
	cp := ui.uiRoot.GetControlPanel()
	cp.AddGroup(PanelMaterialTag, "Material")

	cp.AddNumberEntry(PanelMaterialTag, MatWidthTag, "Material width (mm):", materialDimensionsConfig())
	cp.AddNumberEntry(PanelMaterialTag, MatHeightTag, "Material height (mm):", materialDimensionsConfig())
	cp.AddNumberEntry(PanelMaterialTag, MatThicknessTag, "Material thickness (mm):", materialThicknessConfig())
	cp.AddNumberEntry(PanelMaterialTag, CarvWidthTag, "Carving area width (mm):", materialDimensionsConfig())
	cp.AddNumberEntry(PanelMaterialTag, CarvHeightTag, "Carving area height (mm):", materialDimensionsConfig())
	cp.AddNumberEntry(PanelMaterialTag, CarvOffsetXTag, "Carving offset X (mm):", carvingOffsetConfig())
	cp.AddNumberEntry(PanelMaterialTag, CarvOffsetYTag, "Carving offset Y (mm):", carvingOffsetConfig())
	cp.AddNumberEntry(PanelMaterialTag, CarvBlackDepthTag, "Black carving depth (mm):", carvingDepthConfig())
	cp.AddNumberEntry(PanelMaterialTag, CarvWhiteDepthTag, "White carving depth (mm):", carvingDepthConfig())
}

func (ui *UIManager) buildCarvingPanel() {
	cp := ui.uiRoot.GetControlPanel()
	cp.AddGroup(PanelCarvingTag, "Carving")

	cp.AddNumberEntry(PanelCarvingTag, ToolDiamTag, "Tool diameter (mm):", toolDiameterConfig())
	cp.AddNumberEntry(PanelCarvingTag, StepOverTag, "Tool step over (%):", toolStepOverConfig())
	cp.AddSelector(PanelCarvingTag, ToolTypeTag, "Tool type:", toolTypeChoices)
	cp.AddNumberEntry(PanelCarvingTag, MaxStepDownTag, "Max step down (mm)):", stepDownSizeConfig())
	cp.AddNumberEntry(PanelCarvingTag, HorizFeedRateTag, "Horizontal feed rate (mm/min)):", feedRateConfig())
	cp.AddNumberEntry(PanelCarvingTag, VertFeedRateTag, "Vertical feed rate (mm/min)):", feedRateConfig())
	cp.AddSelector(PanelCarvingTag, CarvDirectionTag, "Carving mode:", carvingDirectionChoices)
}

func (ui *UIManager) buildHeightMapPanel() {
	cp := ui.uiRoot.GetControlPanel()
	cp.AddGroup(PanelHeightMapTag, "Height Map")

	cp.AddSelector(PanelHeightMapTag, ImgFillModeTag, "Image fill mode:", imageFillModeChoices)
	cp.AddCheckbox(PanelHeightMapTag, ImgMirrorXTag, "Image mirror-X:")
	cp.AddCheckbox(PanelHeightMapTag, ImgMirrorYTag, "Image mirror-Y:")
}

func (ui *UIManager) buildMainMenu(w fyne.Window) {
	ui.menu = fui.NewMainMenu(func(itemTag string) {
		if ui.onMenuSelectedListener != nil && !ui.disableListeners {
			ui.onMenuSelectedListener(itemTag)
		}
	})

	ui.menu.AddMenu("File")
	ui.menu.AddMenuItem("File", MenuNewModelTag, "New Model...", false)
	ui.menu.AddSeparator("File")
	ui.menu.AddMenuItem("File", MenuOpenModelTag, "Open Model...", false)
	ui.menu.AddMenuItem("File", MenuSaveModelTag, "Save Model...", true)
	ui.menu.AddMenuItem("File", MenuSaveModelAsTag, "Save Model As...", false)
	ui.menu.AddSeparator("File")
	ui.menu.AddMenuItem("File", MenuOpenImageTag, "Load Image...", false)

	ui.menu.AddMenu("Carve")
	ui.menu.AddMenuItem("Carve", MenuGenGrblTag, "Gen GRBL...", false)

	ui.menu.Realize(w)
}

func (ui *UIManager) onUIItemChange(uiItemTag string) {
	cp := ui.uiRoot.GetControlPanel()
	ip := ui.uiRoot.GetImagePanel()
	switch uiItemTag {
	case MatWidthTag, MatHeightTag:
		w, _ := cp.GetWidgetFloatValue(MatWidthTag)
		h, _ := cp.GetWidgetFloatValue(MatHeightTag)
		ip.SetMaterialDimensions(w, h)
		ui.delayedUpdateImagePanel()

	case CarvWidthTag, CarvHeightTag:
		w, _ := cp.GetWidgetFloatValue(CarvWidthTag)
		h, _ := cp.GetWidgetFloatValue(CarvHeightTag)
		ip.SetCarvingAreaDimensions(w, h)
		ui.delayedUpdateImagePanel()

	case CarvOffsetXTag, CarvOffsetYTag:
		dx, _ := cp.GetWidgetFloatValue(CarvOffsetXTag)
		dy, _ := cp.GetWidgetFloatValue(CarvOffsetYTag)
		ip.SetCarvingAreaOffsets(dx, dy)
		ui.delayedUpdateImagePanel()

	case ImgFillModeTag, ImgMirrorXTag, ImgMirrorYTag:
		mode, _ := cp.GetWidgetIntValue(ImgFillModeTag)
		mirrorX, _ := cp.GetCheckBoxState(ImgMirrorXTag)
		mirrorY, _ := cp.GetCheckBoxState(ImgMirrorYTag)
		ip.SetImageOptions(imgModeIndexToStrMode[mode], mirrorX, mirrorY)
		ui.delayedUpdateImagePanel()

	default:
	}

	if ui.onUIChangeListener != nil && !ui.disableListeners {
		ui.onUIChangeListener(uiItemTag)
	}
}

func (ui *UIManager) delayedUpdateImagePanel() {
	if ui.updateImageTimer == nil {
		ui.updateImageTimer = time.NewTimer(25 * time.Millisecond)
		go func() {
			for {
				<-ui.updateImageTimer.C
				ui.uiRoot.GetImagePanel().Refresh()
			}
		}()

	} else {
		ui.updateImageTimer.Stop()
		ui.updateImageTimer.Reset(25 * time.Millisecond)
	}
}

func materialDimensionsConfig() fui.NumericalEditConfigConfig {
	return fui.NumericalEditConfigConfig{
		MinVal: 10,
		MaxVal: 300,
		Format: "%.1f",
		Regex:  NumberRegex,
	}
}

func carvingOffsetConfig() fui.NumericalEditConfigConfig {
	return fui.NumericalEditConfigConfig{
		MinVal: 0,
		MaxVal: 99,
		Format: "%.1f",
		Regex:  NumberRegex,
	}
}

func carvingDepthConfig() fui.NumericalEditConfigConfig {
	return fui.NumericalEditConfigConfig{
		MinVal: -20,
		MaxVal: 5,
		Format: "%.1f",
		Regex:  SignedNumberRegex,
	}
}

func toolDiameterConfig() fui.NumericalEditConfigConfig {
	return fui.NumericalEditConfigConfig{
		MinVal: 1.0,
		MaxVal: 15.0,
		Format: "%.3f",
		Regex:  NumberRegex,
	}
}

func materialThicknessConfig() fui.NumericalEditConfigConfig {
	return fui.NumericalEditConfigConfig{
		MinVal: 5.0,
		MaxVal: 50.0,
		Format: "%.1f",
		Regex:  NumberRegex,
	}
}

func toolStepOverConfig() fui.NumericalEditConfigConfig {
	return fui.NumericalEditConfigConfig{
		MinVal: 1.0,
		MaxVal: 200.0,
		Format: "%.1f",
		Regex:  NumberRegex,
	}
}

func stepDownSizeConfig() fui.NumericalEditConfigConfig {
	return fui.NumericalEditConfigConfig{
		MinVal: 0.01,
		MaxVal: 9.0,
		Format: "%.2f",
		Regex:  NumberRegex,
	}
}

func feedRateConfig() fui.NumericalEditConfigConfig {
	return fui.NumericalEditConfigConfig{
		MinVal: 10,
		MaxVal: 2000.0,
		Format: "%.1f",
		Regex:  NumberRegex,
	}
}
