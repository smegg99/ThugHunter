// core/scraper/status.go
package scraper

import (
	"sync"
	"time"
)

// RunMode differentiates between scraping and refresh-only runs.
type RunMode string

const (
	RunModeScrape   RunMode = "scrape"
	RunModeRefresh  RunMode = "refresh"
	RunModeRegister RunMode = "register"
)

// EventType identifies the kind of scraper run event.
type EventType string

const (
	EventRunStarted       EventType = "run_started"
	EventRunCompleted     EventType = "run_completed"
	EventQueryStarted     EventType = "query_started"
	EventQueryCompleted   EventType = "query_completed"
	EventQueryFailed      EventType = "query_failed"
	EventCreditsLow       EventType = "credits_low"
	EventAccountCreated   EventType = "account_created"
	EventCreditsRefreshed EventType = "credits_refreshed"
	EventAccountRefreshed EventType = "account_refreshed"
	EventAgentStarted     EventType = "agent_started"
	EventAgentStopped     EventType = "agent_stopped"
	EventAgentError       EventType = "agent_error"
	EventStatusUpdate     EventType = "status_update"
	EventRunSummary       EventType = "run_summary"
)

// RunEvent represents a single event emitted during a scraper run.
type RunEvent struct {
	Type      EventType      `json:"type"`
	Agent     string         `json:"agent,omitempty"`
	Message   string         `json:"message"`
	Timestamp time.Time      `json:"timestamp"`
	Data      map[string]any `json:"data,omitempty"`
}

// RunProgress is a snapshot of the current run state.
type RunProgress struct {
	Running             bool        `json:"running"`
	Mode                RunMode     `json:"mode"`
	TotalQueries        int         `json:"total_queries"`
	CompletedQueries    int         `json:"completed_queries"`
	EmptyQueries        int         `json:"empty_queries"`
	TotalHosts          int         `json:"total_hosts"`
	TotalAccounts       int         `json:"total_accounts"`
	RefreshedAccounts   int         `json:"refreshed_accounts"`
	FailedAccounts      int         `json:"failed_accounts"`
	RemainingAccounts   int         `json:"remaining_accounts"`
	CreatedAccounts     int         `json:"created_accounts"`
	FailedRegistrations int         `json:"failed_registrations"`
	TargetAccounts      int         `json:"target_accounts"`
	DurationSecs        int         `json:"duration_secs"`
	ActiveAgents        int         `json:"active_agents"`
	TotalAgents         int         `json:"total_agents"`
	TotalCredits        int         `json:"total_credits"`
	PossibleQueries     int         `json:"possible_queries"`
	UsedCredits         int         `json:"used_credits"`
	AccountsExhausted   bool        `json:"accounts_exhausted"`
	StartedAt           *time.Time  `json:"started_at,omitempty"`
	Agents              []AgentInfo `json:"agents"`
}

// RegisterOpts configures the account registration run.
type RegisterOpts struct {
	TargetAccounts int `json:"target_accounts"` // 0 = unlimited
	DurationSecs   int `json:"duration_secs"`   // 0 = no time limit
}

// RunSummary captures the final statistics of a completed run.
type RunSummary struct {
	Mode                RunMode   `json:"mode"`
	StoppedEarly        bool      `json:"stopped_early"`
	AccountsExhausted   bool      `json:"accounts_exhausted"`
	StartedAt           time.Time `json:"started_at"`
	FinishedAt          time.Time `json:"finished_at"`
	DurationSecs        float64   `json:"duration_secs"`
	TotalQueries        int       `json:"total_queries"`
	CompletedQueries    int       `json:"completed_queries"`
	EmptyQueries        int       `json:"empty_queries"`
	TotalHosts          int       `json:"total_hosts"`
	TotalAccounts       int       `json:"total_accounts"`
	RefreshedAccounts   int       `json:"refreshed_accounts"`
	FailedAccounts      int       `json:"failed_accounts"`
	CreatedAccounts     int       `json:"created_accounts"`
	FailedRegistrations int       `json:"failed_registrations"`
	TargetAccounts      int       `json:"target_accounts"`
	MaxDurationSecs     int       `json:"max_duration_secs"`
	TotalCredits        int       `json:"total_credits"`
	PossibleQueries     int       `json:"possible_queries"`
	UsedCredits         int       `json:"used_credits"`
}

// AgentInfo describes the current state of an individual agent.
type AgentInfo struct {
	Name             string      `json:"name"`
	Status           AgentStatus `json:"status"`
	StatusText       string      `json:"status_text,omitempty"`
	Account          string      `json:"account,omitempty"`
	Credits          uint        `json:"credits"`
	EstimatedCredits int         `json:"estimated_credits"`
}

// eventBus fans out RunEvents to all subscribers.
type eventBus struct {
	mu          sync.RWMutex
	subscribers map[int]chan RunEvent
	nextID      int
}

func newEventBus() *eventBus {
	return &eventBus{
		subscribers: make(map[int]chan RunEvent),
	}
}

// subscribe registers a new subscriber and returns its ID and receive channel.
func (b *eventBus) subscribe(bufSize int) (int, <-chan RunEvent) {
	b.mu.Lock()
	defer b.mu.Unlock()

	id := b.nextID
	b.nextID++
	ch := make(chan RunEvent, bufSize)
	b.subscribers[id] = ch
	return id, ch
}

// unsubscribe removes a subscriber and closes its channel.
func (b *eventBus) unsubscribe(id int) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if ch, ok := b.subscribers[id]; ok {
		close(ch)
		delete(b.subscribers, id)
	}
}

// emit sends an event to all subscribers (dropped if buffer is full).
func (b *eventBus) emit(event RunEvent) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for _, ch := range b.subscribers {
		select {
		case ch <- event:
		default:
		}
	}
}

// Subscribe registers for real-time run events and returns the subscriber ID and channel.
func (s *Scraper) Subscribe(bufSize int) (int, <-chan RunEvent) {
	return s.events.subscribe(bufSize)
}

// Unsubscribe removes an event subscriber and closes its channel.
func (s *Scraper) Unsubscribe(id int) {
	s.events.unsubscribe(id)
}

// emitEvent is a convenience helper for publishing events from the scraper.
func (s *Scraper) emitEvent(typ EventType, agent, message string, data map[string]any) {
	s.events.emit(RunEvent{
		Type:      typ,
		Agent:     agent,
		Message:   message,
		Timestamp: time.Now(),
		Data:      data,
	})
}
