package main

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
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

	for _, line := range lines[1:] {
		fields := strings.Fields(line)
		if len(fields) < 8 {
			continue
		}

		if fields[len(fields)-1] != "(LISTEN)" {
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

		ports = append(ports, portInfo)
	}

	return ports, nil
}

// 프로세스 종료 가능 여부 확인 함수
func canKillProcess(processName string, pid string) bool {
	// 종료 불가능한 프로세스 목록
	protectedProcesses := map[string]bool{
		"systemd":      true,
		"sshd":         true,
		"init":         true,
		"kernel":       true,
		"launchd":      true, // macOS 시스템 프로세스
		"WindowServer": true, // macOS 시스템 프로세스
		"loginwindow":  true, // macOS 시스템 프로세스
	}

	if protectedProcesses[processName] {
		return false
	}

	// PID가 1000 미만인 경우는 대부분 시스템 프로세스
	pidNum, _ := strconv.Atoi(pid)
	return pidNum >= 1000
}

// 프로세스 종료 함수 추가
func killProcess(pid string) error {
	cmd := exec.Command("kill", pid)
	return cmd.Run()
}

func main() {
	myApp := app.New()
	window := myApp.NewWindow("Port Watch")

	var ports []PortInfo
	var list *widget.List

	// updateList 함수를 먼저 선언
	updateList := func() {
		newPorts, err := getActivePorts()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		ports = newPorts
		list.Refresh()
	}

	// 리스트 정의
	list = widget.NewList(
		func() int {
			return len(ports)
		},
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewLabel(""),
				widget.NewButton("종료", func() {}),
			)
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			container := item.(*fyne.Container)
			label := container.Objects[0].(*widget.Label)
			button := container.Objects[1].(*widget.Button)

			port := ports[id]
			text := fmt.Sprintf("%s (PID:%s) - %s:%s",
				port.ProcessName,
				port.PID,
				port.Protocol,
				port.Port)
			label.SetText(text)

			// 프로세스 종료 가능 여부에 따라 버튼 표시/숨김
			if canKillProcess(port.ProcessName, port.PID) {
				button.Show()
				button.OnTapped = func() {
					dialog.ShowConfirm("프로세스 종료",
						fmt.Sprintf("정말로 %s(PID:%s) 프로세스를 종료하시겠습니까?",
							port.ProcessName, port.PID),
						func(ok bool) {
							if ok {
								if err := killProcess(port.PID); err != nil {
									dialog.ShowError(err, window)
								} else {
									updateList()
								}
							}
						}, window)
				}
			} else {
				button.Hide()
			}
		},
	)

	// 초기 데이터 로드
	updateList()

	// 5초마다 자동 새로고침
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		for range ticker.C {
			updateList()
		}
	}()

	listContainer := container.NewVScroll(list)

	content := container.NewBorder(
		widget.NewLabel("실행 중인 포트 목록"), // 상단
		nil,           // 하단
		nil,           // 좌측
		nil,           // 우측
		listContainer, // 중앙 (나머지 공간 모두 차지)
	)

	window.SetContent(content)
	window.Resize(fyne.NewSize(600, 800))
	window.ShowAndRun()
}
