// core/scraper/scraper.go
package scraper

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"smegg.me/thughunter/common/config"
	"smegg.me/thughunter/common/logger"
)

// Scraper manages a pool of ScraperAgents and orchestrates scraping runs.
type Scraper struct {
	mu     sync.Mutex
	agents map[string]*ScraperAgent
	pool   *accountPool
	events *eventBus

	// Run lifecycle
	running          bool
	mode             RunMode
	totalQueries     atomic.Int32
	completedQueries atomic.Int32
	emptyQueries     atomic.Int32
	totalHosts       atomic.Int32
	usedCredits      atomic.Int32 // credits consumed by successful queries

	// Refresh-mode counters
	totalAccounts     atomic.Int32
	refreshedAccounts atomic.Int32
	failedAccounts    atomic.Int32
	refreshedCredits  atomic.Int32 // sum of credits from successfully refreshed accounts

	// Register-mode counters
	createdAccounts     atomic.Int32
	failedRegistrations atomic.Int32
	targetAccounts      atomic.Int32
	runDurationSecs     atomic.Int32

	// Scrape-mode flags
	accountsExhausted atomic.Bool // set when no usable pool accounts remain

	// Active query list (set per scrape run)
	queries []string

	// Run start timestamp
	startedAt *time.Time

	// Browser-ready tracking (set per run)
	readyCh     chan struct{} // closed when all browsers launched or timed out
	readyTarget int32         // total agents expected
	readyCount  atomic.Int32  // agents that finished launchBrowser

	// Last run summary (cleared on next run start)
	summary *RunSummary
}

// resetCounters zeroes all atomic run/refresh/register counters.
func (s *Scraper) resetCounters() {
	s.totalQueries.Store(0)
	s.completedQueries.Store(0)
	s.emptyQueries.Store(0)
	s.totalHosts.Store(0)
	s.usedCredits.Store(0)
	s.totalAccounts.Store(0)
	s.refreshedAccounts.Store(0)
	s.failedAccounts.Store(0)
	s.refreshedCredits.Store(0)
	s.createdAccounts.Store(0)
	s.failedRegistrations.Store(0)
	s.targetAccounts.Store(0)
	s.runDurationSecs.Store(0)
	s.accountsExhausted.Store(false)
}

var (
	instance *Scraper
	initOnce sync.Once
)

// Initialize creates the global Scraper singleton.
func Initialize() error {
	var initErr error
	initOnce.Do(func() {
		logger.Info().Msg("initializing scraper")

		cfg := config.Get()
		maxAgents := cfg.Scraper.Agents.MaxAgents
		virtualDisplay := cfg.Scraper.VirtualDisplay

		instance = &Scraper{
			agents: make(map[string]*ScraperAgent),
			events: newEventBus(),
		}

		logger.Info().Int64("max_agents", maxAgents).Bool("virtual_display", virtualDisplay).Msg("scraper initialized")
	})
	return initErr
}

func Get() *Scraper {
	return instance
}

// Shutdown closes all agents in parallel and releases resources.
func (c *Scraper) Shutdown() error {
	c.mu.Lock()
	agents := make(map[string]*ScraperAgent, len(c.agents))
	for k, v := range c.agents {
		agents[k] = v
	}
	c.agents = make(map[string]*ScraperAgent)
	c.mu.Unlock()

	logger.Info().Int("total", len(agents)).Msg("shutting down agents")

	var wg sync.WaitGroup
	for name, agent := range agents {
		wg.Add(1)
		go func(n string, a *ScraperAgent) {
			defer wg.Done()
			logger.Debug().Str("agent", n).Msg("force-closing agent")
			a.ForceClose()
		}(name, agent)
	}
	wg.Wait()

	logger.Info().Msg("agents shut down")
	return nil
}

const maxPetnameAttempts = 100

// CreateAgent creates a new named agent and adds it to the map.
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
	agent.onStatusText = func(info AgentInfo) {
		c.emitEvent(EventStatusUpdate, info.Name, info.StatusText, map[string]any{"agent_info": info})
	}
	c.agents[name] = agent

	logger.Info().Str("agent", name).Int("total", len(c.agents)).Msg("agent created")
	return agent, nil
}

// DeleteAgent removes an agent by name and closes its browser.
func (c *Scraper) DeleteAgent(name string) error {
	c.mu.Lock()
	agent, ok := c.agents[name]
	if !ok {
		c.mu.Unlock()
		return fmt.Errorf("%w: %q", ErrAgentNotFound, name)
	}
	delete(c.agents, name)
	remaining := len(c.agents)
	c.mu.Unlock()

	if err := agent.Close(); err != nil {
		logger.Error().Err(err).Str("agent", name).Msg("error closing agent during deletion")
	}

	logger.Info().Str("agent", name).Int("total", remaining).Msg("agent deleted")
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

// AccountPool returns the current account pool (nil when no run is active).
func (c *Scraper) AccountPool() *accountPool {
	return c.pool
}
