package disk

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type DiskInfo struct {
	Total              string
	Free               string
	Used               string
	ReadBytesPerSecond string
	WriteBytesPerSecond string
	TPS                string
	IOWait             string
}

func GetDiskInfo() DiskInfo {
	// Get disk usage of the current working directory
	info, err := diskUsage("/")
	if err != nil {
		fmt.Println("Error:", err)
		return DiskInfo{}
	}

	read1, write1, tps1, ioWait1, total1, err := diskStats()
	if err != nil {
		fmt.Println("Error:", err)
		return info
	}

	time.Sleep(time.Second)

	read2, write2, tps2, ioWait2, total2, err := diskStats()
	if err != nil {
		fmt.Println("Error:", err)
		return info
	}

	info.ReadBytesPerSecond = strconv.FormatUint((read2-read1)*512, 10)
	info.WriteBytesPerSecond = strconv.FormatUint((write2-write1)*512, 10)
	info.TPS = strconv.FormatUint(tps2-tps1, 10)
	info.IOWait = formatPercentageDelta(ioWait1, ioWait2, total1, total2)

	return info
}

// DiskUsage returns the disk usage of the path in bytes
func diskUsage(path string) (usage DiskInfo, err error) {
	// Run the df command to get disk information
	cmd := exec.Command("df", "-BM", path) // Use -BM option to get sizes in megabytes
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("Error running df command:", err)
		return
	}

	// Parse the output to get disk total, usage, and free
	lines := strings.Split(string(output), "\n")
	if len(lines) < 2 {
		fmt.Println("Unexpected output from df command")
		return
	}

	fields := strings.Fields(lines[1])
	if len(fields) < 4 {
		fmt.Println("Unexpected output format")
		return
	}

	usage.Total = strings.Replace(fields[1], "M", "", 1)
	usage.Used = strings.Replace(fields[2], "M", "", 1)
	usage.Free = strings.Replace(fields[3], "M", "", 1)
	return
}

func diskStats() (uint64, uint64, uint64, uint64, uint64, error) {
	procDiskStats, err := os.ReadFile("/proc/diskstats")
	if err != nil {
		return 0, 0, 0, 0, 0, err
	}

	procStat, err := os.ReadFile("/proc/stat")
	if err != nil {
		return 0, 0, 0, 0, 0, err
	}

	var readSectors uint64
	var writeSectors uint64
	var completedIO uint64

	for _, line := range strings.Split(string(procDiskStats), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 14 {
			continue
		}

		name := fields[2]
		if strings.HasPrefix(name, "loop") || strings.HasPrefix(name, "ram") {
			continue
		}

		reads, err := strconv.ParseUint(fields[5], 10, 64)
		if err != nil {
			return 0, 0, 0, 0, 0, err
		}
		writes, err := strconv.ParseUint(fields[9], 10, 64)
		if err != nil {
			return 0, 0, 0, 0, 0, err
		}
		readOps, err := strconv.ParseUint(fields[3], 10, 64)
		if err != nil {
			return 0, 0, 0, 0, 0, err
		}
		writeOps, err := strconv.ParseUint(fields[7], 10, 64)
		if err != nil {
			return 0, 0, 0, 0, 0, err
		}

		readSectors += reads
		writeSectors += writes
		completedIO += readOps + writeOps
	}

	statFields := strings.Fields(strings.Split(string(procStat), "\n")[0])
	if len(statFields) < 8 {
		return 0, 0, 0, 0, 0, fmt.Errorf("unexpected /proc/stat format")
	}

	ioWait, err := strconv.ParseUint(statFields[5], 10, 64)
	if err != nil {
		return 0, 0, 0, 0, 0, err
	}

	var total uint64
	for _, value := range statFields[1:] {
		parsed, parseErr := strconv.ParseUint(value, 10, 64)
		if parseErr != nil {
			return 0, 0, 0, 0, 0, parseErr
		}
		total += parsed
	}

	return readSectors, writeSectors, completedIO, ioWait, total, nil
}

func formatPercentageDelta(ioWait1 uint64, ioWait2 uint64, total1 uint64, total2 uint64) string {
	totalDelta := total2 - total1
	if totalDelta == 0 {
		return "0"
	}

	percentage := (float64(ioWait2-ioWait1) / float64(totalDelta)) * 100
	return strconv.FormatFloat(percentage, 'f', 2, 64)
}
