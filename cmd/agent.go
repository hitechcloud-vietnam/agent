package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/hitechcloud-vietnam/agent/pkg/config"
	"github.com/hitechcloud-vietnam/agent/pkg/cpu"
	"github.com/hitechcloud-vietnam/agent/pkg/disk"
	"github.com/hitechcloud-vietnam/agent/pkg/logger"
	"github.com/hitechcloud-vietnam/agent/pkg/memory"
	"github.com/hitechcloud-vietnam/agent/pkg/network"
)

type Payload struct {
	Load                 float64 `json:"load"`
	CPUUsage             float64 `json:"cpu_usage"`
	CPUCores             int     `json:"cpu_cores"`
	DiskTotal            string  `json:"disk_total"`
	DiskFree             string  `json:"disk_free"`
	DiskUsed             string  `json:"disk_used"`
	DiskRead             string  `json:"disk_read"`
	DiskWrite            string  `json:"disk_write"`
	DiskTPS              string  `json:"disk_tps"`
	IOWait               string  `json:"io_wait"`
	MemoryTotal          string  `json:"memory_total"`
	MemoryFree           string  `json:"memory_free"`
	MemoryUsed           string  `json:"memory_used"`
	NetworkUpstream      string  `json:"network_upstream"`
	NetworkDownstream    string  `json:"network_downstream"`
	NetworkTotalSent     string  `json:"network_total_sent"`
	NetworkTotalReceived string  `json:"network_total_received"`
}

func main() {
	cfg := config.GetConfig()
	appLogger, closer, err := logger.New(cfg)
	if err != nil {
		panic(err)
	}
	defer closer.Close()

	appLogger.Infof("agent started url=%s level=%s", cfg.Url, cfg.LogLevel)
	client := &http.Client{Timeout: 15 * time.Second}

	for {
		func() {
			defer func() {
				if recovered := recover(); recovered != nil {
					appLogger.Errorf("agent loop recovered from panic: %v", recovered)
				}
			}()

			cpuInfo := cpu.GetCPUInfo()
			diskInfo := disk.GetDiskInfo()
			memoryInfo := memory.GetMemoryInfo()
			networkInfo := network.GetNetworkInfo()
			payload := Payload{
				Load:                 cpuInfo.Load,
				CPUUsage:             cpuInfo.Usage,
				CPUCores:             cpuInfo.Cores,
				DiskTotal:            diskInfo.Total,
				DiskFree:             diskInfo.Free,
				DiskUsed:             diskInfo.Used,
				DiskRead:             diskInfo.ReadBytesPerSecond,
				DiskWrite:            diskInfo.WriteBytesPerSecond,
				DiskTPS:              diskInfo.TPS,
				IOWait:               diskInfo.IOWait,
				MemoryTotal:          memoryInfo.Total,
				MemoryFree:           memoryInfo.Free,
				MemoryUsed:           memoryInfo.Used,
				NetworkUpstream:      networkInfo.UpstreamBytesPerSecond,
				NetworkDownstream:    networkInfo.DownstreamBytesPerSecond,
				NetworkTotalSent:     networkInfo.TotalSent,
				NetworkTotalReceived: networkInfo.TotalReceived,
			}

			jsonPayload, err := json.Marshal(payload)
			if err != nil {
				appLogger.Errorf("failed to marshal payload: %v", err)
				return
			}

			req, err := http.NewRequest("POST", cfg.Url, bytes.NewBuffer(jsonPayload))
			if err != nil {
				appLogger.Errorf("failed to build request: %v", err)
				return
			}

			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Secret", cfg.Secret)

			resp, err := client.Do(req)
			if err != nil {
				appLogger.Errorf("failed to send metrics: %v", err)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode >= http.StatusBadRequest {
				responseBody, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
				appLogger.Warnf("metrics sent with non-success status=%s body=%s", resp.Status, string(responseBody))
				return
			}

			appLogger.Infof("metrics sent status=%s load=%.2f", resp.Status, payload.Load)
		}()

		time.Sleep(time.Minute)
	}
}
