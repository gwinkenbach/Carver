package fui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// A value that is two-way bound to a UI widget.
type uiItemValueBinding struct {
	boolVal  binding.Bool
	floatVal binding.Float
	intVal   binding.Int

	widget interface{}
}

// NumericalEditConfigConfig encapsulates the config parameters for a numerical entry widget.
type NumericalEditConfigConfig struct {
	MinVal, MaxVal float64 // Valid range. If minVal >= maxVal, the range is unrestricted.
	Format         string  // Format string for float-to-string conversion, or "".
	Regex          string  // Regex for validation, or "".
}

var _ listenForChangeWithTag = (*ControlPanel)(nil)

type ControlPanel struct {
	root *container.AppTabs

	groups  map[string]*container.TabItem
	uiItems map[string]*uiItemValueBinding

	enableChangeNotifications bool
	changeListener            func(uiItemTag string)
}

func NewControlPanel() *ControlPanel {
	cp := &ControlPanel{}
	cp.root = container.NewAppTabs()
	cp.groups = make(map[string]*container.TabItem)
	cp.uiItems = make(map[string]*uiItemValueBinding)

	return cp
}

func (cp *ControlPanel) Finalize() {
	cp.equalizeTabPanels()
	cp.enableChangeNotifications = true
}

func (cp *ControlPanel) SetChangeListener(l func(tag string)) {
	v := cp.disableChangeNotification()
	defer cp.setChangeNotification(v)
	cp.changeListener = l
}

func (cp *ControlPanel) SetWidgetFloatValue(tag string, val float64) {
	v := cp.disableChangeNotification()
	defer cp.setChangeNotification(v)

	item, ok := cp.uiItems[tag]
	if ok {
		item.floatVal.Set(val)
	}
}

func (cp *ControlPanel) SetChoiceByIndex(tag string, index int) {
	v := cp.disableChangeNotification()
	defer cp.setChangeNotification(v)

	item, ok := cp.uiItems[tag]
	if ok {
		item.intVal.Set(index)
		w := item.widget.(*widget.Select)
		if w != nil {
			w.SetSelectedIndex(index)
		}
	}
}

func (cp *ControlPanel) SetCheckboxState(tag string, checked bool) {
	v := cp.disableChangeNotification()
	defer cp.setChangeNotification(v)

	item, ok := cp.uiItems[tag]
	if ok {
		item.boolVal.Set(checked)
		w := item.widget.(*widget.Check)
		if w != nil {
			w.SetChecked(checked)
		}
	}
}

func (cp *ControlPanel) GetCheckBoxState(tag string) (val bool, ok bool) {
	item, ok := cp.uiItems[tag]
	if ok {
		v, err := item.boolVal.Get()
		if err != nil {
			fyne.LogError("Error reading UI item", err)
			ok = false
		}

		val = v
	}
	return
}

func (cp *ControlPanel) GetWidgetFloatValue(tag string) (val float64, ok bool) {
	item, ok := cp.uiItems[tag]
	if ok {
		v, err := item.floatVal.Get()
		if err != nil {
			fyne.LogError("Error reading UI item", err)
			ok = false
		}

		val = v
	}
	return
}

func (cp *ControlPanel) GetWidgetIntValue(tag string) (val int, ok bool) {
	item, ok := cp.uiItems[tag]
	if ok {
		v, err := item.intVal.Get()

		if err != nil {
			fyne.LogError("Error reading UI item", err)
			ok = false
		}

		val = v
	}
	return
}

func (cp *ControlPanel) AddGroup(tag string, title string) {
	_, ok := cp.groups[tag]
	if !ok {
		content := container.NewGridWithColumns(2)
		content.Resize(fyne.NewSize(400, 10))
		g := container.NewTabItem(title, content)
		cp.root.Append(g)
		cp.groups[tag] = g
	}
}

func (cp *ControlPanel) AddNumberEntry(
	addToGroupTag string,
	itemTag string,
	label string,
	config NumericalEditConfigConfig) {

	// Find the container for the host group.
	tabItem, ok := cp.groups[addToGroupTag]
	if !ok {
		fyne.LogError("Unknown group tag", nil)
		return
	}

	// Enter a new UI item in the map ensuring there's no tag duplication.
	item, ok := cp.uiItems[itemTag]
	if ok {
		fyne.LogError("Duplicate UI item tag: "+itemTag, nil)
		return
	}
	item = &uiItemValueBinding{
		floatVal: binding.NewFloat(),
	}
	cp.uiItems[itemTag] = item
	item.floatVal.AddListener(newTaggedChangeListener(itemTag, cp))

	// Create the UI elements and insert in the grid.
	w := widget.NewEntryWithData(
		NewFloatRangeBindingWithFormat(
			item.floatVal, config.MinVal, config.MaxVal, config.Format, config.Regex))
	item.widget = w

	panel := tabItem.Content.(*fyne.Container)
	if panel != nil {
		panel.Add(widget.NewLabel(label))
		panel.Add(w)
	}
}

func (cp *ControlPanel) AddSelector(
	addToGroupTag string,
	itemTag string,
	label string,
	choices []string) {

	// Find the container for the host group.
	tabItem, ok := cp.groups[addToGroupTag]
	if !ok {
		fyne.LogError("Unknown group tag", nil)
		return
	}

	// Enter a new UI item in the map ensuring there's no tag duplication.
	item, ok := cp.uiItems[itemTag]
	if ok {
		fyne.LogError("Duplicate UI item tag: "+itemTag, nil)
		return
	}
	item = &uiItemValueBinding{
		intVal: binding.NewInt(),
	}
	item.intVal.AddListener(newTaggedChangeListener(itemTag, cp))

	// Create the UI elements and insert in the grid.
	w := widget.NewSelect(choices, nil)
	w.OnChanged = func(choice string) {
		item.intVal.Set(w.SelectedIndex())
	}
	item.widget = w
	cp.uiItems[itemTag] = item

	panel := tabItem.Content.(*fyne.Container)
	if panel != nil {
		panel.Add(widget.NewLabel(label))
		panel.Add(w)
	}
}

func (cp *ControlPanel) AddCheckbox(
	addToGroupTag string,
	itemTag string,
	label string) {

	// Find the container for the host group.
	tabItem, ok := cp.groups[addToGroupTag]
	if !ok {
		fyne.LogError("Unknown group tag", nil)
		return
	}

	// Enter a new UI item in the map ensuring there's no tag duplication.
	item, ok := cp.uiItems[itemTag]
	if ok {
		fyne.LogError("Duplicate UI item tag: "+itemTag, nil)
		return
	}
	item = &uiItemValueBinding{
		boolVal: binding.NewBool(),
	}
	item.boolVal.AddListener(newTaggedChangeListener(itemTag, cp))

	// Create the UI elements and insert in the grid.
	w := widget.NewCheckWithData("", item.boolVal)
	item.widget = w
	cp.uiItems[itemTag] = item

	panel := tabItem.Content.(*fyne.Container)
	if panel != nil {
		panel.Add(widget.NewLabel(label))
		panel.Add(w)
	}
}

func (cp *ControlPanel) getRoot() fyne.CanvasObject {
	return cp.root
}

// Interface taggedChangeListener.
func (cp *ControlPanel) onChange(tag string) {
	if cp.enableChangeNotifications && cp.changeListener != nil {
		cp.changeListener(tag)
	}
}

func (cp *ControlPanel) disableChangeNotification() (previousState bool) {
	previousState = cp.enableChangeNotifications
	cp.enableChangeNotifications = false
	return
}

func (cp *ControlPanel) setChangeNotification(newState bool) {
	cp.enableChangeNotifications = newState
}

func (cp *ControlPanel) equalizeTabPanels() {
	// Find the largest number of grid rows in all tab panes.
	maxGridRows := 0
	for _, g := range cp.groups {
		panel := g.Content.(*fyne.Container)
		if panel != nil {
			n := len(panel.Objects) / 2
			if n > maxGridRows {
				maxGridRows = n
			}
		}
	}

	// Fill all the grids with spacers so that they have the same number of rows. This
	// is necessary to insure that rows have the same heights across all the tab panes.
	for _, g := range cp.groups {
		panel := g.Content.(*fyne.Container)
		if panel != nil {
			n := len(panel.Objects) / 2
			for n < maxGridRows {
				panel.Add(layout.NewSpacer())
				panel.Add(layout.NewSpacer())
				n++
			}
		}
	}
}
