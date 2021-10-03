package qtui

import (
	"log"

	"github.com/therecipe/qt/widgets"
)

// These strings are both label in the UI and tags used to uniquely identify
// model/ui items.
const (
	ItemMaterialWidth      = "Material width:"
	ItemMaterialHeight     = "Material Height:"
	ItemMaterialThickness  = "Material thickness"
	ItemCarvingAreaWidth   = "Carving area width:"
	ItemCarvingAreaHeight  = "Carving area height:"
	ItemCarvingAreaOffsetX = "Carving area offset X:"
	ItemCarvingAreaOffsetY = "Carving area offset Y:"
	ItemBlackCarvingDepth  = "Black carving depth:"
	ItemWhiteCarvingDepth  = "White carving depth:"
	ItemToolDiameter       = "Tool diameter:"
	ItemToolType           = "Tool type:"
	ItemToolStepOver       = "Tool step-over:"
	ItemMaxStepDownSize    = "Max step-down size:"
	ItemHorizontalFeedRate = "Horizontal feed rate:"
	ItemVerticalFeedRate   = "Vertical feedrate:"
	ItemCarvingMode        = "Carving mode:"
	ItemImageMode          = "Image fill mode:"
	ItemMirrorImageX       = "Image mirror-X:"
	ItemMirrorImageY       = "Image mirror-Y:"
)

// AllUIItemTags is a table with all the valid UI item tags.
var AllUIItemTags = [...]string{ItemMaterialWidth, ItemMaterialHeight, ItemMaterialThickness,
	ItemToolDiameter, ItemToolStepOver, ItemMaxStepDownSize, ItemCarvingAreaWidth,
	ItemCarvingAreaHeight, ItemCarvingAreaOffsetX, ItemCarvingAreaOffsetY, ItemBlackCarvingDepth,
	ItemWhiteCarvingDepth, ItemHorizontalFeedRate, ItemVerticalFeedRate, ItemCarvingMode,
	ItemImageMode, ItemMirrorImageX, ItemMirrorImageY, ItemToolType}

// Choice strings.
var toolTypeChoices = []string{"Ball nose", "Straight"}
var carvingDirectionChoices = []string{"Along X", "Along Y", "First along X then along Y"}
var imageFillModeChoices = []string{"Fill", "Fit", "Crop"}

// Function type, invoked when a UI item changes.
type onItemChangedFunc func(um *UIManager, tag string)

// Function type, provides a formatter/validator pair that can be associated with a UI item.
type provideUIConfig func() (f formatter, v validator)

// UIManager manages the UI.
type UIManager struct {
	liveDataObserver

	// The main panel layout.
	mainLayout *MainLayout
	// All items identified by Item tags. Each UI is associated with a liveData instance,
	// which is used to mitigate data interchange between the UI widget and the data value.
	uiItems map[string]*liveData

	onItemChanged onItemChangedFunc // If set, this function is called when a UI item changes.
}

// NewUIManager creates and returns a new manager.
func NewUIManager() *UIManager {
	return &UIManager{
		uiItems: make(map[string]*liveData),
	}
}

// BuildUI must be called once to build the entire UI.
func (um *UIManager) BuildUI() {
	um.buildLayout()
	um.populateControlPanel()
}

// GetRootPanel returns the main layout's root canvas object.
func (um *UIManager) GetRootPanel() *widgets.QWidget {
	return um.mainLayout.GetRoot()
}

// GetMainLayout return the main layout object.
func (um *UIManager) GetMainLayout() *MainLayout {
	return um.mainLayout
}

// SetOnItemChanged sets the onItemChanged, which is called each time a UI item changes.
func (um *UIManager) SetOnItemChanged(oic onItemChangedFunc) {
	um.onItemChanged = oic
}

func (um *UIManager) buildLayout() {
	um.mainLayout = NewMainLayout()
}

func (um *UIManager) populateControlPanel() {
	cp := um.mainLayout.GetControlPanel()

	um.addEntryWidget(ItemMaterialWidth, cp, materialDimensionsConfig)
	um.addEntryWidget(ItemMaterialHeight, cp, materialDimensionsConfig)
	um.addEntryWidget(ItemMaterialThickness, cp, materialThicknessConfig)
	um.addEntryWidget(ItemCarvingAreaWidth, cp, materialDimensionsConfig)
	um.addEntryWidget(ItemCarvingAreaHeight, cp, materialDimensionsConfig)
	um.addEntryWidget(ItemCarvingAreaOffsetX, cp, carvingOffsetConfig)
	um.addEntryWidget(ItemCarvingAreaOffsetY, cp, carvingOffsetConfig)
	um.addEntryWidget(ItemBlackCarvingDepth, cp, carvingDepthConfig)
	um.addEntryWidget(ItemWhiteCarvingDepth, cp, carvingDepthConfig)

	cp.AddVerticalSpace(10)

	um.addEntryWidget(ItemToolDiameter, cp, toolDiameterConfig)
	um.addSelectorWidget(ItemToolType, cp, toolTypeChoices)
	um.addEntryWidget(ItemToolStepOver, cp, toolStepOverConfig)
	um.addEntryWidget(ItemMaxStepDownSize, cp, stepDownSizeConfig)
	um.addEntryWidget(ItemHorizontalFeedRate, cp, feedRateConfig)
	um.addEntryWidget(ItemVerticalFeedRate, cp, feedRateConfig)
	um.addSelectorWidget(ItemCarvingMode, cp, carvingDirectionChoices)

	cp.AddVerticalSpace(10)

	um.addSelectorWidget(ItemImageMode, cp, imageFillModeChoices)
	um.addCheckBoxWidget(ItemMirrorImageX, cp)
	um.addCheckBoxWidget(ItemMirrorImageY, cp)
}

func (um *UIManager) addEntryWidget(label string, cp *ControlPanel, cfg provideUIConfig) {
	f, v := cfg()
	e := widgets.NewQLineEdit(nil)

	cp.AddLabeledWidget(label, e)

	ld := NewLiveData(label, f, v, NewFloat32LineEditDataSource(e))
	ld.SetObserver(um)
	um.uiItems[label] = ld
}

func (um *UIManager) addSelectorWidget(label string, cp *ControlPanel, options []string) {
	sel := widgets.NewQComboBox(nil)
	sel.AddItems(options)

	f := NewSelectFormatter(options)
	v := NewSelectValidator(options)
	ld := NewLiveData(label, f, v, NewSelectDataSource(sel))
	ld.SetObserver(um)
	um.uiItems[label] = ld

	cp.AddLabeledWidget(label, sel)
}

func (um *UIManager) addCheckBoxWidget(label string, cp *ControlPanel) {
	cb := widgets.NewQCheckBox(nil)

	f := NewBoolFormatter()
	v := NewBoolValidator()
	ld := NewLiveData(label, f, v, NewBoolDataSource(cb))
	ld.SetObserver(um)
	um.uiItems[label] = ld

	cp.AddLabeledWidget(label, cb)
}

// SetValueFloat32 sets a UI item identified by its Item tag using a float data value.
func (um *UIManager) SetValueFloat32(label string, val float32) {
	ld := um.uiItems[label]
	if ld == nil {
		log.Fatalf("No ui item with label \"%s\"", label)
	}

	ld.updateValue(val)
	um.refreshRelatedUIComponents(label)
}

// GetValueFloat32 reads a float value from a UI item identified by its Item tag.
func (um *UIManager) GetValueFloat32(label string) float32 {
	ld := um.uiItems[label]
	if ld == nil {
		log.Fatalf("No ui item with label \"%s\"", label)
	}

	val, ok := ld.getValue().(float32)
	if !ok {
		log.Fatalf("UI item \"%s\" does not produce a float32", label)
	}

	return val
}

// SetChoice sets a choice in a widget, such as a selector, by the choice index.
func (um *UIManager) SetChoice(label string, val int) {
	ld := um.uiItems[label]
	if ld == nil {
		log.Fatalf("No ui item with label \"%s\"", label)
	}

	ld.updateValue(val)
	um.refreshRelatedUIComponents(label)
}

// GetChoice reads the choice from a widget, such as a selector widget, as an index.
func (um *UIManager) GetChoice(label string) int {
	ld := um.uiItems[label]
	if ld == nil {
		log.Fatalf("No ui item with label \"%s\"", label)
	}

	val, ok := ld.getValue().(int)
	if !ok {
		log.Fatalf("UI item \"%s\" does not produce a int", label)
	}

	return val
}

// SetBool sets a state in a widget, such as a checkbox.
func (um *UIManager) SetBool(label string, val bool) {
	ld := um.uiItems[label]
	if ld == nil {
		log.Fatalf("No ui item with label \"%s\"", label)
	}

	ld.updateValue(val)
	um.refreshRelatedUIComponents(label)
}

// GetBool reads the state from a widget, such as a checkbox.
func (um *UIManager) GetBool(label string) bool {
	ld := um.uiItems[label]
	if ld == nil {
		log.Fatalf("No ui item with label \"%s\"", label)
	}

	val, ok := ld.getValue().(bool)
	if !ok {
		log.Fatalf("UI item \"%s\" does not produce a bool", label)
	}

	return val
}

func (um *UIManager) onMaterialDimensionsChanged() {
	w := um.GetValueFloat32(ItemMaterialWidth)
	h := um.GetValueFloat32(ItemMaterialHeight)
	um.mainLayout.GetImagePanel().UpdateMaterialSize(w, h)
}

func (um *UIManager) onCarvingAreaChanged() {
	w := um.GetValueFloat32(ItemCarvingAreaWidth)
	h := um.GetValueFloat32(ItemCarvingAreaHeight)
	x := um.GetValueFloat32(ItemCarvingAreaOffsetX)
	y := um.GetValueFloat32(ItemCarvingAreaOffsetY)
	um.mainLayout.GetImagePanel().UpdateCarvingArea(w, h, x, y)
}

func (um *UIManager) onImageMappingChanged() {
	mode := um.GetChoice(ItemImageMode)
	mirrorX := um.GetBool(ItemMirrorImageX)
	mirrorY := um.GetBool(ItemMirrorImageY)
	um.mainLayout.GetImagePanel().UpdateImageParameters(mode, mirrorX, mirrorY)
}

func materialDimensionsConfig() (f formatter, v validator) {
	f = NewFloat32Formatter("%4.0f mm")
	v = NewFloat32Validator(`[ ]*([0-9]*\.?[0-9]*)`, 10.0, 300.0)
	return
}

func materialThicknessConfig() (f formatter, v validator) {
	f = NewFloat32Formatter("%5.1f mm")
	v = NewFloat32Validator(`[ ]*([0-9]*\.?[0-9]*)`, 5.0, 50.0)
	return
}

func carvingOffsetConfig() (f formatter, v validator) {
	f = NewFloat32Formatter("%4.1f mm")
	v = NewFloat32Validator(`[ ]*([0-9]*\.?[0-9]*)`, 0.0, 99.0)
	return
}

func carvingDepthConfig() (f formatter, v validator) {
	f = NewFloat32Formatter("%5.1f mm")
	v = NewFloat32Validator(`[ ]*(-?[0-9]*\.?[0-9]*)`, -20.0, 5.0)
	return
}

func toolDiameterConfig() (f formatter, v validator) {
	f = NewFloat32Formatter("%6.3f mm")
	v = NewFloat32Validator(`[ ]*([0-9]*\.?[0-9]*)`, 1.0, 15.0)
	return
}

func toolStepOverConfig() (f formatter, v validator) {
	f = NewFloat32Formatter("%5.1f %%")
	v = NewFloat32Validator(`[ ]*([0-9]*\.?[0-9]*)`, 1.0, 200.0)
	return
}

func stepDownSizeConfig() (f formatter, v validator) {
	f = NewFloat32Formatter("%4.2f mm")
	v = NewFloat32Validator(`[ ]*([0-9]*\.?[0-9]*)`, 0.01, 9.0)
	return
}

func feedRateConfig() (f formatter, v validator) {
	f = NewFloat32Formatter("%6.1f mm/min")
	v = NewFloat32Validator(`[ ]*([0-9]*\.?[0-9]*)`, 10, 2000.0)
	return
}

// Implement interface liveDataObserver.
func (um *UIManager) notifyDataChanged(tag string) {
	if um.onItemChanged != nil {
		um.onItemChanged(um, tag)
	}

	um.refreshRelatedUIComponents(tag)
}

func (um *UIManager) refreshRelatedUIComponents(tag string) {
	switch tag {
	case ItemMaterialHeight, ItemMaterialWidth:
		um.onMaterialDimensionsChanged()
	case ItemCarvingAreaHeight, ItemCarvingAreaWidth, ItemCarvingAreaOffsetX, ItemCarvingAreaOffsetY:
		um.onCarvingAreaChanged()
	case ItemImageMode, ItemMirrorImageX, ItemMirrorImageY:
		um.onImageMappingChanged()
	default:
	}
}
