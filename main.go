package main

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type PortInfo struct {
	ProcessName string
	PID         string
	Protocol    string
	Port        string
}

func getActivePorts() ([]PortInfo, error) {
	cmd := exec.Command("lsof", "-i", "-P", "-n")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var ports []PortInfo
	lines := strings.Split(string(output), "\n")

	for _, line := range lines[1:] { // 헤더 스킵
		fields := strings.Fields(line)
		if len(fields) < 9 {
			continue
		}

		if strings.Contains(fields[8], "LISTEN") {
			portInfo := PortInfo{
				ProcessName: fields[0],
				PID:         fields[1],
				Protocol:    fields[7],
				Port:        strings.Split(fields[8], ":")[1],
			}
			ports = append(ports, portInfo)
		}
	}

	return ports, nil
}

func main() {
	myApp := app.New()
	window := myApp.NewWindow("Port Watch")

	// 리스트 위젯 생성
	list := widget.NewList(
		func() int { return 0 },
		func() fyne.CanvasObject {
			return widget.NewLabel("Template")
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			// 업데이트 시 구현
		},
	)

	// 자동 새로고침 함수
	updateList := func() {
		ports, err := getActivePorts()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		list.Length = func() int { return len(ports) }
		list.UpdateItem = func(id widget.ListItemID, item fyne.CanvasObject) {
			label := item.(*widget.Label)
			label.SetText(fmt.Sprintf("%s (PID: %s) - %s:%s",
				ports[id].ProcessName,
				ports[id].PID,
				ports[id].Protocol,
				ports[id].Port))
		}
	}
	list.Refresh()

	// 초기 데이터 로드
	updateList()

	// 5초마다 자동 새로고침
	ticker := time.NewTicker(5 * time.Second)
	go func() {
		for range ticker.C {
			updateList()
		}
	}()

	window.SetContent(container.NewVBox(
		widget.NewLabel("실행 중인 포트 목록"),
		list,
	))

	window.Resize(fyne.NewSize(400, 600))
	window.ShowAndRun()
}
