// core/scraper/agent_login.go
package scraper

import (
	"fmt"
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

// Login authenticates the agent with the given account (or re-uses the current one).
func (a *ScraperAgent) Login(account *models.Account) error {
	if account == nil {
		if a.account == nil {
			return fmt.Errorf("login: %w", ErrNoAccountProvided)
		}
		logger.Debug().Str("agent", a.Name).Str("email", a.account.Email).Msg("no account provided, using agent's current account")
		account = a.account
	}

	logger.Info().Str("agent", a.Name).Str("email", account.Email).Msg("logging in")

	return a.runTask(func() error {
		page, cursor, err := a.loginOpenPage()
		if err != nil {
			return err
		}
		defer page.MustClose()

		if err := a.loginFillFields(page, cursor, account); err != nil {
			return err
		}
		if err := a.loginSubmit(page, cursor, account); err != nil {
			return err
		}
		if err := a.loginAwaitHomePage(page); err != nil {
			return err
		}

		a.account = account
		logger.Info().Str("agent", a.Name).Str("email", account.Email).Msg("agent is now logged in")
		return nil
	})
}

func (a *ScraperAgent) loginOpenPage() (*rod.Page, *human.Cursor, error) {
	ep := config.Get().Scraper.Endpoints
	page, cursor, err := a.openPageWithJitter(ep.LoginEndpoint, loginPageScanBaseMs, loginPageScanJitterMs)
	if err != nil {
		return nil, nil, fmt.Errorf("login: %w", err)
	}
	return page, cursor, nil
}

func (a *ScraperAgent) loginFillFields(page *rod.Page, cursor *human.Cursor, account *models.Account) error {
	return a.fillFormFields(page, cursor, "login", []formField{
		{`input[name="identifier"]`, account.Email, "email"},
		{`input[name="password"]`, account.Password, "password"},
	})
}

func (a *ScraperAgent) loginSubmit(page *rod.Page, cursor *human.Cursor, account *models.Account) error {
	logger.Debug().Str("agent", a.Name).Msg("waiting for Log in button to become clickable")

	loginBtn, err := awaitElement(page, `button[role="button"][class*="_loginButton_"]:not([disabled])`, loginBtnTimeout, "login: Log in button")
	if err != nil {
		logger.Error().Str("agent", a.Name).Err(err).Msg("Log in button did not become clickable in time")
		return err
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
	homeURL := config.Get().Scraper.Endpoints.HomeEndpoint

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

	return fmt.Errorf("login: %w: home page did not load within %s (current: %s)", ErrLoginFailed, loginHomeTimeout, currentURL)
}
