// core/scanner/types.go
package scanner

import (
	"time"

	"smegg.me/thughunter/core/models"
)

// Status represents the outcome of a ping or probe.
type Status string

const (
	StatusOK      Status = "ok"
	StatusTimeout Status = "timeout"
	StatusError   Status = "error"
	StatusSkipped Status = "skipped"
)

// PingMode controls how the scanner treats hosts that don't respond to ping.
type PingMode string

const (
	PingModeSoft   PingMode = "soft"   // run all probes regardless of ping result
	PingModeStrict PingMode = "strict" // skip probes when ping fails
)

// ScreenshotStage tracks the lifecycle phase of the screenshot pass.
type ScreenshotStage int32

const (
	ScreenshotNotStarted ScreenshotStage = 0
	ScreenshotRunning    ScreenshotStage = 1
	ScreenshotDone       ScreenshotStage = 2
)

// PingResult holds the outcome of an ICMP ping attempt.
type PingResult struct {
	Alive   bool
	Latency time.Duration
	Status  Status
	Error   string
}

// ServiceResult holds the outcome of probing one service on one port.
type ServiceResult struct {
	Service models.ServiceType
	Port    int
	Open    bool
	Latency time.Duration
	Status  Status
	Error   string
	Detail  any // service-specific payload (e.g. *VNCDetail)
}

// VNCDetail carries VNC-specific probe information.
type VNCDetail struct {
	RFBVersion string
	AuthType   models.VNCAuthType
	NoAuth     bool
	Screenshot []byte // captured during scan when NoAuth is true
}

// HostResult aggregates the full scan outcome for a single host.
type HostResult struct {
	Host           *models.Host
	Ping           PingResult
	ServiceResults []ServiceResult
}
