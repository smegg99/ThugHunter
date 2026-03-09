// core/agent/agent.go
package scraper

import (
	"fmt"
	"sync"

	"smegg.me/thughunter/common/config"
	"smegg.me/thughunter/common/logger"
	"smegg.me/thughunter/core/repositories"
)

type Scraper struct {
	mu     sync.Mutex
	agents map[string]*ScraperAgent
}

var (
	instance *Scraper
	initOnce sync.Once
)

func Initialize() error {
	var initErr error
	initOnce.Do(func() {
		logger.Info().Msg("initializing agent")

		cfg := config.Get()

		instance = &Scraper{
			agents: make(map[string]*ScraperAgent),
		}

		logger.Info().
			Int64("max_agents", cfg.Scraper.Agents.MaxAgents).
			Msg("agent initialized")
	})
	return initErr
}

func Get() *Scraper {
	return instance
}

func (c *Scraper) Shutdown() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	logger.Info().Int("total", len(c.agents)).Msg("shutting down agent")

	for name, agent := range c.agents {
		logger.Debug().Str("agent", name).Msg("closing agent")
		if err := agent.Close(); err != nil {
			logger.Error().Err(err).Str("agent", name).Msg("error closing agent")
		}
		delete(c.agents, name)
	}

	logger.Info().Msg("agent shut down")
	return nil
}

const maxPetnameAttempts = 100

func (c *Scraper) CreateAgent() (*ScraperAgent, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	cfg := config.Get()
	maxAgents := int(cfg.Scraper.Agents.MaxAgents)
	if len(c.agents) >= maxAgents {
		return nil, fmt.Errorf("%w (%d)", ErrMaxAgentsReached, maxAgents)
	}

	var name string
	for range maxPetnameAttempts {
		candidate := generatePetname()
		if _, exists := c.agents[candidate]; !exists {
			name = candidate
			break
		}
	}
	if name == "" {
		return nil, fmt.Errorf("%w after %d attempts", ErrNameGenerationFailed, maxPetnameAttempts)
	}

	agent := newScraperAgent(name)
	c.agents[name] = agent

	logger.Info().Str("agent", name).Int("total", len(c.agents)).Msg("agent created")
	return agent, nil
}

func (c *Scraper) DeleteAgent(name string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	agent, ok := c.agents[name]
	if !ok {
		return fmt.Errorf("%w: %q", ErrAgentNotFound, name)
	}

	if err := agent.Close(); err != nil {
		return fmt.Errorf("close agent %q: %w", name, err)
	}

	delete(c.agents, name)
	logger.Info().Str("agent", name).Int("total", len(c.agents)).Msg("agent deleted")
	return nil
}

func (c *Scraper) GetAgent(name string) *ScraperAgent {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.agents[name]
}

func (c *Scraper) ListAgents() []string {
	c.mu.Lock()
	defer c.mu.Unlock()

	names := make([]string, 0, len(c.agents))
	for name := range c.agents {
		names = append(names, name)
	}
	return names
}

func (c *Scraper) AgentCount() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.agents)
}

func (c *Scraper) UpdateCreditsAllAccounts() {
	accountRepo := repositories.GetAccountRepository()
	if accountRepo == nil {
		logger.Error().Msg("account repository is nil, cannot update credits for accounts")
		return
	}

	// accounts, err := accountRepo.ListAll()
	// if err != nil {
	// 	logger.Error().Err(err).Msg("failed to list accounts for credit update")
	// 	return
	// }
}
