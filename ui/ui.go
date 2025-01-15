package ui

import (
	"fmt"
	"image/color"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/Kim-DaeHan/port-watch/process"
	"github.com/Kim-DaeHan/port-watch/types"
)

type PortManagerUI struct {
	window        fyne.Window
	ports         []types.PortInfo
	list          *widget.List
	searchEntry   *widget.Entry
	filteredPorts []types.PortInfo
}

func NewPortManagerUI(window fyne.Window) *PortManagerUI {
	return &PortManagerUI{
		window:        window,
		filteredPorts: []types.PortInfo{},
	}
}

func (ui *PortManagerUI) updateList() {
	newPorts, err := process.GetActivePorts()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	ui.ports = newPorts
	ui.filterPorts(ui.searchEntry.Text)
}

func (ui *PortManagerUI) filterPorts(searchText string) {
	var filteredPorts []types.PortInfo
	seen := make(map[string]bool) // 중복 체크를 위한 맵

	// 중복 체크 키 생성 함수
	makeKey := func(port types.PortInfo) string {
		return fmt.Sprintf("%s:%s:%s", port.ProcessName, port.Port, port.PID)
	}

	// 검색어로 필터링하고 중복 제거
	for _, port := range ui.ports {
		// 검색어가 있으면 프로세스 이름으로 필터링
		if searchText != "" && !strings.Contains(
			strings.ToLower(port.ProcessName),
			strings.ToLower(searchText)) {
			continue
		}

		// 중복 체크
		key := makeKey(port)
		if !seen[key] {
			seen[key] = true
			filteredPorts = append(filteredPorts, port)
		}
	}

	ui.filteredPorts = filteredPorts
	if ui.list != nil {
		ui.list.Refresh()
	}
}

func (ui *PortManagerUI) Setup() {
	ui.searchEntry = widget.NewEntry()
	ui.searchEntry.SetPlaceHolder("프로세스 이름으로 검색...")
	ui.searchEntry.OnChanged = func(text string) {
		ui.filterPorts(text)
	}

	header := container.NewVBox(
		container.NewHBox(
			widget.NewLabel("Port Watch"),
			layout.NewSpacer(),
			widget.NewButton("Refresh", ui.updateList),
		),
		ui.searchEntry,
	)

	ui.list = widget.NewList(
		func() int {
			return len(ui.filteredPorts)
		},
		func() fyne.CanvasObject {
			text := canvas.NewText("", &color.NRGBA{R: 57, G: 255, B: 20, A: 255})
			text.TextStyle = fyne.TextStyle{Bold: true}
			text.TextSize = 14

			button := widget.NewButtonWithIcon("", theme.CancelIcon(), nil)
			button.Importance = widget.DangerImportance
			button.Resize(fyne.NewSize(45, 35))

			return container.NewHBox(
				text,
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
	text := container.Objects[0].(*canvas.Text)
	buttonContainer := container.Objects[2].(*fyne.Container)
	button := buttonContainer.Objects[1].(*widget.Button)

	port := ui.filteredPorts[id]

	text.Text = fmt.Sprintf("%s:%s", port.ProcessName, port.Port)
	text.Refresh()

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
