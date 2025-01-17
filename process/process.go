package process

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/Kim-DaeHan/port-watch/types"
)

func GetActivePorts() ([]types.PortInfo, error) {
	cmd := exec.Command("lsof", "-i", "-P", "-n")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	seen := make(map[string]bool)
	var ports []types.PortInfo
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

		portInfo := types.PortInfo{
			ProcessName: fields[0],
			PID:         fields[1],
			Protocol:    protocol,
			Port:        parts[len(parts)-1],
		}

		key := fmt.Sprintf("%s:%s:%s", portInfo.ProcessName, portInfo.Port, portInfo.PID)
		if !seen[key] {
			seen[key] = true
			ports = append(ports, portInfo)
		}
	}

	return ports, nil
}

func CanKillProcess(processName string, pid string) (bool, bool) {
	protectedProcesses := map[string]bool{
		"systemd":      true,
		"sshd":         true,
		"init":         true,
		"kernel":       true,
		"launchd":      true,
		"WindowServer": true,
		"loginwindow":  true,
	}

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

func KillProcess(pid string) error {
	cmd := exec.Command("kill", pid)
	err := cmd.Run()

	if err != nil {
		cmd = exec.Command("kill", "-9", pid)
		return cmd.Run()
	}

	return nil
}

func ForceKillProcess(processName string) error {
	cmd := exec.Command("pkill", "-9", processName)
	return cmd.Run()
}
