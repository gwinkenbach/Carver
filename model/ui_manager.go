package model

import (
	"image"
	"log"
	"time"

	"alvin.com/GoCarver/fui"
	"fyne.io/fyne/v2"
)

const (
	NumberRegex       = `[ ]*([0-9]*\.?[0-9]*)`
	SignedNumberRegex = `[ ]*(-?[0-9]*\.?[0-9]*)`

	PanelMaterialTag      = "material_panel"
	PanelCarvingTag       = "carving_panel"
	PanelHeightMapTag     = "height_map_panel"
	PanelContourMachining = "contour_panel"

	MatWidthTag       = "mat_width"
	MatHeightTag      = "mat_height"
	MatThicknessTag   = "mat_tick"
	CarvWidthTag      = "carv_width"
	CarvHeightTag     = "carv_height"
	CarvOffsetXTag    = "carv_offset_X"
	CarvOffsetYTag    = "carv_offset_Y"
	CarvBlackDepthTag = "carv_black_depth_tag"
	CarvWhiteDepthTag = "carv_white_depth_tag"

	ToolDiamTag                = "tool_diam"
	StepOverTag                = "step_over"
	ToolTypeTag                = "tool_type"
	MaxStepDownTag             = "max_step_down"
	HorizFeedRateTag           = "horiz_feed_rate"
	VertFeedRateTag            = "vert_feed_rate"
	CarvDirectionTag           = "carv_direction"
	UseFinishPassTag           = "use_finishing_pass"
	FinishPassReductionTag     = "finish_pass_reduc"
	FinishPassModeTag          = "finish_pass_mode"
	FinishPassHorizFeedRateTag = "finish_pass_horiz_feed"

	EnableContourTag         = "enable_contour_machining"
	ContourToolTypeTag       = "contour_tool_type"
	ContourToolDiameterTag   = "contour_tool_diameter"
	ContourMaxStepDownTag    = "contour_max_step_down_size"
	ContourHorizFeedRateTag  = "contour_horizontal_feed_rate"
	ContourVertFeedRateTag   = "contour_vertical_feed_rate"
	ContourCornerRadiusTag   = "contour_corner_radius"
	ContourNubTabsPerSideTag = "contour_num_tabs_per_side"
	ContourTabWidthTag       = "contour_tab_width"
	ContourTabHeightTag      = "contour_tab_height"

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
var finishPassModeChoices = []string{
	"First direction only", "Last direction only", "All directions"}
var imageFillModeChoices = []string{"Fill", "Fit", "Crop"}
var numTabPerSideChoices = []string{"no tab", "1 tab", "2 tabs", "3 tabs", "4 tabs"}

// Map image mode index from UI item to string mode used by Image Panel.
var imgModeIndexToStrMode = []string{fui.ImgModeFill, fui.ImgModeFit, fui.ImgModeCrop}

type UIManager struct {
	uiRoot *fui.MainLayout
	menu   *fui.MainMenu

	allUIItemTags      []string
	numEntryUIItemTags []string
	selectorUIItemTags []string
	checkboxUIItemTags []string

	onUIChangeListener     func(uiItemTag string)
	onMenuSelectedListener func(menuTag string)
	disableListeners       bool

	updateImageTimer *time.Timer
}

func NewUIManager() *UIManager {
	return &UIManager{
		allUIItemTags:      make([]string, 0),
		numEntryUIItemTags: make([]string, 0),
		selectorUIItemTags: make([]string, 0),
		checkboxUIItemTags: make([]string, 0),
	}
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
	return ui.allUIItemTags
}

func (ui *UIManager) IsNumEntryUIItem(itemTag string) bool {
	for _, tag := range ui.numEntryUIItemTags {
		if tag == itemTag {
			return true
		}
	}
	return false
}

func (ui *UIManager) IsSelectorUIItem(itemTag string) bool {
	for _, tag := range ui.selectorUIItemTags {
		if tag == itemTag {
			return true
		}
	}
	return false
}

func (ui *UIManager) IsCheckboxUIItem(itemTag string) bool {
	for _, tag := range ui.checkboxUIItemTags {
		if tag == itemTag {
			return true
		}
	}
	return false
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
	ui.buildContourMachiningPanel()
	ui.uiRoot.GetControlPanel().Finalize()
}

func (ui *UIManager) buildMaterialPanel() {
	cp := ui.uiRoot.GetControlPanel()
	cp.AddGroup(PanelMaterialTag, "Material")

	cp.AddSeparator(PanelMaterialTag, "Material area:", true)
	ui.addNumberEntry(PanelMaterialTag, MatWidthTag, "Material width (mm):", materialDimensionsConfig())
	ui.addNumberEntry(PanelMaterialTag, MatHeightTag, "Material height (mm):", materialDimensionsConfig())
	ui.addNumberEntry(PanelMaterialTag, MatThicknessTag, "Material thickness (mm):", materialThicknessConfig())
	cp.AddSeparator(PanelMaterialTag, "Carving area:", true)
	ui.addNumberEntry(PanelMaterialTag, CarvWidthTag, "Carving area width (mm):", materialDimensionsConfig())
	ui.addNumberEntry(PanelMaterialTag, CarvHeightTag, "Carving area height (mm):", materialDimensionsConfig())
	ui.addNumberEntry(PanelMaterialTag, CarvOffsetXTag, "Carving offset X (mm):", carvingOffsetConfig())
	ui.addNumberEntry(PanelMaterialTag, CarvOffsetYTag, "Carving offset Y (mm):", carvingOffsetConfig())
	ui.addNumberEntry(PanelMaterialTag, CarvBlackDepthTag, "Black carving depth (mm):", carvingDepthConfig())
	ui.addNumberEntry(PanelMaterialTag, CarvWhiteDepthTag, "White carving depth (mm):", carvingDepthConfig())
}

func (ui *UIManager) buildCarvingPanel() {
	cp := ui.uiRoot.GetControlPanel()
	cp.AddGroup(PanelCarvingTag, "Carving")

	cp.AddSeparator(PanelCarvingTag, "Tool:", true)
	ui.addNumberEntry(PanelCarvingTag, ToolDiamTag, "Tool diameter (mm):", toolDiameterConfig())
	ui.addNumberEntry(PanelCarvingTag, StepOverTag, "Tool step over (%):", toolStepOverConfig())
	ui.addSelector(PanelCarvingTag, ToolTypeTag, "Tool type:", toolTypeChoices)
	cp.AddSeparator(PanelCarvingTag, "Carving:", true)
	ui.addNumberEntry(PanelCarvingTag, MaxStepDownTag, "Max step down (mm)):", stepDownSizeConfig())
	ui.addNumberEntry(PanelCarvingTag, HorizFeedRateTag, "Horizontal feed rate (mm/min)):", feedRateConfig())
	ui.addNumberEntry(PanelCarvingTag, VertFeedRateTag, "Vertical feed rate (mm/min)):", feedRateConfig())
	ui.addSelector(PanelCarvingTag, CarvDirectionTag, "Carving mode:", carvingDirectionChoices)
	cp.AddSeparator(PanelCarvingTag, "Optional finish pass:", true)
	ui.addCheckbox(PanelCarvingTag, UseFinishPassTag, "Enable finishing pass:")
	ui.addNumberEntry(PanelCarvingTag, FinishPassReductionTag, "Finishing step reduction (%):", finishingPassConfig())
	ui.addSelector(PanelCarvingTag, FinishPassModeTag, "Finish mode:", finishPassModeChoices)
	ui.addNumberEntry(PanelCarvingTag, FinishPassHorizFeedRateTag, "Finish pass horiz feed rate (mm/min)):", feedRateConfig())
}

func (ui *UIManager) buildHeightMapPanel() {
	cp := ui.uiRoot.GetControlPanel()
	cp.AddGroup(PanelHeightMapTag, "Height Map")

	ui.addSelector(PanelHeightMapTag, ImgFillModeTag, "Image fill mode:", imageFillModeChoices)
	ui.addCheckbox(PanelHeightMapTag, ImgMirrorXTag, "Image mirror-X:")
	ui.addCheckbox(PanelHeightMapTag, ImgMirrorYTag, "Image mirror-Y:")
}

func (ui *UIManager) buildContourMachiningPanel() {
	cp := ui.uiRoot.GetControlPanel()
	cp.AddGroup(PanelContourMachining, "Contour Machining")

	ui.addCheckbox(PanelContourMachining, EnableContourTag, "Enable contour maching:")
	ui.addSelector(PanelContourMachining, ContourToolTypeTag, "Tool type:", toolTypeChoices)
	ui.addNumberEntry(PanelContourMachining, ContourToolDiameterTag, "Tool diameter (mm):", toolDiameterConfig())
	ui.addNumberEntry(PanelContourMachining, ContourMaxStepDownTag, "Max step down (mm)):", stepDownSizeConfig())
	ui.addNumberEntry(PanelContourMachining, ContourHorizFeedRateTag, "Horizontal feed rate (mm/min)):", feedRateConfig())
	ui.addNumberEntry(PanelContourMachining, ContourVertFeedRateTag, "Vertical feed rate (mm/min)):", feedRateConfig())
	ui.addNumberEntry(PanelContourMachining, ContourCornerRadiusTag, "Corner radius (mm)):", cornerRadiusConfig())
	ui.addSelector(PanelContourMachining, ContourNubTabsPerSideTag, "Number of tabs on each side:", numTabPerSideChoices)
	ui.addNumberEntry(PanelContourMachining, ContourTabWidthTag, "Width of tabs (mm)):", tabWidthConfig())
	ui.addNumberEntry(PanelContourMachining, ContourTabHeightTag, "Height of tabs (mm)):", tabHeightConfig())
}

func (ui *UIManager) addNumberEntry(
	panel string, uiItemTag string, label string, config fui.NumericalEditConfigConfig) {

	cp := ui.uiRoot.GetControlPanel()
	cp.AddNumberEntry(panel, uiItemTag, label, config)
	ui.allUIItemTags = append(ui.allUIItemTags, uiItemTag)
	ui.numEntryUIItemTags = append(ui.numEntryUIItemTags, uiItemTag)
}

func (ui *UIManager) addSelector(panel string, uiItemTag string, label string, choices []string) {
	cp := ui.uiRoot.GetControlPanel()
	cp.AddSelector(panel, uiItemTag, label, choices)
	ui.allUIItemTags = append(ui.allUIItemTags, uiItemTag)
	ui.selectorUIItemTags = append(ui.selectorUIItemTags, uiItemTag)
}

func (ui *UIManager) addCheckbox(panel string, uiItemTag string, label string) {
	cp := ui.uiRoot.GetControlPanel()
	cp.AddCheckbox(panel, uiItemTag, label)
	ui.allUIItemTags = append(ui.allUIItemTags, uiItemTag)
	ui.checkboxUIItemTags = append(ui.checkboxUIItemTags, uiItemTag)
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

func GetUiValueByTag[T any](ui *UIManager, tag string) T {
	var ret T
	switch p := any(&ret).(type) {
	case *int:
		*p = ui.GetUIItemIntValue(tag)
	case *float32:
		*p = ui.GetUIItemFloatValue(tag)
	case *float64:
		*p = float64(ui.GetUIItemFloatValue(tag))
	case *bool:
		*p = ui.GetUIItemBoolValue(tag)
	default:
		log.Fatalf("GetUiValueByTag: Unsupported type: %T\n", ret)
	}

	return ret
}

func SetUiValueByTag[T any](ui *UIManager, tag string, val T) {
	switch p := any(&val).(type) {
	case *int:
		ui.SetUIItemIntValue(tag, *p)
	case *float32:
		ui.SetUIItemFloatValue(tag, *p)
	case *float64:
		ui.SetUIItemFloatValue(tag, float32(*p))
	case *bool:
		ui.SetUIItemBoolValue(tag, *p)
	default:
		log.Fatalf("SetUiValueByTag: Unsupported type: %T\n", val)
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

func finishingPassConfig() fui.NumericalEditConfigConfig {
	return fui.NumericalEditConfigConfig{
		MinVal: 1.0,
		MaxVal: 90.0,
		Format: "%.1f",
		Regex:  NumberRegex,
	}
}

func cornerRadiusConfig() fui.NumericalEditConfigConfig {
	return fui.NumericalEditConfigConfig{
		MinVal: 0.0,
		MaxVal: 20.0,
		Format: "%.1f",
		Regex:  NumberRegex,
	}
}

func tabWidthConfig() fui.NumericalEditConfigConfig {
	return fui.NumericalEditConfigConfig{
		MinVal: 2.0,
		MaxVal: 10.0,
		Format: "%.1f",
		Regex:  NumberRegex,
	}
}

func tabHeightConfig() fui.NumericalEditConfigConfig {
	return fui.NumericalEditConfigConfig{
		MinVal: 0.2,
		MaxVal: 2.0,
		Format: "%.1f",
		Regex:  NumberRegex,
	}
}
