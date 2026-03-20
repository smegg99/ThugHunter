// services/monitor/memory.go
package monitor

import (
	"time"

	"github.com/shirou/gopsutil/v4/mem"
)

// collectRAM gathers physical memory statistics.
func collectRAM() (RAMStats, error) {
	v, err := mem.VirtualMemory()
	if err != nil {
		return RAMStats{}, err
	}

	return RAMStats{
		Total:           v.Total,
		Used:            v.Used,
		Available:       v.Available,
		Free:            v.Free,
		UsedPercent:     v.UsedPercent,
		SampledAtUnixMs: time.Now().UnixMilli(),
	}, nil
}

// collectSwap gathers swap/page-file statistics.
func collectSwap() (SwapStats, error) {
	s, err := mem.SwapMemory()
	if err != nil {
		return SwapStats{}, err
	}

	return SwapStats{
		Total:           s.Total,
		Used:            s.Used,
		Free:            s.Free,
		UsedPercent:     s.UsedPercent,
		SampledAtUnixMs: time.Now().UnixMilli(),
	}, nil
}
