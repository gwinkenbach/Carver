package qtui

import (
	"github.com/therecipe/qt/widgets"
)

type MainLayout struct {
	root    *widgets.QWidget
	img     *ImagePanel
	control *ControlPanel
}

func NewMainLayout() *MainLayout {
	ml := MainLayout{}

	layout := widgets.NewQHBoxLayout()
	ml.root = widgets.NewQWidget(nil, 0)
	ml.root.SetLayout(layout)

	ml.img = NewImagePanel(layout)

	ml.control = NewControlPanel(layout)

	return &ml
}

func (ml *MainLayout) GetRoot() *widgets.QWidget {
	return ml.root
}

func (ml *MainLayout) GetControlPanel() *ControlPanel {
	return ml.control
}

func (ml *MainLayout) GetImagePanel() *ImagePanel {
	return ml.img
}
