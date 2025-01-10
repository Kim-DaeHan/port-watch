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
	"fyne.io/fyne/v2/layout"
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
func canKillProcess(processName string, pid string) (bool, bool) {
	// 종료 불가능한 프로세스
	protectedProcesses := map[string]bool{
		"systemd":      true,
		"sshd":         true,
		"init":         true,
		"kernel":       true,
		"launchd":      true,
		"WindowServer": true,
		"loginwindow":  true,
	}

	// 종료 위험한 프로세스
	dangerousProcesses := map[string]bool{
		"postgres":  true,
		"mongod":    true,
		"redis-ser": true,
		"mysqld":    true,
		"nginx":     true,
		"docker":    true,
	}

	if protectedProcesses[processName] {
		return false, false
	}

	pidNum, _ := strconv.Atoi(pid)
	if pidNum < 1000 {
		return false, false
	}

	return true, dangerousProcesses[processName]
}

// 프로세스 종료 함수 추가
func killProcess(pid string) error {
	// 먼저 일반적인 SIGTERM으로 종료 시도
	cmd := exec.Command("kill", pid)
	err := cmd.Run()

	if err != nil {
		// SIGTERM으로 종료 실패한 경우에만 SIGKILL(-9) 사용
		cmd = exec.Command("kill", "-9", pid)
		return cmd.Run()
	}

	return nil
}

func main() {
	myApp := app.New()
	window := myApp.NewWindow("Port Manager")
	window.Resize(fyne.NewSize(800, 600))

	var ports []PortInfo
	var list *widget.List

	updateList := func() {
		newPorts, err := getActivePorts()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		ports = newPorts
		if list != nil {
			list.Refresh()
		}
	}

	// 헤더 생성
	header := container.NewHBox(
		widget.NewLabel("Active Ports"),
		layout.NewSpacer(),
		widget.NewButton("Refresh", updateList),
	)

	// 리스트 정의
	list = widget.NewList(
		func() int {
			return len(ports)
		},
		func() fyne.CanvasObject {
			button := widget.NewButton("Terminate", nil)
			return container.NewHBox(
				container.NewVBox(
					widget.NewLabel(""),
					widget.NewLabel(""),
				),
				layout.NewSpacer(),
				button,
			)
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			container := item.(*fyne.Container)
			infoContainer := container.Objects[0].(*fyne.Container)
			processLabel := infoContainer.Objects[0].(*widget.Label)
			detailsLabel := infoContainer.Objects[1].(*widget.Label)
			button := container.Objects[2].(*widget.Button)

			port := ports[id]

			processLabel.SetText(port.ProcessName)
			processLabel.TextStyle = fyne.TextStyle{Bold: true}
			processLabel.Resize(fyne.NewSize(400, 30))

			detailsLabel.SetText(fmt.Sprintf("PID: %s • %s:%s",
				port.PID,
				port.Protocol,
				port.Port))
			detailsLabel.TextStyle = fyne.TextStyle{Monospace: true}
			detailsLabel.Resize(fyne.NewSize(400, 30))

			canKill, isDangerous := canKillProcess(port.ProcessName, port.PID)
			if canKill {
				if isDangerous {
					button.Importance = widget.DangerImportance
				} else {
					button.Importance = widget.MediumImportance
				}
				button.Show()
				button.Resize(fyne.NewSize(100, 40))

				button.OnTapped = func() {
					if isDangerous {
						dialog.ShowConfirm(
							"경고: 위험한 프로세스",
							fmt.Sprintf("%s (PID: %s)을 종료하시겠습니까?\n이 프로세스는 시스템에 중요합니다.",
								port.ProcessName,
								port.PID),
							func(ok bool) {
								if ok {
									if err := killProcess(port.PID); err != nil {
										dialog.ShowError(err, window)
									} else {
										updateList()
									}
								}
							},
							window,
						)
					} else {
						dialog.ShowConfirm(
							"프로세스 종료",
							fmt.Sprintf("%s (PID: %s)을 종료하시겠습니까?",
								port.ProcessName,
								port.PID),
							func(ok bool) {
								if ok {
									if err := killProcess(port.PID); err != nil {
										dialog.ShowError(err, window)
									} else {
										updateList()
									}
								}
							},
							window,
						)
					}
				}
			} else {
				button.Hide()
			}
		},
	)

	// 자동 새로고침
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		for range ticker.C {
			updateList()
		}
	}()

	// 메인 컨테이너 구성
	mainContainer := container.NewBorder(
		container.NewVBox(
			header,
			widget.NewSeparator(),
		),
		nil,
		nil,
		nil,
		container.NewVScroll(list),
	)

	window.SetContent(mainContainer)
	window.CenterOnScreen()

	// 초기 데이터 로드
	updateList()

	window.ShowAndRun()
}
