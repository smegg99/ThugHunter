// services/monitor/cpu.go
package monitor

import (
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
)

// collectCPU gathers per-core and total CPU usage percentages.
func collectCPU() (CPUStats, error) {
	logical, err := cpu.Counts(true)
	if err != nil {
		return CPUStats{}, err
	}

	perCore, err := cpu.Percent(0, true)
	if err != nil {
		return CPUStats{}, err
	}

	total, err := cpu.Percent(0, false)
	if err != nil {
		return CPUStats{}, err
	}

	var totalPct float64
	if len(total) > 0 {
		totalPct = total[0]
	}

	cores := make([]CoreUsage, len(perCore))
	for i, pct := range perCore {
		cores[i] = CoreUsage{Index: i, Percent: pct}
	}

	return CPUStats{
		LogicalCores:    logical,
		TotalPercent:    totalPct,
		Cores:           cores,
		SampledAtUnixMs: time.Now().UnixMilli(),
	}, nil
}
