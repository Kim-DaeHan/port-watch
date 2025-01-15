package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"github.com/Kim-DaeHan/port-watch/ui"
)

func main() {
	myApp := app.New()
	window := myApp.NewWindow("Port Watch")
	window.SetTitle("")

	window.Resize(fyne.NewSize(400, 600))

	portManager := ui.NewPortManagerUI(window)
	portManager.Setup()

	window.CenterOnScreen()
	window.ShowAndRun()
}
