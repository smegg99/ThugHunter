// core/scanner/worker.go
package scanner

import (
	"context"
	"fmt"
	"strconv"
	"sync"

	"smegg.me/thughunter/common/logger"
	"smegg.me/thughunter/core/models"
)

// runWorker is the main loop for a single worker goroutine.
// It pulls hosts from the shared queue until it is closed or ctx is cancelled.
func runWorker(
	ctx context.Context,
	queue <-chan *models.Host,
	results chan<- HostResult,
	probes []Probe,
	sc scanConfig,
	stats *Stats,
) {
	for {
		select {
		case <-ctx.Done():
			return
		case host, ok := <-queue:
			if !ok {
				return
			}
			result := processHost(ctx, host, probes, sc, stats)
			select {
			case results <- result:
			case <-ctx.Done():
				return
			}
		}
	}
}

// processHost performs the full scan sequence for a single host.
func processHost(
	ctx context.Context,
	host *models.Host,
	probes []Probe,
	sc scanConfig,
	stats *Stats,
) HostResult {
	result := HostResult{Host: host}

	if ctx.Err() != nil {
		return result
	}

	logger.Debug().Str("ip", host.IP).Msg("processing host")

	result.Ping = pingHost(ctx, host.IP, sc, stats)

	if shouldSkipProbes(result.Ping, sc.pingMode) {
		logger.Debug().Str("ip", host.IP).Msg("ping failed in strict mode, skipping probes")
		return result
	}

	result.ServiceResults = probeServicesConcurrent(ctx, host, probes, sc, stats)
	logger.Debug().Str("ip", host.IP).Int("services", len(result.ServiceResults)).Msg("host processing complete")
	return result
}

// pingHost runs the ping step and records the outcome in stats.
func pingHost(ctx context.Context, ip string, sc scanConfig, stats *Stats) PingResult {
	result := ping(ctx, ip, sc.pingTimeout, sc.icmpPing)
	stats.addPing(result.Status)
	return result
}

// shouldSkipProbes returns true when ping failed and mode is strict.
func shouldSkipProbes(pr PingResult, pingMode PingMode) bool {
	return !pr.Alive && pingMode == PingModeStrict
}

// portJob pairs a probe with a specific port for concurrent execution.
type portJob struct {
	probe Probe
	port  int
}

// portsForProbe looks up ports from the host's Services map using the probe's
// ServiceLabel. Returns nil when the host has no services for this probe.
func portsForProbe(host *models.Host, probe Probe) []int {
	portStrs, ok := host.Services[probe.ServiceLabel()]
	if !ok || len(portStrs) == 0 {
		return nil
	}

	ports := make([]int, 0, len(portStrs))
	for _, s := range portStrs {
		p, err := strconv.Atoi(s)
		if err != nil {
			logger.Warn().
				Str("ip", host.IP).
				Str("label", probe.ServiceLabel()).
				Str("port", s).
				Msg("invalid port in host services map, skipping")
			continue
		}
		ports = append(ports, p)
	}
	return ports
}

// probeServicesConcurrent fans out one goroutine per service/port combination,
// collects results, and returns them. Each port probe runs concurrently within
// the host context.
func probeServicesConcurrent(ctx context.Context, host *models.Host, probes []Probe, sc scanConfig, stats *Stats) []ServiceResult {
	// Build list of all port jobs across all probes for this host.
	var jobs []portJob
	for _, probe := range probes {
		for _, port := range portsForProbe(host, probe) {
			jobs = append(jobs, portJob{probe: probe, port: port})
		}
	}
	if len(jobs) == 0 {
		return nil
	}

	logger.Debug().Str("ip", host.IP).Int("jobs", len(jobs)).Msg("probing services")

	// Pre-allocate results slice; each goroutine writes to its own index.
	results := make([]ServiceResult, len(jobs))
	var wg sync.WaitGroup

	for i, job := range jobs {
		if ctx.Err() != nil {
			break
		}
		wg.Add(1)
		go func(idx int, j portJob) {
			defer wg.Done()
			if ctx.Err() != nil {
				return
			}
			sr := j.probe.Check(ctx, host.IP, j.port, sc)
			results[idx] = sr
			stats.addProbe(sr.Status)
			logProbeResult(host.IP, j.probe.Name(), j.port, sr.Open)
		}(i, job)
	}
	wg.Wait()

	// Filter out zero-value entries from cancelled probes.
	filtered := make([]ServiceResult, 0, len(results))
	for _, sr := range results {
		if sr.Service != "" {
			filtered = append(filtered, sr)
		}
	}
	return filtered
}

// logProbeResult emits a debug log for a single port probe.
func logProbeResult(ip, service string, port int, open bool) {
	label := fmt.Sprintf("%s:%d", service, port)
	if open {
		logger.Debug().Str("ip", ip).Str("service", label).Msg("port open")
	} else {
		logger.Debug().Str("ip", ip).Str("service", label).Msg("port closed/error")
	}
}
