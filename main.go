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
		fmt.Println("Error executing lsof:", err)
		return nil, err
	}

	var ports []PortInfo
	lines := strings.Split(string(output), "\n")

	for _, line := range lines[1:] {
		fields := strings.Fields(line)
		if len(fields) < 9 {
			continue
		}

		addr := fields[8]
		protocol := "TCP"
		if strings.Contains(fields[7], "UDP") {
			protocol = "UDP"
		}

		parts := strings.Split(addr, ":")
		if len(parts) < 2 {
			continue
		}

		portInfo := PortInfo{
			ProcessName: fields[0],
			PID:         fields[1],
			Protocol:    protocol,
			Port:        parts[len(parts)-1],
		}
		fmt.Printf("Found port: %+v\n", portInfo)
		ports = append(ports, portInfo)
	}

	fmt.Printf("Total ports found: %d\n", len(ports))
	return ports, nil
}

func main() {
	myApp := app.New()
	window := myApp.NewWindow("Port Watch")

	var ports []PortInfo
	list := widget.NewList(
		func() int {
			fmt.Printf("Length function called, returning: %d\n", len(ports))
			return len(ports)
		},
		func() fyne.CanvasObject {
			label := widget.NewLabel("")
			label.Resize(fyne.NewSize(550, 30))
			return label
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			label := item.(*widget.Label)
			port := ports[id]
			text := fmt.Sprintf("%s (PID:%s) - %s:%s",
				port.ProcessName,
				port.PID,
				port.Protocol,
				port.Port)
			fmt.Printf("Setting text for id %d: %s\n", id, text)
			label.SetText(text)
		},
	)

	updateList := func() {
		newPorts, err := getActivePorts()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		fmt.Printf("UpdateList called with %d ports\n", len(newPorts))
		ports = newPorts
		list.Refresh()
	}

	// 초기 데이터 로드
	updateList()

	// 5초마다 자동 새로고침
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		for range ticker.C {
			updateList()
		}
	}()

	window.SetContent(container.NewVBox(
		widget.NewLabel("실행 중인 포트 목록"),
		list,
	))

	window.Resize(fyne.NewSize(600, 800))
	window.ShowAndRun()
}
