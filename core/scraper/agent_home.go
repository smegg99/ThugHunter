// core/scraper/agent_home.go
package scraper

import (
	"github.com/go-rod/rod"

	"smegg.me/thughunter/common/config"
	"smegg.me/thughunter/common/logger"
)

// Home navigates the agent to the home page.
func (a *ScraperAgent) Home() error {
	if !a.IsLoggedIn() {
		return ErrAgentNotLoggedIn
	}

	logger.Info().Str("agent", a.Name).Msg("navigating to home page")

	return a.runTask(func() error {
		page, _, err := a.newTab(config.Get().Scraper.Endpoints.HomeEndpoint)
		if err != nil {
			return err
		}
		defer page.MustClose()

		a.dismissPendoDialog(page)
		return nil
	})
}

// dismissPendoDialog removes the Pendo walkthrough overlay if present.
func (a *ScraperAgent) dismissPendoDialog(page *rod.Page) {
	_, err := page.Eval(`() => {
		const el = document.getElementById('pendo-base');
		if (el) el.remove();
	}`)
	if err != nil {
		logger.Debug().Str("agent", a.Name).Err(err).Msg("pendo dialog not found or already dismissed")
	}
}
