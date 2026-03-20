// core/scraper/worker_browser.go
package scraper

import (
	"context"
	"math/rand/v2"
	"strings"
	"time"

	"smegg.me/thughunter/common/i18n"
	"smegg.me/thughunter/common/logger"
)

// isConnectionError reports whether the error indicates a dead browser connection.
func isConnectionError(err error) bool {
	if err == nil {
		return false
	}
	s := err.Error()
	return strings.Contains(s, "closed network connection") ||
		strings.Contains(s, "connection refused") ||
		strings.Contains(s, "broken pipe") ||
		strings.Contains(s, "connection reset")
}

// maxStartupDelay is the upper bound of the random stagger applied before
// each agent launches its browser for the first time. This helps avoid overwhelming the system with simultaneous browser startups. And probably goes easier on the censys servers too.
const maxStartupDelay = 15 * time.Second

// launchBrowser ensures the agent's browser is running and emits the
// appropriate start/error events. Returns false if the worker should exit.
func (s *Scraper) launchBrowser(ctx context.Context, agent *ScraperAgent) bool {
	if delay := rand.N(maxStartupDelay); delay > 0 {
		agent.SetStatusText(i18n.T("agent.startupWaiting"))
		logger.Debug().Str("agent", agent.Name).Dur("delay", delay).Msg("staggering startup")
		select {
		case <-ctx.Done():
			return false
		case <-time.After(delay):
		}
	}

	agent.SetStatusText(i18n.T("agent.launchingBrowser"))
	if err := agent.ensureBrowser(ctx); err != nil {
		s.markBrowserReady()
		if ctx.Err() != nil {
			return false
		}
		logger.Error().Err(err).Str("agent", agent.Name).Msg("failed to launch browser")
		agent.SetStatus(AgentStatusError)
		agent.SetStatusText(i18n.T("agent.browserLaunchFailed"))
		s.emitEvent(EventAgentError, agent.Name, "browser launch failed", nil)
		return false
	}
	s.markBrowserReady()
	logger.Debug().Str("agent", agent.Name).Msg("browser launched successfully")
	return true
}

// shutdownWorker sets the agent to offline, force-closes it, and emits a stop event.
func (s *Scraper) shutdownWorker(agent *ScraperAgent) {
	agent.SetStatusText(i18n.T("agent.stopped"))
	agent.SetStatus(AgentStatusOffline)
	agent.ForceClose()
	s.emitEvent(EventAgentStopped, agent.Name, "worker stopped", nil)
}

// relaunchOnConnectionError checks if err is a connection error and, if so,
// relaunches the browser. Returns true if the relaunch succeeded or the error
// was not connection-related. Returns false if the relaunch failed.
func (s *Scraper) relaunchOnConnectionError(ctx context.Context, agent *ScraperAgent, err error) bool {
	if !isConnectionError(err) {
		return true
	}
	logger.Warn().Err(err).Str("agent", agent.Name).Msg("browser connection dead, relaunching")
	agent.SetStatusText(i18n.T("agent.relaunchingBrowser"))
	if rerr := agent.RelaunchBrowser(ctx); rerr != nil {
		logger.Error().Err(rerr).Str("agent", agent.Name).Msg("failed to relaunch browser")
		agent.SetStatus(AgentStatusError)
		agent.SetStatusText(i18n.T("agent.browserRelaunchFailed"))
		return false
	}
	return true
}
