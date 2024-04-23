// Package psutil
// Date: 2024/3/6 10:55
// Author: Amu
// Description:
package psutil

import "testing"

func TestGetCPUPercent(t *testing.T) {
	cpuPercent, err := GetCPUPercent()
	t.Log(cpuPercent, err)
}

func TestGetMemInfo(t *testing.T) {
	memPercent, total, used, err := GetMemInfo()
	t.Log(memPercent, total, used, err)
}

func TestGetDiskPercent(t *testing.T) {
	devices := map[string]struct{}{"/dev/disk3s1s1": {}}
	diskMap, err := GetDiskInfo(devices)
	t.Log(diskMap, err)
	diskIoMap, errIo := GetDiskIO(devices)
	t.Log(diskIoMap, errIo)
}

func TestGetNetworkPercent(t *testing.T) {
	devices := map[string]struct{}{"en0": {}}
	netMap, err := GetNetworkIO(devices)
	t.Log(netMap, err)
}

func TestGetDiskIO(t *testing.T) {
	devices := map[string]struct{}{"disk0": {}}
	diskMap, err := GetDiskIO(devices)
	t.Log(diskMap, err)
}
