package raw

import (
	"time"

	"github.com/docker/docker/api/types"
)

// Metrics holds containers raw metric values as they are extracted from the system
type Metrics struct {
	Time        time.Time
	ContainerID string
	Memory      Memory
	Network     Network
	CPU         CPU
	Pids        Pids
	Blkio       Blkio
}

// Memory usage snapshot
type Memory struct {
	UsageLimit uint64
	Cache      uint64
	RSS        uint64
	SwapUsage  uint64
	FuzzUsage  uint64
}

// CPU usage snapshot
type CPU struct {
	TotalUsage        uint64
	UsageInUsermode   uint64
	UsageInKernelmode uint64
	PercpuUsage       []uint64
	ThrottledPeriods  uint64
	ThrottledTimeNS   uint64
	SystemUsage       uint64
	OnlineCPUs        uint
}

// Pids inside the container
type Pids struct {
	Current uint64
	Limit   uint64
}

// Blkio stores multiple entries of the Block I/O stats
type Blkio struct {
	IoServiceBytesRecursive []BlkioEntry
	IoServicedRecursive     []BlkioEntry
}

// BlkioEntry stores basic information of a simple blkio operation
type BlkioEntry struct {
	Op    string
	Value uint64
}

// Network transmission and receive metrics
type Network struct {
	RxBytes   int64
	RxDropped int64
	RxErrors  int64
	RxPackets int64
	TxBytes   int64
	TxDropped int64
	TxErrors  int64
	TxPackets int64
}

// Fetcher is the minimal abstraction of any raw metrics fetcher implementation
type Fetcher interface {
	Fetch(types.ContainerJSON) (Metrics, error)
}
