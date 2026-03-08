// core/scraper/agent_login.go
package scraper

import (
	"fmt"
	"math/rand/v2"
	"strings"
	"time"

	"github.com/go-rod/rod"

	"smegg.me/thughunter/common/config"
	"smegg.me/thughunter/common/logger"
	"smegg.me/thughunter/core/human"
	"smegg.me/thughunter/core/models"
)

const (
	loginPageScanBaseMs   = 100
	loginPageScanJitterMs = 200
	loginBtnTimeout       = 30 * time.Second
	loginHomeTimeout      = 30 * time.Second
	loginHomePollInterval = 500 * time.Millisecond
)

func (a *ScraperAgent) Login(account *models.Account) error {
	logger.Info().Str("agent", a.Name).Str("email", account.Email).Msg("logging in")

	if account == nil {
		if a.account == nil {
			return fmt.Errorf("login: no account provided and agent has no account")
		}
		logger.Debug().Str("agent", a.Name).Str("email", a.account.Email).Msg("no account provided, using agent's current account")
		account = a.account
	}

	a.SetStatus(AgentStatusBusy)

	page, cursor, err := a.loginOpenPage()
	if err != nil {
		a.SetStatus(AgentStatusError)
		return err
	}
	defer page.MustClose()

	if err := a.loginFillFields(page, cursor, account); err != nil {
		a.SetStatus(AgentStatusError)
		return err
	}

	if err := a.loginSubmit(page, cursor, account); err != nil {
		a.SetStatus(AgentStatusError)
		return err
	}

	if err := a.loginAwaitHomePage(page); err != nil {
		a.SetStatus(AgentStatusError)
		return err
	}

	a.account = account
	a.SetStatus(AgentStatusIdle)

	logger.Info().Str("agent", a.Name).Str("email", account.Email).Msg("agent is now logged in")

	return nil
}

func (a *ScraperAgent) loginOpenPage() (*rod.Page, *human.Cursor, error) {
	ep := config.Get().Scraper.Endpoints
	page, cursor, err := a.newTab(ep.LoginEndpoint)
	if err != nil {
		return nil, nil, fmt.Errorf("login: %w", err)
	}
	time.Sleep(time.Duration(loginPageScanBaseMs+rand.IntN(loginPageScanJitterMs)) * time.Millisecond)
	return page, cursor, nil
}

func (a *ScraperAgent) loginFillFields(page *rod.Page, cursor *human.Cursor, account *models.Account) error {
	fields := []struct {
		selector string
		value    string
		label    string
	}{
		{`input[name="identifier"]`, account.Email, "email"},
		{`input[name="password"]`, account.Password, "password"},
	}

	for _, f := range fields {
		logger.Debug().Str("agent", a.Name).Str("field", f.label).Msg("filling field")
		el, err := page.Element(f.selector)
		if err != nil {
			return fmt.Errorf("login: find %s: %w", f.label, err)
		}
		if err := cursor.ClickAndType(el, f.value); err != nil {
			return fmt.Errorf("login: fill %s: %w", f.label, err)
		}
	}
	return nil
}

func (a *ScraperAgent) loginSubmit(page *rod.Page, cursor *human.Cursor, account *models.Account) error {
	logger.Debug().Str("agent", a.Name).Msg("waiting for Log in button to become clickable")

	loginBtn, err := page.Timeout(loginBtnTimeout).Element(`button[role="button"][class*="_loginButton_"]:not([disabled])`)
	if err != nil {
		logger.Error().Str("agent", a.Name).Err(err).Msg("Log in button did not become clickable in time")
		return fmt.Errorf("login: Log in button timed out: %w", err)
	}

	logger.Info().Str("agent", a.Name).Msg("Log in button is now clickable, clicking it")

	if err := cursor.Click(loginBtn); err != nil {
		return fmt.Errorf("login: click Log in button: %w", err)
	}

	logger.Info().
		Str("agent", a.Name).
		Str("email", account.Email).
		Msg("login submitted")

	return nil
}

func (a *ScraperAgent) loginAwaitHomePage(page *rod.Page) error {
	homeURL := config.Get().Scraper.Endpoints.HomeUrl

	logger.Debug().Str("agent", a.Name).Str("expected_url", homeURL).Msg("waiting for redirect to home page")

	deadline := time.Now().Add(loginHomeTimeout)
	for time.Now().Before(deadline) {
		info, err := page.Info()
		if err == nil && info != nil && strings.HasPrefix(info.URL, homeURL) {
			logger.Info().Str("agent", a.Name).Str("url", info.URL).Msg("home page loaded after login")
			return nil
		}
		time.Sleep(loginHomePollInterval)
	}

	info, _ := page.Info()
	currentURL := ""
	if info != nil {
		currentURL = info.URL
	}

	logger.Error().
		Str("agent", a.Name).
		Str("current_url", currentURL).
		Str("expected_url", homeURL).
		Msg("home page did not load after login")

	return fmt.Errorf("login: home page did not load within %s (current: %s)", loginHomeTimeout, currentURL)
}
