// services/monitor/monitor.go
package monitor

type Service struct{}

const (
	EventMonitorCPU    = "monitor:cpu"
	EventMonitorRAM    = "monitor:ram"
	EventMonitorSwap   = "monitor:swap"
	EventMonitorSystem = "monitor:system"
)

type CoreUsage struct {
	Index   int     `json:"index"`
	Percent float64 `json:"percent"`
}

type CPUStats struct {
	LogicalCores    int         `json:"logicalCores"`
	TotalPercent    float64     `json:"totalPercent"`
	Cores           []CoreUsage `json:"cores"`
	SampledAtUnixMs int64       `json:"sampledAtUnixMs"`
}

type RAMStats struct {
	Total           uint64  `json:"total"`
	Used            uint64  `json:"used"`
	Available       uint64  `json:"available"`
	Free            uint64  `json:"free"`
	UsedPercent     float64 `json:"usedPercent"`
	SampledAtUnixMs int64   `json:"sampledAtUnixMs"`
}

type SwapStats struct {
	Total           uint64  `json:"total"`
	Used            uint64  `json:"used"`
	Free            uint64  `json:"free"`
	UsedPercent     float64 `json:"usedPercent"`
	SampledAtUnixMs int64   `json:"sampledAtUnixMs"`
}

type SystemSnapshot struct {
	CPU             CPUStats  `json:"cpu"`
	RAM             RAMStats  `json:"ram"`
	Swap            SwapStats `json:"swap"`
	IntervalMs      int64     `json:"intervalMs"`
	SampledAtUnixMs int64     `json:"sampledAtUnixMs"`
}

// GetCPU returns a single CPU usage snapshot.
func (s *Service) GetCPU() (CPUStats, error) {
	return collectCPU()
}

// GetRAM returns a single RAM usage snapshot.
func (s *Service) GetRAM() (RAMStats, error) {
	return collectRAM()
}

// GetSwap returns a single swap usage snapshot.
func (s *Service) GetSwap() (SwapStats, error) {
	return collectSwap()
}

// GetSnapshot returns a combined CPU + RAM + Swap snapshot.
func (s *Service) GetSnapshot() (*SystemSnapshot, error) {
	cpuStats, err := collectCPU()
	if err != nil {
		return nil, err
	}
	ramStats, err := collectRAM()
	if err != nil {
		return nil, err
	}
	swapStats, err := collectSwap()
	if err != nil {
		return nil, err
	}

	return &SystemSnapshot{
		CPU:  cpuStats,
		RAM:  ramStats,
		Swap: swapStats,
	}, nil
}

// StartPolling begins emitting monitor events at the given interval (ms).
// A minimum interval of 500 ms is enforced.
func (s *Service) StartPolling(intervalMs int64) {
	if intervalMs < 500 {
		intervalMs = 500
	}
	startPolling(intervalMs)
}

// StopPolling stops the background event emission.
func (s *Service) StopPolling() {
	stopPolling()
}
