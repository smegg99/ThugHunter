// core/crawler/models.go
package crawler

type AgentStatus string

const (
	AgentStatusIdle    AgentStatus = "idle"
	AgentStatusBusy    AgentStatus = "busy"
	AgentStatusStopped AgentStatus = "stopped"
	AgentStatusWorking AgentStatus = "working"
	AgentStatusError   AgentStatus = "error"
)

type CrawlerCredits struct {
	Current int `json:"current"`
	Total   int `json:"total"`
}
