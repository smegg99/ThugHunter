// services/scanner/events.go
package scanner

import "github.com/wailsapp/wails/v3/pkg/application"

const (
	EventScanStarted   = "scanner:scan_started"
	EventScanProgress  = "scanner:scan_progress"
	EventScanCompleted = "scanner:scan_completed"
	EventScanError     = "scanner:scan_error"
)

// emitEvent sends a Wails event from the service layer.
func emitEvent(name string, data any) {
	if app := application.Get(); app != nil {
		app.Event.Emit(name, data)
	}
}
