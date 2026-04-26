package cpu

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type CPUInfo struct {
	Load  float64
	Usage float64
	Cores int
}

func GetCPUInfo() CPUInfo {
	load, err := loadAvarage()
	if err != nil {
		fmt.Println("Error:", err)
	}
	usage, err := cpuUsage()
	if err != nil {
		fmt.Println("Error:", err)
	}
	return CPUInfo{
		Load:  load,
		Usage: usage,
		Cores: runtime.NumCPU(),
	}
}

func loadAvarage() (float64, error) {
	// Read the contents of /proc/loadavg
	data, err := os.ReadFile("/proc/loadavg")
	if err != nil {
		return 0, err
	}

	// Parse the contents
	loadavg := strings.Fields(string(data))

	// string to float64
	load, err := strconv.ParseFloat(loadavg[0], 64)
	if err != nil {
		return 0, err
	}

	return load, nil
}

func cpuUsage() (float64, error) {
	idle1, total1, err := cpuTimes()
	if err != nil {
		return 0, err
	}

	time.Sleep(time.Second)

	idle2, total2, err := cpuTimes()
	if err != nil {
		return 0, err
	}

	totalDelta := total2 - total1
	idleDelta := idle2 - idle1
	if totalDelta <= 0 {
		return 0, nil
	}

	return (1 - float64(idleDelta)/float64(totalDelta)) * 100, nil
}

func cpuTimes() (uint64, uint64, error) {
	data, err := os.ReadFile("/proc/stat")
	if err != nil {
		return 0, 0, err
	}

	lines := strings.Split(string(data), "\n")
	if len(lines) == 0 {
		return 0, 0, fmt.Errorf("/proc/stat is empty")
	}

	fields := strings.Fields(lines[0])
	if len(fields) < 8 {
		return 0, 0, fmt.Errorf("unexpected /proc/stat format")
	}

	var total uint64
	for _, value := range fields[1:] {
		parsed, parseErr := strconv.ParseUint(value, 10, 64)
		if parseErr != nil {
			return 0, 0, parseErr
		}
		total += parsed
	}

	idle, err := strconv.ParseUint(fields[4], 10, 64)
	if err != nil {
		return 0, 0, err
	}

	return idle, total, nil
}
