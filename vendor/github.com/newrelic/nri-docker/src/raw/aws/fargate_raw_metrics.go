package aws

import (
	"time"

	"github.com/newrelic/infra-integrations-sdk/log"

	"github.com/newrelic/nri-docker/src/raw"
)

func fargateRawMetrics(fargateStats FargateStats) map[string]*raw.Metrics {
	rawMetrics := make(map[string]*raw.Metrics, len(fargateStats))
	now := time.Now()

	for containerID, stats := range fargateStats {
		if stats == nil {
			log.Debug("did not find container stats for %s, skipping", containerID)
			continue
		}
		rawMetrics[containerID] = &raw.Metrics{
			Time:        now,
			ContainerID: containerID,
			Memory: raw.Memory{
				UsageLimit: stats.MemoryStats.Limit,
				Cache:      stats.MemoryStats.Stats["cache"],
				RSS:        stats.MemoryStats.Stats["rss"],
				SwapUsage:  0,
				FuzzUsage:  0,
			},
			Network: raw.Network{},
			CPU: raw.CPU{
				TotalUsage:        stats.CPUStats.CPUUsage.TotalUsage,
				UsageInUsermode:   stats.CPUStats.CPUUsage.UsageInUsermode,
				UsageInKernelmode: stats.CPUStats.CPUUsage.UsageInKernelmode,
				PercpuUsage:       stats.CPUStats.CPUUsage.PercpuUsage,
				ThrottledPeriods:  stats.CPUStats.ThrottlingData.ThrottledPeriods,
				ThrottledTimeNS:   stats.CPUStats.ThrottlingData.ThrottledTime,
				SystemUsage:       stats.CPUStats.SystemUsage,
				OnlineCPUs:        uint(stats.CPUStats.OnlineCPUs),
			},
			Pids: raw.Pids{
				Current: stats.PidsStats.Current,
				Limit:   stats.PidsStats.Limit,
			},
			Blkio: raw.Blkio{},
		}

		for _, s := range stats.BlkioStats.IoServiceBytesRecursive {
			entry := raw.BlkioEntry{Op: s.Op, Value: s.Value}
			rawMetrics[containerID].Blkio.IoServiceBytesRecursive = append(
				rawMetrics[containerID].Blkio.IoServiceBytesRecursive,
				entry,
			)
		}

		for _, s := range stats.BlkioStats.IoServicedRecursive {
			entry := raw.BlkioEntry{Op: s.Op, Value: s.Value}
			rawMetrics[containerID].Blkio.IoServicedRecursive = append(
				rawMetrics[containerID].Blkio.IoServicedRecursive,
				entry,
			)
		}
	}

	return rawMetrics
}
