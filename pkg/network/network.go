package network

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type NetworkInfo struct {
	UpstreamBytesPerSecond   string
	DownstreamBytesPerSecond string
	TotalSent                string
	TotalReceived            string
}

func GetNetworkInfo() NetworkInfo {
	rx1, tx1, err := networkCounters()
	if err != nil {
		fmt.Println("Error:", err)
		return NetworkInfo{}
	}

	time.Sleep(time.Second)

	rx2, tx2, err := networkCounters()
	if err != nil {
		fmt.Println("Error:", err)
		return NetworkInfo{}
	}

	return NetworkInfo{
		UpstreamBytesPerSecond:   strconv.FormatUint(tx2-tx1, 10),
		DownstreamBytesPerSecond: strconv.FormatUint(rx2-rx1, 10),
		TotalSent:                strconv.FormatUint(tx2, 10),
		TotalReceived:            strconv.FormatUint(rx2, 10),
	}
}

func networkCounters() (uint64, uint64, error) {
	data, err := os.ReadFile("/proc/net/dev")
	if err != nil {
		return 0, 0, err
	}

	var rxTotal uint64
	var txTotal uint64

	lines := strings.Split(string(data), "\n")
	for _, line := range lines[2:] {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.Fields(strings.ReplaceAll(line, ":", " "))
		if len(fields) < 17 || fields[0] == "lo" {
			continue
		}

		rx, parseErr := strconv.ParseUint(fields[1], 10, 64)
		if parseErr != nil {
			return 0, 0, parseErr
		}

		tx, parseErr := strconv.ParseUint(fields[9], 10, 64)
		if parseErr != nil {
			return 0, 0, parseErr
		}

		rxTotal += rx
		txTotal += tx
	}

	return rxTotal, txTotal, nil
}