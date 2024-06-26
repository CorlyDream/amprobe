// Package schema
// Date: 2024/3/6 13:20
// Author: Amu
// Description:
package schema

type Usage struct {
	Timestamp int64   `json:"timestamp"`
	Value     float64 `json:"value"`
}

type NetIO struct {
	Timestamp int64   `json:"timestamp"`
	BytesSent float64 `json:"bytes_sent"`
	BytesRecv float64 `json:"bytes_recv"`
}

type DiskIO struct {
	Timestamp int64   `json:"timestamp"`
	IORead    float64 `json:"io_read"`
	IOWrite   float64 `json:"io_write"`
}

type HostInfoReply struct {
	Timestamp       int64  `json:"timestamp"`
	Uptime          string `json:"uptime"`
	Hostname        string `json:"hostname"`
	OS              string `json:"os"`
	Platform        string `json:"platform"`
	PlatformVersion string `json:"platform_version"`
	KernelVersion   string `json:"kernel_version"`
	KernelArch      string `json:"kernel_arch"`
}

type CPUInfoReply struct {
	Percent float64 `json:"percent"`
}

type CPUUsageArgs struct {
	StartTime int64 `query:"start_time"`
	EndTime   int64 `query:"end_time"`
}

type CPUUsageReply struct {
	Data []Usage `json:"data"`
}

type MemoryInfoReply struct {
	Percent float64 `json:"percent"`
	Total   float64 `json:"total"`
	Used    float64 `json:"used"`
}

type MemoryUsageArgs struct {
	StartTime int64 `query:"start_time"`
	EndTime   int64 `query:"end_time"`
}

type MemoryUsageReply struct {
	Data []Usage `json:"data"`
}

type DiskInfo struct {
	Device  string  `json:"device"`
	Percent float64 `json:"percent"`
	Total   float64 `json:"total"`
	Used    float64 `json:"used"`
}

type DiskInfoReply struct {
	Info []DiskInfo `json:"info"`
}

type DiskUsageArgs struct {
	StartTime int64 `query:"start_time"`
	EndTime   int64 `query:"end_time"`
}

type DiskUsageReply struct {
	Device string   `json:"device"`
	Mountpoint string `json:"mountpoint"`
	Data   []DiskIO `json:"data"`
	Total   uint64	`json:"total"`
	Percent float64	`json:"percent"`
	Used    uint64	`json:"used"`
}

type NetworkUsageArgs struct {
	StartTime int64 `query:"start_time"`
	EndTime   int64 `query:"end_time"`
}

type NetworkUsageReply struct {
	Ethernet string  `json:"ethernet"`
	Data     []NetIO `json:"data"`
}
