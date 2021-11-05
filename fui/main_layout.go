package fui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
)

const (
	imageCanvasSize float32 = 512.
)

type MainLayout struct {
	root  *fyne.Container
	image *ImagePanel
	cp    *ControlPanel
}

func NewTopLayout() *MainLayout {
	ml := &MainLayout{}

	ml.root = container.NewHBox(
		ml.createImagePanel(),
		ml.createControlPanel(),
	)

	return ml
}

func (t *MainLayout) GetRootContainer() *fyne.Container {
	return t.root
}

func (t *MainLayout) GetControlPanel() *ControlPanel {
	return t.cp
}

func (t *MainLayout) GetImagePanel() *ImagePanel {
	return t.image
}

func (t *MainLayout) createImagePanel() *fyne.Container {
	t.image = NewImagePanel(imageCanvasSize)
	return container.NewPadded(t.image.getRootContainer())
}

func (t *MainLayout) createControlPanel() *fyne.Container {
	t.cp = NewControlPanel()
	return container.NewVBox(
		container.NewPadded(t.cp.getRoot()),
		layout.NewSpacer(),
	)
}
