package ui

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"github.com/Kim-DaeHan/port-watch/process"
	"github.com/Kim-DaeHan/port-watch/types"
)

type PortManagerUI struct {
	window fyne.Window
	ports  []types.PortInfo
	list   *widget.List
}

func NewPortManagerUI(window fyne.Window) *PortManagerUI {
	return &PortManagerUI{
		window: window,
	}
}

func (ui *PortManagerUI) updateList() {
	newPorts, err := process.GetActivePorts()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	ui.ports = newPorts
	if ui.list != nil {
		ui.list.Refresh()
	}
}

func (ui *PortManagerUI) Setup() {
	header := container.NewHBox(
		widget.NewLabel("Active Ports"),
		layout.NewSpacer(),
		widget.NewButton("Refresh", ui.updateList),
	)

	ui.list = widget.NewList(
		func() int {
			return len(ui.ports)
		},
		func() fyne.CanvasObject {
			button := widget.NewButton("Terminate", nil)
			button.Resize(fyne.NewSize(100, 35))

			return container.NewHBox(
				container.NewVBox(
					widget.NewLabel(""),
					widget.NewLabel(""),
				),
				layout.NewSpacer(),
				container.NewHBox(
					layout.NewSpacer(),
					button,
				),
			)
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			ui.updateListItem(id, item)
		},
	)

	go ui.startAutoRefresh()

	mainContainer := container.NewBorder(
		container.NewVBox(
			header,
			widget.NewSeparator(),
		),
		nil,
		nil,
		nil,
		container.NewVScroll(ui.list),
	)

	ui.window.SetContent(mainContainer)
	ui.updateList()
}

func (ui *PortManagerUI) updateListItem(id widget.ListItemID, item fyne.CanvasObject) {
	container := item.(*fyne.Container)
	infoContainer := container.Objects[0].(*fyne.Container)
	processLabel := infoContainer.Objects[0].(*widget.Label)
	detailsLabel := infoContainer.Objects[1].(*widget.Label)
	buttonContainer := container.Objects[2].(*fyne.Container)
	button := buttonContainer.Objects[1].(*widget.Button)

	port := ui.ports[id]

	processLabel.SetText(port.ProcessName)
	processLabel.TextStyle = fyne.TextStyle{Bold: true}
	processLabel.Resize(fyne.NewSize(400, 30))

	detailsLabel.SetText(fmt.Sprintf("PID: %s • %s:%s",
		port.PID,
		port.Protocol,
		port.Port))
	detailsLabel.TextStyle = fyne.TextStyle{Monospace: true}
	detailsLabel.Resize(fyne.NewSize(400, 30))

	canKill, isDangerous := process.CanKillProcess(port.ProcessName, port.PID)
	if canKill {
		button.Show()
		button.OnTapped = func() {
			ui.showKillConfirmDialog(port, isDangerous)
		}
	} else {
		button.Hide()
	}
}

func (ui *PortManagerUI) showKillConfirmDialog(port types.PortInfo, isDangerous bool) {
	title := "프로세스 종료"
	message := fmt.Sprintf("%s (PID: %s)을 종료하시겠습니까?", port.ProcessName, port.PID)

	if isDangerous {
		title = "경고: 위험한 프로세스"
		message = fmt.Sprintf("%s (PID: %s)을 종료하시겠습니까?\n이 프로세스는 시스템에 중요합니다.",
			port.ProcessName, port.PID)
	}

	dialog.ShowConfirm(title, message,
		func(ok bool) {
			if ok {
				if err := process.KillProcess(port.PID); err != nil {
					dialog.ShowError(err, ui.window)
				} else {
					ui.updateList()
				}
			}
		},
		ui.window,
	)
}

func (ui *PortManagerUI) startAutoRefresh() {
	ticker := time.NewTicker(5 * time.Second)
	for range ticker.C {
		ui.updateList()
	}
}
