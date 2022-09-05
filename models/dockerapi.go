package models

type Stats struct {
	CPUStats    CPUStats `json:"cpu_stats"`
	PreCPUStats CPUStats `json:"precpu_stats"`
}

type CPUStats struct {
	CPUUsage       CPUUsage `json:"cpu_usage"`
	SystemCPUUsage int64    `json:"system_cpu_usage"`
}

type CPUUsage struct {
	TotalUsage int64 `json:"total_usage"`
}
