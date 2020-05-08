// Package nri uses Docker API information and sampled containers and presents it in a format that is accepted
// by the New Relic Infrastructure Agent
package nri

import "github.com/newrelic/infra-integrations-sdk/data/metric"

var (
	metricCommandLine             = metricFunc("commandLine", metric.ATTRIBUTE)
	metricContainerImage          = metricFunc("image", metric.ATTRIBUTE)
	metricContainerImageName      = metricFunc("imageName", metric.ATTRIBUTE)
	metricContainerName           = metricFunc("name", metric.ATTRIBUTE)
	metricState                   = metricFunc("state", metric.ATTRIBUTE)
	metricStatus                  = metricFunc("status", metric.ATTRIBUTE)
	metricRestartCount            = metricFunc("restartCount", metric.GAUGE)
	metricCPUUsedCores            = metricFunc("cpuUsedCores", metric.GAUGE)
	metricCPUUsedCoresPercent     = metricFunc("cpuUsedCoresPercent", metric.GAUGE)
	metricCPULimitCores           = metricFunc("cpuLimitCores", metric.GAUGE)
	metricCPUPercent              = metricFunc("cpuPercent", metric.GAUGE)
	metricCPUKernelPercent        = metricFunc("cpuKernelPercent", metric.GAUGE)
	metricCPUUserPercent          = metricFunc("cpuUserPercent", metric.GAUGE)
	metricCPUThrottleTimeMS       = metricFunc("cpuThrottleTimeMs", metric.GAUGE)
	metricCPUThrottlePeriods      = metricFunc("cpuThrottlePeriods", metric.GAUGE)
	metricMemoryUsageBytes        = metricFunc("memoryUsageBytes", metric.GAUGE)
	metricMemoryCacheBytes        = metricFunc("memoryCacheBytes", metric.GAUGE)
	metricMemoryResidentSizeBytes = metricFunc("memoryResidentSizeBytes", metric.GAUGE)
	metricMemorySizeLimitBytes    = metricFunc("memorySizeLimitBytes", metric.GAUGE)
	metricMemoryUsageLimitPercent = metricFunc("memoryUsageLimitPercent", metric.GAUGE)
	metricIOReadCountPerSecond    = metricFunc("ioReadCountPerSecond", metric.RATE)
	metricIOWriteCountPerSecond   = metricFunc("ioWriteCountPerSecond", metric.RATE)
	metricIOReadBytesPerSecond    = metricFunc("ioReadBytesPerSecond", metric.RATE)
	metricIOWriteBytesPerSecond   = metricFunc("ioWriteBytesPerSecond", metric.RATE)
	metricIOTotalReadCount        = metricFunc("ioTotalReadCount", metric.GAUGE)
	metricIOTotalWriteCount       = metricFunc("ioTotalWriteCount", metric.GAUGE)
	metricIOTotalReadBytes        = metricFunc("ioTotalReadBytes", metric.GAUGE)
	metricIOTotalWriteBytes       = metricFunc("ioTotalWriteBytes", metric.GAUGE)
	metricIOTotalBytes            = metricFunc("ioTotalBytes", metric.GAUGE)
	metricThreadCount             = metricFunc("threadCount", metric.GAUGE)
	metricThreadCountLimit        = metricFunc("threadCountLimit", metric.GAUGE)
	metricRxBytes                 = metricFunc("networkRxBytes", metric.GAUGE)
	metricRxDropped               = metricFunc("networkRxDropped", metric.GAUGE)
	metricRxErrors                = metricFunc("networkRxErrors", metric.GAUGE)
	metricRxPackets               = metricFunc("networkRxPackets", metric.GAUGE)
	metricTxBytes                 = metricFunc("networkTxBytes", metric.GAUGE)
	metricTxDropped               = metricFunc("networkTxDropped", metric.GAUGE)
	metricTxErrors                = metricFunc("networkTxErrors", metric.GAUGE)
	metricTxPackets               = metricFunc("networkTxPackets", metric.GAUGE)
	metricRxBytesPerSecond        = metricFunc("networkRxBytesPerSecond", metric.RATE)
	metricRxDroppedPerSecond      = metricFunc("networkRxDroppedPerSecond", metric.RATE)
	metricRxErrorsPerSecond       = metricFunc("networkRxErrorsPerSecond", metric.RATE)
	metricRxPacketsPerSecond      = metricFunc("networkRxPacketsPerSecond", metric.RATE)
	metricTxBytesPerSecond        = metricFunc("networkTxBytesPerSecond", metric.RATE)
	metricTxDroppedPerSecond      = metricFunc("networkTxDroppedPerSecond", metric.RATE)
	metricTxErrorsPerSecond       = metricFunc("networkTxErrorsPerSecond", metric.RATE)
	metricTxPacketsPerSecond      = metricFunc("networkTxPacketsPerSecond", metric.RATE)
)

type entry struct {
	Name  string
	Type  metric.SourceType
	Value interface{}
}

func metricFunc(name string, sType metric.SourceType) func(interface{}) entry {
	return func(value interface{}) entry {
		return entry{Name: name, Type: sType, Value: value}
	}
}
