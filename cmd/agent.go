package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/hitechcloud-vietnam/agent/pkg/config"
	"github.com/hitechcloud-vietnam/agent/pkg/cpu"
	"github.com/hitechcloud-vietnam/agent/pkg/disk"
	"github.com/hitechcloud-vietnam/agent/pkg/logger"
	"github.com/hitechcloud-vietnam/agent/pkg/memory"
)

type Payload struct {
	Load        float64 `json:"load"`
	DiskTotal   string  `json:"disk_total"`
	DiskFree    string  `json:"disk_free"`
	DiskUsed    string  `json:"disk_used"`
	MemoryTotal string  `json:"memory_total"`
	MemoryFree  string  `json:"memory_free"`
	MemoryUsed  string  `json:"memory_used"`
}

func main() {
	cfg := config.GetConfig()
	appLogger, closer, err := logger.New(cfg)
	if err != nil {
		panic(err)
	}
	defer closer.Close()

	appLogger.Infof("agent started url=%s level=%s", cfg.Url, cfg.LogLevel)

	for {
		cpuInfo := cpu.GetCPUInfo()
		diskInfo := disk.GetDiskInfo()
		memoryInfo := memory.GetMemoryInfo()
		payload := Payload{
			Load:        cpuInfo.Load,
			DiskTotal:   diskInfo.Total,
			DiskFree:    diskInfo.Free,
			DiskUsed:    diskInfo.Used,
			MemoryTotal: memoryInfo.Total,
			MemoryFree:  memoryInfo.Free,
			MemoryUsed:  memoryInfo.Used,
		}
		jsonPayload, err := json.Marshal(payload)
		if err != nil {
			appLogger.Errorf("failed to marshal payload: %v", err)
			continue
		}
		req, err := http.NewRequest("POST", cfg.Url, bytes.NewBuffer(jsonPayload))
		if err != nil {
			appLogger.Errorf("failed to build request: %v", err)
			continue
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Secret", cfg.Secret)
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			appLogger.Errorf("failed to send metrics: %v", err)
			continue
		}
		if resp.StatusCode >= http.StatusBadRequest {
			appLogger.Warnf("metrics sent with non-success status=%s load=%.2f", resp.Status, payload.Load)
		} else {
			appLogger.Infof("metrics sent status=%s load=%.2f", resp.Status, payload.Load)
		}
		resp.Body.Close()
		time.Sleep(time.Minute)
	}
}
