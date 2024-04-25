// Package service
// Date: 2024/3/6 13:34
// Author: Amu
// Description:
package service

import (
	"context"
	"fmt"
	"github.com/patrickmn/go-cache"
	"log/slog"
	"time"

	"github.com/amuluze/amprobe/pkg/psutil"
	"github.com/amuluze/amprobe/service/model"
	"github.com/amuluze/amutool/database"
	"github.com/amuluze/amutool/docker"
	"github.com/amuluze/amutool/timex"
)

type TimedTask struct {
	db               *database.DB
	manager          *docker.Manager
	devices          map[string]struct{}
	ethernet         map[string]struct{}
	ticker           timex.Ticker
	stopCh           chan struct{}
	cache            *cache.Cache
	notMonitorDocker bool
}

func NewTimedTask(conf *Config, db *database.DB) *TimedTask {
	interval := conf.Task.Interval
	tk := timex.NewTicker(time.Duration(interval) * time.Second)
	manager, err := docker.NewManager()
	if err != nil {
		return nil
	}

	dev := make(map[string]struct{})
	for _, d := range conf.Disk.Devices {
		dev[d] = struct{}{}
	}

	eth := make(map[string]struct{})
	for _, d := range conf.Ethernet.Names {
		eth[d] = struct{}{}
	}

	return &TimedTask{
		devices:          dev,
		ethernet:         eth,
		ticker:           tk,
		stopCh:           make(chan struct{}),
		db:               db,
		manager:          manager,
		cache:            cache.New(5*time.Minute, 60*time.Second),
		notMonitorDocker: conf.Task.NotMonitorDocker,
	}
}

func (a *TimedTask) Execute() {
	timestamp := time.Now()
	// 处理数组指标
	go a.host(timestamp)
	go a.cpu(timestamp)
	go a.memory(timestamp)
	go a.disk()
	go a.network(timestamp)

	if a.notMonitorDocker {
		// 处理 Docker 容器指标
		go a.container(timestamp)
		go func() {
			a.docker(timestamp)
			a.image(timestamp)
		}()
	}

	go a.clearOldRecord()
}

func (a *TimedTask) Run() {
	for {
		select {
		case <-a.ticker.Chan():
			go a.Execute()
		case <-a.stopCh:
			fmt.Println("task exit")
			return
		}
	}
}

func (a *TimedTask) Stop() {
	close(a.stopCh)
}

func (a *TimedTask) host(timestamp time.Time) {
	info, _ := psutil.GetSystemInfo()
	if err := a.db.Unscoped().Where("1 = 1").Delete(&model.Host{}).Error; err != nil {
		slog.Error("Error deleting host table", "error", err)
	}
	a.db.Model(&model.Host{}).Create(&model.Host{
		Timestamp:       timestamp,
		Uptime:          info.Uptime,
		Hostname:        info.Hostname,
		Os:              info.Os,
		Platform:        info.Platform,
		PlatformVersion: info.PlatformVersion,
		KernelVersion:   info.KernelVersion,
		KernelArch:      info.KernelArch,
	})
}

func (a *TimedTask) cpu(timestamp time.Time) {
	cpuPercent, _ := psutil.GetCPUPercent()
	a.db.Model(&model.CPU{}).Create(&model.CPU{
		Timestamp:  timestamp,
		CPUPercent: cpuPercent,
	})
}

func (a *TimedTask) memory(timestamp time.Time) {
	memPercent, memTotal, memUsed, _ := psutil.GetMemInfo()
	a.db.Model(&model.Memory{}).Create(&model.Memory{
		Timestamp:  timestamp,
		MemPercent: memPercent,
		MemTotal:   float64(memTotal),
		MemUsed:    float64(memUsed),
	})
}

func (a *TimedTask) disk() {
	diskMap, _ := psutil.GetDiskIO(a.devices)
	var diskInfos []model.Disk
	for device, state := range diskMap {
		disk := model.Disk{Device: device}
		disk.DiskRead = float64(state.Read)
		disk.DiskWrite = float64(state.Write)
		diskInfos = append(diskInfos, disk)
	}
	// check diskInfos is empty
	if len(diskInfos) == 0 {
		slog.Error("diskInfos is empty")
		return
	}
	a.db.Model(&model.Disk{}).Create(diskInfos)
}

func (a *TimedTask) network(timestamp time.Time) {
	netMap, _ := psutil.GetNetworkIO(a.ethernet)
	time.Sleep(1 * time.Second)
	netMapAfterSecond, _ := psutil.GetNetworkIO(a.ethernet)
	var netInfos []model.Net
	for eth, info := range netMap {
		for e, i := range netMapAfterSecond {
			if eth == e {
				net := model.Net{}
				net.Ethernet = eth
				net.NetSend = float64(i.Send - info.Send)
				net.NetRecv = float64(i.Recv - info.Recv)
				netInfos = append(netInfos, net)
			}
		}
	}
	a.db.Model(&model.Net{}).Create(netInfos)
}

func (a *TimedTask) container(timestamp time.Time) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	cs, err := a.manager.ListContainer(ctx)
	if err != nil {
		slog.Error("failed to list containers", "error", err)
		return
	}
	var containers []model.Container
	for _, info := range cs {
		var d model.Container
		d.Timestamp = timestamp
		d.ContainerID = info.ID[:6]
		d.Name = info.Name
		d.State = info.State
		d.Image = info.Image
		d.Uptime = info.Uptime
		d.IP = info.IP

		cpuPercent, err := a.manager.GetContainerCPU(ctx, info.ID[:6])
		if err != nil {
			slog.Error("failed to get container cpu", "error", err)
		}
		d.CPUPercent = cpuPercent

		memPercent, used, limit, err := a.manager.GetContainerMem(ctx, info.ID[:6])
		if err != nil {
			slog.Error("failed to get container mem", "error", err)
		}
		d.MemPercent = memPercent

		d.MemUsage = used
		d.MemLimit = limit
		if _, ok := a.cache.Get(info.Image); !ok {
			a.cache.Set(info.Image, 1, 2*time.Minute)
		} else {
			count, err := a.cache.IncrementInt(info.Image, 1)
			slog.Info("container image cache", "image", info.Image, "count", count, "error", err)
		}
		containers = append(containers, d)
	}
	if err := a.db.Unscoped().Where("1 = 1").Delete(&model.Container{}).Error; err != nil {
		slog.Error("failed to delete container", "error", err)
	}
	a.db.Model(&model.Container{}).Create(&containers)
}

func (a *TimedTask) docker(timestamp time.Time) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	dockerVersion, err := a.manager.Version(ctx)
	if err != nil {
		slog.Error("failed to get docker version", "error", err)
		return
	}
	if err := a.db.Unscoped().Where("1 = 1").Delete(&model.Docker{}).Error; err != nil {
		slog.Error("failed to delete docker container", "error", err)
	}
	a.db.Model(&model.Docker{}).Create(&model.Docker{
		Timestamp:     timestamp,
		DockerVersion: dockerVersion.DockerVersion,
		APIVersion:    dockerVersion.APIVersion,
		MinAPIVersion: dockerVersion.MinAPIVersion,
		GitCommit:     dockerVersion.GitCommit,
		GoVersion:     dockerVersion.GoVersion,
		Os:            dockerVersion.OS,
		Arch:          dockerVersion.Arch,
	})
}

func (a *TimedTask) image(timestamp time.Time) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	images, err := a.manager.ListImage(ctx)
	if err != nil {
		slog.Error("failed to get version", "error", err)
		return
	}
	var list model.Images
	duplicateImage := make(map[string]struct{})
	for _, im := range images {
		val, ok := a.cache.Get(im.Name + ":" + im.Tag)
		if !ok {
			slog.Error("failed to get image cache", "error", err)
			val = 0
		}
		if _, ok := duplicateImage[im.ID]; !ok {
			duplicateImage[im.ID] = struct{}{}
		} else {
			if im.Tag != "latest" {
				continue
			}
		}
		list = append(list, model.Image{
			Timestamp: timestamp,
			ImageID:   im.ID[7:19],
			Name:      im.Name,
			Number:    val.(int),
			Tag:       im.Tag,
			Created:   im.Created,
			Size:      im.Size,
		})
		a.cache.Delete(im.Name + ":" + im.Tag)
	}
	if err := a.db.Unscoped().Where("1 = 1").Delete(&model.Image{}).Error; err != nil {
		slog.Error("failed to delete image", "error", err)
	}
	a.db.Model(&model.Image{}).Create(&list)
}

func (a *TimedTask) clearOldRecord() {
	a.db.Where("timestamp < ?", time.Now().Add(-time.Minute*5)).Delete(&model.Host{})
	a.db.Where("timestamp < ?", time.Now().Add(-time.Minute*5)).Delete(&model.Container{})
	a.db.Where("timestamp < ?", time.Now().Add(-time.Minute*5)).Delete(&model.Image{})
	a.db.Where("timestamp < ?", time.Now().Add(-time.Minute*5)).Delete(&model.Docker{})
	a.db.Where("timestamp < ?", time.Now().Add(-time.Hour*24*2)).Delete(&model.CPU{})
	a.db.Where("timestamp < ?", time.Now().Add(-time.Hour*24*2)).Delete(&model.Memory{})
	a.db.Where("timestamp < ?", time.Now().Add(-time.Hour*24*2)).Delete(&model.Disk{})
	a.db.Where("timestamp < ?", time.Now().Add(-time.Hour*24*2)).Delete(&model.Net{})
}
