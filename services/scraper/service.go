// services/scraper/service.go
package scraper

import (
	"context"
	"fmt"
	"sync"

	"smegg.me/thughunter/common/logger"
	corescraper "smegg.me/thughunter/core/scraper"
)

// Service is the Wails-bound scraper service that bridges core/scraper to the frontend.
type Service struct {
	mu     sync.Mutex
	cancel context.CancelFunc
	done   chan struct{} // closed when Run goroutine finishes
}

// Start begins the scraping run in a background goroutine.
func (s *Service) Start() error {
	return s.startRun(func(sc *corescraper.Scraper, ctx context.Context) error {
		return sc.Run(ctx)
	})
}

// RefreshAccounts begins a refresh-only run in a background goroutine.
// It logs into each existing account to update credits without scraping.
func (s *Service) RefreshAccounts() error {
	return s.startRun(func(sc *corescraper.Scraper, ctx context.Context) error {
		return sc.RunRefresh(ctx)
	})
}

// RegisterAccounts begins a registration-only run in a background goroutine.
// It creates new accounts without logging into them.
func (s *Service) RegisterAccounts(opts corescraper.RegisterOpts) error {
	return s.startRun(func(sc *corescraper.Scraper, ctx context.Context) error {
		return sc.RunRegister(ctx, opts)
	})
}

// startRun is the shared run launcher used by Start and RefreshAccounts.
func (s *Service) startRun(runFn func(*corescraper.Scraper, context.Context) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	sc, err := s.requireScraper()
	if err != nil {
		return err
	}

	if sc.IsRunning() {
		return corescraper.ErrScraperAlreadyRunning
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	s.cancel = cancel
	s.done = done

	sub := startEventBridge(sc)
	emitServiceEvent(EventRunStateChanged, true)

	cleanup := func() {
		emitServiceEvent(EventRunStateChanged, false)
		stopEventBridge(sc, sub)
		cancel()
		close(done)
	}

	go func() {
		defer cleanup()
		if err := runFn(sc, ctx); err != nil {
			logger.Error().Err(err).Msg("scraper run failed")
		}
	}()

	return nil
}

// Stop cancels the active scraping run and waits for it to finish.
// It first waits for all browsers to launch (or a timeout) so an
// immediate stop cannot kill a run before it even starts.
func (s *Service) Stop() {
	emitServiceEvent(EventStopping, nil)

	sc, _ := s.requireScraper()
	if sc != nil {
		sc.WaitReady()
	}

	cancel, done := s.takeRunHandles()
	if cancel != nil {
		cancel()
	}
	if done != nil {
		<-done
	}
}

// takeRunHandles atomically retrieves and clears the cancel/done handles.
func (s *Service) takeRunHandles() (context.CancelFunc, chan struct{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	cancel := s.cancel
	done := s.done
	s.cancel = nil
	s.done = nil
	return cancel, done
}

// requireScraper returns the global scraper instance or an error if not initialized.
func (s *Service) requireScraper() (*corescraper.Scraper, error) {
	sc := corescraper.Get()
	if sc == nil {
		return nil, fmt.Errorf("scraper not initialized")
	}
	return sc, nil
}

// IsRunning reports whether a scraping run is in progress.
func (s *Service) IsRunning() bool {
	sc, err := s.requireScraper()
	if err != nil {
		return false
	}
	return sc.IsRunning()
}

// GetProgress returns a point-in-time snapshot of the current run.
func (s *Service) GetProgress() *corescraper.RunProgress {
	sc, err := s.requireScraper()
	if err != nil {
		return &corescraper.RunProgress{}
	}
	return sc.Progress()
}

// GetSummary returns the summary of the last completed run.
func (s *Service) GetSummary() *corescraper.RunSummary {
	sc, err := s.requireScraper()
	if err != nil {
		return nil
	}
	return sc.Summary()
}

// GetQueries returns the query list used by the active (or last) scrape run.
func (s *Service) GetQueries() []string {
	sc, err := s.requireScraper()
	if err != nil {
		return nil
	}
	return sc.Queries()
}

// Shutdown stops any active run and cleans up all agents.
func (s *Service) Shutdown() error {
	s.Stop()

	sc, err := s.requireScraper()
	if err != nil {
		return nil
	}

	shutdownErr := sc.Shutdown()
	emitServiceEvent(EventShutdown, nil)
	emitServiceEvent(EventAgentsChanged, nil)
	return shutdownErr
}
