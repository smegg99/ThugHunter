// services/scraper/agents.go
package scraper

import (
	"fmt"

	corescraper "smegg.me/thughunter/core/scraper"
)

// ListAgents returns information about all active agents.
func (s *Service) ListAgents() []corescraper.AgentInfo {
	sc := corescraper.Get()
	if sc == nil {
		return nil
	}

	names := sc.ListAgents()
	infos := make([]corescraper.AgentInfo, 0, len(names))
	for _, name := range names {
		if a := sc.GetAgent(name); a != nil {
			infos = append(infos, a.Info())
		}
	}
	return infos
}

// GetAgent returns information about a single agent by name.
func (s *Service) GetAgent(name string) (*corescraper.AgentInfo, error) {
	sc := corescraper.Get()
	if sc == nil {
		return nil, fmt.Errorf("scraper not initialized")
	}

	a := sc.GetAgent(name)
	if a == nil {
		return nil, corescraper.ErrAgentNotFound
	}
	info := a.Info()
	return &info, nil
}

// AgentCount returns the number of active agents.
func (s *Service) AgentCount() int {
	sc := corescraper.Get()
	if sc == nil {
		return 0
	}
	return sc.AgentCount()
}
