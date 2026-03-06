// core/crawler/crawler.go
package crawler

import (
	"context"
	"fmt"
	"sync"

	"smegg.me/thughunter/core/unrevealed"

	"smegg.me/thughunter/common/config"
	"smegg.me/thughunter/common/logger"
)

type Crawler struct {
	mu      sync.Mutex
	browser *unrevealed.Browser
	agents  map[int]*CrawlerAgent
	nextID  int
}

var (
	instance *Crawler
	initOnce sync.Once
)

func Initialize() error {
	var initErr error
	initOnce.Do(func() {
		logger.Info().Msg("initializing crawler")

		cfg := config.Get()

		browser, err := unrevealed.New(context.Background(), unrevealed.Config{
			ChromePath: cfg.Crawler.BrowserBinaryPath,
			// VirtualDisplay: cfg.Crawler.VirtualDisplay,
			Headless: false,
		})
		if err != nil {
			initErr = fmt.Errorf("launch browser: %w", err)
			return
		}

		logger.Debug().Msg("browser launched via unrevealed")

		instance = &Crawler{
			browser: browser,
			agents:  make(map[int]*CrawlerAgent),
			nextID:  1,
		}

		logger.Info().
			Int64("max_agents", cfg.Crawler.Agents.MaxAgents).
			Msg("crawler initialized")
	})
	return initErr
}

func Get() *Crawler {
	return instance
}

func (c *Crawler) Shutdown() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	logger.Info().Int("total", len(c.agents)).Msg("shutting down crawler")

	for id, agent := range c.agents {
		logger.Debug().Int("agent_id", id).Msg("closing agent")
		if err := agent.Close(); err != nil {
			logger.Error().Err(err).Int("agent_id", id).Msg("error closing agent")
		}
		delete(c.agents, id)
	}

	logger.Debug().Msg("closing browser")
	if err := c.browser.Close(); err != nil {
		logger.Error().Err(err).Msg("error closing browser")
	}

	logger.Info().Msg("crawler shut down")
	return nil
}

func (c *Crawler) CreateAgent() (*CrawlerAgent, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	maxAgents := int(config.Get().Crawler.Agents.MaxAgents)
	if len(c.agents) >= maxAgents {
		return nil, fmt.Errorf("max agents reached (%d)", maxAgents)
	}

	id := c.nextID
	agent, err := newCrawlerAgent(id, c.browser.Browser)
	if err != nil {
		return nil, err
	}

	c.agents[id] = agent
	c.nextID++

	logger.Info().Int("agent_id", id).Int("total", len(c.agents)).Msg("agent created")
	return agent, nil
}

func (c *Crawler) DeleteAgent(id int) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	agent, ok := c.agents[id]
	if !ok {
		return fmt.Errorf("agent %d not found", id)
	}

	if err := agent.Close(); err != nil {
		return fmt.Errorf("close agent %d: %w", id, err)
	}

	delete(c.agents, id)
	logger.Info().Int("agent_id", id).Int("total", len(c.agents)).Msg("agent deleted")
	return nil
}

func (c *Crawler) GetAgent(id int) *CrawlerAgent {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.agents[id]
}

func (c *Crawler) ListAgents() []int {
	c.mu.Lock()
	defer c.mu.Unlock()

	ids := make([]int, 0, len(c.agents))
	for id := range c.agents {
		ids = append(ids, id)
	}
	return ids
}

func (c *Crawler) AgentCount() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.agents)
}
