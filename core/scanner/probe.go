// core/scanner/probe.go
package scanner

import "context"

// Probe defines the interface that every service module must implement.
type Probe interface {
	// Name returns the human-readable service name.
	Name() string

	// ServiceLabel returns the key used in the host's Services map
	// (e.g. "VNC"). Ports are resolved dynamically from the host.
	ServiceLabel() string

	// Check probes a single host:port and returns the result.
	// The context is used for cancellation of in-flight network operations.
	Check(ctx context.Context, ip string, port int, sc scanConfig) ServiceResult
}
