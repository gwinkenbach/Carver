package main

import (
	"os"

	"github.com/therecipe/qt/gui"

	"alvin.com/GoCarver/model"
	"alvin.com/GoCarver/qtui"

	"github.com/therecipe/qt/widgets"
)

var menuBar *qtui.MenuBar

func main() {
	app := widgets.NewQApplication(len(os.Args), os.Args)

	window := widgets.NewQMainWindow(nil, 0)
	window.SetMinimumSize2(680, 512)
	window.SetWindowTitle("Carver")

	uiMgr := qtui.NewUIManager()
	uiMgr.BuildUI()

	m := model.NewModel()
	c := model.NewController(uiMgr, m)
	c.ConnectUI(window)

	menuBar := qtui.CreateMenuBar(window)
	c.SetMenuBar(menuBar)

	window.SetCentralWidget(uiMgr.GetRootPanel())
	window.ConnectCloseEvent(func(event *gui.QCloseEvent) {
		c.CheckShouldClose(event)
	})
	window.Show()

	app.Exec()
}
