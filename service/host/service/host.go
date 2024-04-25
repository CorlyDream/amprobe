// Package service
// Date: 2024/3/6 12:53
// Author: Amu
// Description:
package service

import (
	"context"
	"sort"

	"github.com/amuluze/amprobe/pkg/psutil"
	"github.com/amuluze/amprobe/service/host/repository"
	"github.com/amuluze/amprobe/service/schema"
	"github.com/google/wire"
)

var HostServiceSet = wire.NewSet(NewHostService, wire.Bind(new(IHostService), new(*HostService)))

type IHostService interface {
	HostInfo(ctx context.Context) (schema.HostInfoReply, error)
	CPUInfo(ctx context.Context) (schema.CPUInfoReply, error)
	CPUUsage(ctx context.Context, args schema.CPUUsageArgs) (schema.CPUUsageReply, error)
	MemInfo(ctx context.Context) (schema.MemoryInfoReply, error)
	MemUsage(ctx context.Context, args schema.MemoryUsageArgs) (schema.MemoryUsageReply, error)
	DiskUsage(ctx context.Context, args schema.DiskUsageArgs) (schema.DiskUsageReply, error)
	DiskUsages(ctx context.Context, args schema.DiskUsageArgs) ([]schema.DiskUsageReply, error)
	NetUsage(ctx context.Context, args schema.NetworkUsageArgs) ([]schema.NetworkUsageReply, error)
}

type HostService struct {
	HostRepo repository.IHostRepo
}

func NewHostService(hostRepo repository.IHostRepo) *HostService {
	return &HostService{HostRepo: hostRepo}
}

func (h HostService) HostInfo(ctx context.Context) (schema.HostInfoReply, error) {
	info, err := h.HostRepo.HostInfo(ctx)
	if err != nil {
		return schema.HostInfoReply{}, err
	}
	return schema.HostInfoReply{
		Timestamp:       info.Timestamp.Unix(),
		Uptime:          info.Uptime,
		Hostname:        info.Hostname,
		OS:              info.Os,
		Platform:        info.Platform,
		PlatformVersion: info.PlatformVersion,
		KernelVersion:   info.KernelVersion,
		KernelArch:      info.KernelArch,
	}, err
}

func (h HostService) CPUInfo(ctx context.Context) (schema.CPUInfoReply, error) {
	cpuInfo, err := h.HostRepo.CPUInfo(ctx)
	if err != nil {
		return schema.CPUInfoReply{}, err
	}
	return schema.CPUInfoReply{Percent: cpuInfo.CPUPercent}, nil
}

func (h HostService) CPUUsage(ctx context.Context, args schema.CPUUsageArgs) (schema.CPUUsageReply, error) {
	mHosts, err := h.HostRepo.CPUUsage(ctx, args)
	if err != nil {
		return schema.CPUUsageReply{}, err
	}
	var list []schema.Usage
	for _, item := range mHosts {
		list = append(list, schema.Usage{
			Timestamp: item.Timestamp.Unix(),
			Value:     item.CPUPercent,
		})
	}
	return schema.CPUUsageReply{Data: list}, nil
}

func (h HostService) MemInfo(ctx context.Context) (schema.MemoryInfoReply, error) {
	memInfo, err := h.HostRepo.MemInfo(ctx)
	if err != nil {
		return schema.MemoryInfoReply{}, err
	}
	return schema.MemoryInfoReply{Percent: memInfo.MemPercent, Total: memInfo.MemTotal, Used: memInfo.MemUsed}, nil
}

func (h HostService) MemUsage(ctx context.Context, args schema.MemoryUsageArgs) (schema.MemoryUsageReply, error) {
	memInfos, err := h.HostRepo.MemUsage(ctx, args)
	if err != nil {
		return schema.MemoryUsageReply{}, err
	}
	var list []schema.Usage
	for _, item := range memInfos {
		list = append(list, schema.Usage{
			Timestamp: item.Timestamp.Unix(),
			Value:     item.MemPercent,
		})
	}
	return schema.MemoryUsageReply{Data: list}, nil
}

func (h HostService) DiskUsages(ctx context.Context, args schema.DiskUsageArgs) ([]schema.DiskUsageReply, error) {

	diskUsages, err := h.HostRepo.DiskUsage(ctx, args)
	if err != nil {
		return []schema.DiskUsageReply{}, err
	}

	diskMap := make(map[string][]schema.DiskIO)
	devices := make(map[string]struct{})
	// add serias data
	for _, item := range diskUsages {
		device := item.Device
		// 获取当前设备的切片
		diskIOs, ok := diskMap[device]
		if !ok {
			// 如果设备还没有在 map 中，创建一个新的切片
			diskIOs = []schema.DiskIO{}
			devices[device] = struct{}{}
		}
		// 将新的 DiskIO 添加到切片中
		diskIOs = append(diskIOs, schema.DiskIO{
			Timestamp: item.CreatedAt.Unix(),
			IORead:    item.DiskRead,
			IOWrite:   item.DiskWrite,
		})
		// 将更新后的切片放回 map 中
		diskMap[device] = diskIOs
	}
	diskInfos,_ := psutil.GetDiskInfo(devices)
	var list []schema.DiskUsageReply
	for device, diskIOs := range diskMap {
		list = append(list, schema.DiskUsageReply{
			Device: device,
			Data:   diskIOs,
			Mountpoint: diskInfos[device].Mountpoint,
			Total: diskInfos[device].Total,
			Percent: diskInfos[device].Percent,
			Used: diskInfos[device].Used,
		})
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].Device < list[j].Device
	})
	return list, nil
}

func (h HostService) DiskUsage(ctx context.Context, args schema.DiskUsageArgs) (schema.DiskUsageReply, error) {
	diskInfos, err := h.HostRepo.DiskUsage(ctx, args)
	if err != nil {
		return schema.DiskUsageReply{}, err
	}

	mDisk := make([]schema.DiskIO, 0)
	device := ""
	for _, item := range diskInfos {
		device = item.Device
		mDisk = append(mDisk, schema.DiskIO{
			Timestamp: item.CreatedAt.Unix(),
			IORead:    item.DiskRead,
			IOWrite:   item.DiskWrite,
		})
	}
	return schema.DiskUsageReply{Device: device, Data: mDisk}, nil
}

func (h HostService) NetUsage(ctx context.Context, args schema.NetworkUsageArgs) ([]schema.NetworkUsageReply, error) {
	netInfos, err := h.HostRepo.NetUsage(ctx, args)
	if err != nil {
		return []schema.NetworkUsageReply{}, err
	}
	netMap := make(map[string]schema.NetworkUsageReply)
	for _, item := range netInfos {
		usage, ok := netMap[item.Ethernet]
		if !ok {
			usage = schema.NetworkUsageReply{Ethernet: item.Ethernet}
			usage.Data = make([]schema.NetIO, 100)
			netMap[item.Ethernet] = usage
		}
		usage.Data = append(usage.Data, schema.NetIO{
			Timestamp: item.CreatedAt.Unix(),
			BytesSent: item.NetSend,
			BytesRecv: item.NetRecv,
		})
	}
	list := make([]schema.NetworkUsageReply, 1)
	for _, item := range netMap {
		list = append(list, item)
	}
	return list, nil
}
