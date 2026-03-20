// core/scraper/run.go
package scraper

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"smegg.me/thughunter/common/config"
	"smegg.me/thughunter/common/logger"
	"smegg.me/thughunter/core/models"
)

// maxReadyWait is the maximum time Stop will block waiting for browsers to launch.
const maxReadyWait = 8 * time.Second

// Run starts the scraping process: creates agents, dispatches queries, and
// blocks until all queries are processed or ctx is cancelled.
func (s *Scraper) Run(ctx context.Context) error {
	if err := s.acquireRun(RunModeScrape); err != nil {
		return err
	}
	defer s.releaseRun(ctx)

	if err := s.loadAccountPool(); err != nil {
		return err
	}

	usable := s.pool.usableCount()
	if usable == 0 {
		return ErrNoUsableAccounts
	}

	queries, err := s.prepareQueries()
	if err != nil {
		return err
	}

	agentCap := min(int(config.Get().Scraper.Agents.MaxAgents), usable)
	agents, err := s.createAgents(ctx, agentCap)
	if err != nil {
		return fmt.Errorf("create run agents: %w", err)
	}
	s.initReady(len(agents))
	stopForceClose := s.forceCloseOnCancel(ctx)
	defer stopForceClose()
	defer s.cleanupRunAgents()

	return s.dispatch(ctx, queries, agents)
}

// RunRefresh logs into each existing account to refresh credits.
// No scraping or account registration is performed.
func (s *Scraper) RunRefresh(ctx context.Context) error {
	if err := s.acquireRun(RunModeRefresh); err != nil {
		return err
	}
	defer s.releaseRun(ctx)

	if err := s.loadAccountPool(); err != nil {
		return err
	}

	accountCount := s.pool.totalCount()
	if accountCount == 0 {
		return fmt.Errorf("no accounts to refresh")
	}
	s.totalAccounts.Store(int32(accountCount))

	logger.Info().
		Int("accounts", accountCount).
		Msg("starting account credits refresh")

	s.emitEvent(EventRunStarted, "", fmt.Sprintf("refreshing credits for %d accounts", accountCount), map[string]any{
		"total_accounts": accountCount,
		"mode":           string(RunModeRefresh),
	})

	agentCap := min(int(config.Get().Scraper.Agents.MaxAgents), accountCount)
	agents, err := s.createAgents(ctx, agentCap)
	if err != nil {
		return fmt.Errorf("create refresh agents: %w", err)
	}
	s.initReady(len(agents))
	stopForceClose := s.forceCloseOnCancel(ctx)
	defer stopForceClose()
	defer s.cleanupRunAgents()

	return s.refreshDispatch(ctx, agents)
}

// acquireRun claims exclusive ownership of the run lifecycle.
func (s *Scraper) acquireRun(mode RunMode) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return ErrScraperAlreadyRunning
	}
	s.running = true
	s.mode = mode

	s.resetCounters()
	s.summary = nil

	now := time.Now()
	s.startedAt = &now

	return nil
}

// initReady sets up the browser-ready gate for the current run.
func (s *Scraper) initReady(agentCount int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.readyCh = make(chan struct{})
	s.readyTarget = int32(agentCount)
	s.readyCount.Store(0)
}

// markBrowserReady is called by each worker after launchBrowser completes.
// When every agent has reported, the ready channel is closed.
func (s *Scraper) markBrowserReady() {
	if s.readyCount.Add(1) >= atomic.LoadInt32(&s.readyTarget) {
		s.mu.Lock()
		defer s.mu.Unlock()
		if s.readyCh != nil {
			select {
			case <-s.readyCh:
			default:
				close(s.readyCh)
			}
		}
	}
}

// WaitReady blocks until all browsers have launched or the timeout elapses.
func (s *Scraper) WaitReady() {
	s.mu.Lock()
	ch := s.readyCh
	s.mu.Unlock()
	if ch == nil {
		return
	}
	select {
	case <-ch:
	case <-time.After(maxReadyWait):
	}
}

// releaseRun builds the run summary, emits it, and marks the run as finished.
func (s *Scraper) releaseRun(ctx context.Context) {
	summary := s.buildSummary(ctx)
	s.summary = summary

	s.emitEvent(EventRunSummary, "", "run summary", map[string]any{
		"summary": summary,
	})

	s.mu.Lock()
	s.running = false
	s.startedAt = nil
	s.mu.Unlock()
}

// buildSummary captures all counters into a RunSummary snapshot.
func (s *Scraper) buildSummary(ctx context.Context) *RunSummary {
	now := time.Now()
	s.mu.Lock()
	mode := s.mode
	startedAt := s.startedAt
	s.mu.Unlock()

	var start time.Time
	if startedAt != nil {
		start = *startedAt
	}

	totalCredits := int(s.refreshedCredits.Load())
	possibleQueries := 0
	if CreditsAmountPerQuery > 0 && totalCredits > 0 {
		possibleQueries = totalCredits / CreditsAmountPerQuery
	}

	return &RunSummary{
		Mode:                mode,
		StoppedEarly:        ctx.Err() != nil,
		AccountsExhausted:   s.accountsExhausted.Load(),
		StartedAt:           start,
		FinishedAt:          now,
		DurationSecs:        now.Sub(start).Seconds(),
		TotalQueries:        int(s.totalQueries.Load()),
		CompletedQueries:    int(s.completedQueries.Load()),
		EmptyQueries:        int(s.emptyQueries.Load()),
		TotalHosts:          int(s.totalHosts.Load()),
		TotalAccounts:       int(s.totalAccounts.Load()),
		RefreshedAccounts:   int(s.refreshedAccounts.Load()),
		FailedAccounts:      int(s.failedAccounts.Load()),
		CreatedAccounts:     int(s.createdAccounts.Load()),
		FailedRegistrations: int(s.failedRegistrations.Load()),
		TargetAccounts:      int(s.targetAccounts.Load()),
		MaxDurationSecs:     int(s.runDurationSecs.Load()),
		TotalCredits:        totalCredits,
		PossibleQueries:     possibleQueries,
		UsedCredits:         int(s.usedCredits.Load()),
	}
}

// Summary returns the last run's summary (nil if no run has completed yet).
func (s *Scraper) Summary() *RunSummary {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.summary
}

// Queries returns the query list for the active (or last) scrape run.
func (s *Scraper) Queries() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]string, len(s.queries))
	copy(out, s.queries)
	return out
}

// loadAccountPool creates a fresh pool and populates it from the database.
func (s *Scraper) loadAccountPool() error {
	s.pool = newAccountPool()
	if err := s.pool.loadFromDB(); err != nil {
		return fmt.Errorf("load account pool: %w", err)
	}
	return nil
}

// prepareQueries builds the full query list and emits the run-started event.
func (s *Scraper) prepareQueries() ([]string, error) {
	queries := buildQueries()
	if len(queries) == 0 {
		return nil, ErrNoQueries
	}
	s.totalQueries.Store(int32(len(queries)))
	s.mu.Lock()
	s.queries = queries
	s.mu.Unlock()

	logger.Info().
		Int("queries", len(queries)).
		Int("accounts", s.pool.totalCount()).
		Msg("starting scraper run")

	s.emitEvent(EventRunStarted, "", fmt.Sprintf("starting run with %d queries", len(queries)), map[string]any{
		"total_queries":  len(queries),
		"total_accounts": s.pool.totalCount(),
	})

	return queries, nil
}

// dispatch fans out queries to agent workers and collects results.
func (s *Scraper) dispatch(ctx context.Context, queries []string, agents []*ScraperAgent) error {
	queryCh := fillQueryChannel(queries)
	resultCh := make(chan queryResult, len(queries))

	var wg sync.WaitGroup
	var completed, empty, hostCount atomic.Int32

	for _, agent := range agents {
		wg.Add(1)
		go s.runWorker(ctx, agent, queryCh, resultCh, &completed, &empty, &hostCount, &wg)
	}
	logger.Debug().Int("agents", len(agents)).Int("queries", len(queries)).Msg("scrape dispatch: workers spawned")

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	s.drainResults(resultCh, &completed, &empty, &hostCount)

	s.emitEvent(EventRunCompleted, "", "run completed", map[string]any{
		"completed_queries": completed.Load(),
		"empty_queries":     empty.Load(),
		"total_hosts":       hostCount.Load(),
	})

	return nil
}

// fillQueryChannel returns a closed, pre-loaded buffered channel.
func fillQueryChannel(queries []string) <-chan string {
	ch := make(chan string, len(queries))
	for _, q := range queries {
		ch <- q
	}
	close(ch)
	return ch
}

// refreshResult tracks the outcome of a single account refresh.
type refreshResult struct {
	Email   string
	Error   error
	Agent   string
	Account *models.Account // non-nil when Retry is true
	Retry   bool            // re-queue account for later retry
	Retries int             // how many times this account has been retried
}

// refreshDispatch fans out accounts to refresh workers and collects results.
// Accounts whose login timed out are re-queued to the end of the channel
// so they are retried later (after other accounts have been processed).
func (s *Scraper) refreshDispatch(ctx context.Context, agents []*ScraperAgent) error {
	poolSize := s.pool.totalCount()
	accountCh := make(chan refreshAccountEntry, poolSize*2)
	skipped := s.pool.fillRefreshChan(accountCh)
	resultCh := make(chan refreshResult, poolSize)

	activeCount := int32(poolSize - skipped)
	if activeCount <= 0 {
		logger.Info().Int("skipped", skipped).Msg("all accounts skipped, nothing to refresh")
		s.emitEvent(EventRunCompleted, "", "refresh skipped: no accounts need refreshing", map[string]any{
			"skipped_accounts": skipped,
			"mode":             string(RunModeRefresh),
		})
		return nil
	}
	s.totalAccounts.Store(activeCount)

	logger.Info().Int("active", int(activeCount)).Int("skipped", skipped).Msg("refresh dispatch: accounts queued")

	var wg sync.WaitGroup
	for _, agent := range agents {
		wg.Add(1)
		go s.runRefreshWorker(ctx, agent, accountCh, resultCh, &wg)
	}

	// Drain results. Re-queue retryable accounts back into the channel
	// so idle workers can pick them up again.
	done := make(chan struct{})
	go func() {
		defer close(done)
		for r := range resultCh {
			if r.Retry && r.Account != nil {
				logger.Info().Str("email", r.Account.Email).Int("retries", r.Retries).Msg("re-queuing account for later retry")
				select {
				case accountCh <- refreshAccountEntry{Account: r.Account, Retries: r.Retries}:
				default:
					logger.Warn().Str("email", r.Account.Email).Msg("account channel full, cannot re-queue")
					s.failedAccounts.Add(1)
				}
				continue
			}
			if r.Error != nil {
				if errors.Is(r.Error, context.Canceled) || errors.Is(r.Error, context.DeadlineExceeded) {
					continue
				}
				s.failedAccounts.Add(1)
			} else {
				s.refreshedAccounts.Add(1)
			}

			// Close the account channel when all non-retry results have
			// been collected so workers can exit.
			total := s.refreshedAccounts.Load() + s.failedAccounts.Load()
			if total >= s.totalAccounts.Load() {
				close(accountCh)
			}
		}
	}()

	wg.Wait()
	close(resultCh)
	<-done

	logger.Info().
		Int32("refreshed", s.refreshedAccounts.Load()).
		Int32("failed", s.failedAccounts.Load()).
		Msg("account refresh run finished")

	s.emitEvent(EventRunCompleted, "", "refresh completed", map[string]any{
		"refreshed_accounts": s.refreshedAccounts.Load(),
		"failed_accounts":    s.failedAccounts.Load(),
		"mode":               string(RunModeRefresh),
	})

	return nil
}

// drainResults reads query results and keeps the progress counters in sync.
func (s *Scraper) drainResults(
	results <-chan queryResult,
	completed, empty, hostCount *atomic.Int32,
) {
	for range results {
		s.completedQueries.Store(completed.Load())
		s.emptyQueries.Store(empty.Load())
		s.totalHosts.Store(hostCount.Load())
	}

	logger.Info().
		Int32("completed", completed.Load()).
		Int32("empty", empty.Load()).
		Int32("hosts", hostCount.Load()).
		Msg("scraper run finished")
}

// IsRunning reports whether a scraper run is currently in progress.
func (s *Scraper) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}

// Mode returns the current run mode. Only meaningful when IsRunning is true.
func (s *Scraper) Mode() RunMode {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.mode
}

// RunRegister creates new accounts in a loop without logging into them.
// It stops when the target count is reached, the duration expires, or ctx is cancelled.
func (s *Scraper) RunRegister(ctx context.Context, opts RegisterOpts) error {
	if err := s.acquireRun(RunModeRegister); err != nil {
		return err
	}

	s.targetAccounts.Store(int32(opts.TargetAccounts))
	s.runDurationSecs.Store(int32(opts.DurationSecs))

	if opts.DurationSecs > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(opts.DurationSecs)*time.Second)
		defer cancel()
	}
	defer s.releaseRun(ctx)

	logger.Info().
		Int("target", opts.TargetAccounts).
		Int("duration_secs", opts.DurationSecs).
		Msg("starting account registration run")

	s.emitEvent(EventRunStarted, "", "starting registration run", map[string]any{
		"mode":            string(RunModeRegister),
		"target_accounts": opts.TargetAccounts,
		"duration_secs":   opts.DurationSecs,
	})

	agents, err := s.createAgents(ctx, int(config.Get().Scraper.Agents.MaxAgents))
	if err != nil {
		return fmt.Errorf("create register agents: %w", err)
	}
	s.initReady(len(agents))
	stopForceClose := s.forceCloseOnCancel(ctx)
	defer stopForceClose()
	defer s.cleanupRunAgents()

	return s.registerDispatch(ctx, agents)
}

// registerDispatch fans out registration work to agent workers.
func (s *Scraper) registerDispatch(ctx context.Context, agents []*ScraperAgent) error {
	resultCh := make(chan registerResult, len(agents)*4)

	var wg sync.WaitGroup
	for _, agent := range agents {
		wg.Add(1)
		go s.runRegisterWorker(ctx, agent, resultCh, &wg)
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		for r := range resultCh {
			if r.Error != nil {
				if errors.Is(r.Error, context.Canceled) || errors.Is(r.Error, context.DeadlineExceeded) {
					continue
				}
				s.failedRegistrations.Add(1)
			} else {
				s.createdAccounts.Add(1)
			}
		}
	}()

	wg.Wait()
	close(resultCh)
	<-done

	logger.Info().
		Int32("created", s.createdAccounts.Load()).
		Int32("failed", s.failedRegistrations.Load()).
		Msg("registration run finished")

	s.emitEvent(EventRunCompleted, "", "registration completed", map[string]any{
		"created_accounts":     s.createdAccounts.Load(),
		"failed_registrations": s.failedRegistrations.Load(),
		"mode":                 string(RunModeRegister),
	})

	return nil
}

// registerResult tracks the outcome of a single account registration.
type registerResult struct {
	Email string
	Error error
	Agent string
}
