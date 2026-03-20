// services/scraper/events.go
package scraper

import (
	"github.com/wailsapp/wails/v3/pkg/application"
	corescraper "smegg.me/thughunter/core/scraper"
)

// Wails event names forwarded from the core event bus during runs.
const (
	EventRunStarted       = "scraper:run_started"
	EventRunCompleted     = "scraper:run_completed"
	EventQueryStarted     = "scraper:query_started"
	EventQueryCompleted   = "scraper:query_completed"
	EventQueryFailed      = "scraper:query_failed"
	EventCreditsLow       = "scraper:credits_low"
	EventAccountCreated   = "scraper:account_created"
	EventCreditsRefreshed = "scraper:credits_refreshed"
	EventAccountRefreshed = "scraper:account_refreshed"
	EventAgentStarted     = "scraper:agent_started"
	EventAgentStopped     = "scraper:agent_stopped"
	EventAgentError       = "scraper:agent_error"
	EventStatusUpdate     = "scraper:status_update"
	EventRunSummary       = "scraper:run_summary"
	EventProgress         = "scraper:progress"
)

// Service-level events emitted directly from service methods,
// independent of whether a run is active.
const (
	EventAgentStatusChanged = "scraper:service:agent_status_changed"
	EventAgentsChanged      = "scraper:service:agents_changed"
	EventAccountsChanged    = "scraper:service:accounts_changed"
	EventRunStateChanged    = "scraper:service:run_state_changed"
	EventStopping           = "scraper:service:stopping"
	EventShutdown           = "scraper:service:shutdown"
)

// emitServiceEvent sends a Wails event directly from the service layer.
func emitServiceEvent(name string, data any) {
	if app := application.Get(); app != nil {
		app.Event.Emit(name, data)
	}
}

// coreToWailsEvent maps core event types to wails event names.
var coreToWailsEvent = map[corescraper.EventType]string{
	corescraper.EventRunStarted:       EventRunStarted,
	corescraper.EventRunCompleted:     EventRunCompleted,
	corescraper.EventQueryStarted:     EventQueryStarted,
	corescraper.EventQueryCompleted:   EventQueryCompleted,
	corescraper.EventQueryFailed:      EventQueryFailed,
	corescraper.EventCreditsLow:       EventCreditsLow,
	corescraper.EventAccountCreated:   EventAccountCreated,
	corescraper.EventCreditsRefreshed: EventCreditsRefreshed,
	corescraper.EventAccountRefreshed: EventAccountRefreshed,
	corescraper.EventAgentStarted:     EventAgentStarted,
	corescraper.EventAgentStopped:     EventAgentStopped,
	corescraper.EventAgentError:       EventAgentError,
	corescraper.EventStatusUpdate:     EventStatusUpdate,
	corescraper.EventRunSummary:       EventRunSummary,
}

// startEventBridge subscribes to core events and forwards them as Wails events.
func startEventBridge(sc *corescraper.Scraper) int {
	id, ch := sc.Subscribe(512)

	go func() {
		app := application.Get()

		for ev := range ch {
			wailsName, ok := coreToWailsEvent[ev.Type]
			if !ok {
				continue
			}

			// Emit the raw core event for all types except those
			// that are re-emitted with extracted data below.
			if ev.Type != corescraper.EventRunSummary {
				app.Event.Emit(wailsName, ev)
			}

			// Emit derived service-level events based on core event type.
			switch ev.Type {
			case corescraper.EventAgentStarted,
				corescraper.EventAgentStopped,
				corescraper.EventAgentError:
				emitServiceEvent(EventAgentStatusChanged, ev.Agent)
				emitServiceEvent(EventAgentsChanged, nil)
				app.Event.Emit(EventProgress, sc.Progress())

			case corescraper.EventStatusUpdate:
				prog := sc.Progress()
				if info, ok := ev.Data["agent_info"]; ok {
					if agentInfo, ok := info.(corescraper.AgentInfo); ok {
						for i := range prog.Agents {
							if prog.Agents[i].Name == agentInfo.Name {
								prog.Agents[i] = agentInfo
								break
							}
						}
					}
				}
				app.Event.Emit(EventProgress, prog)

			case corescraper.EventQueryCompleted,
				corescraper.EventQueryFailed:
				emitServiceEvent(EventAccountsChanged, nil)
				app.Event.Emit(EventProgress, sc.Progress())

			case corescraper.EventAccountCreated:
				emitServiceEvent(EventAccountsChanged, nil)
				emitServiceEvent(EventAgentsChanged, nil)
				app.Event.Emit(EventProgress, sc.Progress())

			case corescraper.EventAccountRefreshed:
				emitServiceEvent(EventAccountsChanged, nil)
				app.Event.Emit(EventProgress, sc.Progress())

			case corescraper.EventCreditsRefreshed:
				emitServiceEvent(EventAccountsChanged, nil)

			case corescraper.EventRunSummary:
				app.Event.Emit(EventRunSummary, ev.Data["summary"])
			}
		}
	}()

	return id
}

// stopEventBridge unsubscribes the event bridge goroutine.
func stopEventBridge(sc *corescraper.Scraper, id int) {
	sc.Unsubscribe(id)
}
