package qtui

import (
	"github.com/therecipe/qt/widgets"
)

type ControlPanel struct {
	root *widgets.QFormLayout
}

func NewControlPanel(parent *widgets.QHBoxLayout) *ControlPanel {
	cp := ControlPanel{}

	cp.root = widgets.NewQFormLayout(nil)
	parent.AddLayout(cp.root, 0)

	return &cp
}

func (cp *ControlPanel) AddLabeledWidget(label string, w widgets.QWidget_ITF) {
	cp.root.AddRow3(label, w)
}

func (cp *ControlPanel) AddVerticalSpace(s int) {
	spacer := widgets.NewQSpacerItem(10, s,
		widgets.QSizePolicy__Expanding, widgets.QSizePolicy__Minimum)
	cp.root.AddItem(spacer)
}
