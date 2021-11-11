package main

import (
	"alvin.com/GoCarver/model"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/theme"
)

func main() {
	a := app.New()
	a.Settings().SetTheme(theme.LightTheme())

	w := a.NewWindow("Go Carver")

	uiManager := model.NewUIManager()
	m := model.NewModel()
	c := model.NewController(m)
	c.ConnectUI(uiManager, w)

	w.SetCloseIntercept(func() {
		if c.CheckShouldClose() {
			w.Close()
		}
	})

	a.Lifecycle().SetOnStarted(func() {
		uiManager.FinalizeUI()
	})

	w.Resize(fyne.NewSize(820, 600))
	w.SetFixedSize(true)
	w.ShowAndRun()
}
