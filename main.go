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

	// 개발 모드에서 기본 아이콘 설정 (선택사항)
	if icon, err := fyne.LoadResourceFromPath("icon.icns"); err == nil {
		window.SetIcon(icon)
	}

	window.Resize(fyne.NewSize(400, 600))

	portManager := ui.NewPortManagerUI(window)
	portManager.Setup()

	window.CenterOnScreen()
	window.ShowAndRun()
}
