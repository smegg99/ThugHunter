// core/crawler/agent_login.go
package crawler

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

func (a *CrawlerAgent) Login(account *models.Account) error {
	logger.Info().Int("agent_id", a.ID).Str("email", account.Email).Msg("logging in")

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

	a.loggedIn = true
	a.account = account

	logger.Info().Int("agent_id", a.ID).Str("email", account.Email).Msg("agent is now logged in")

	return nil
}

func (a *CrawlerAgent) loginOpenPage() (*rod.Page, *human.Cursor, error) {
	ep := config.Get().Crawler.Endpoints
	page, cursor, err := a.newTab(ep.LoginEndpoint)
	if err != nil {
		return nil, nil, fmt.Errorf("login: %w", err)
	}
	time.Sleep(time.Duration(loginPageScanBaseMs+rand.IntN(loginPageScanJitterMs)) * time.Millisecond)
	return page, cursor, nil
}

func (a *CrawlerAgent) loginFillFields(page *rod.Page, cursor *human.Cursor, account *models.Account) error {
	fields := []struct {
		selector string
		value    string
		label    string
	}{
		{`input[name="identifier"]`, account.Email, "email"},
		{`input[name="password"]`, account.Password, "password"},
	}

	for _, f := range fields {
		logger.Debug().Int("agent_id", a.ID).Str("field", f.label).Msg("filling field")
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

func (a *CrawlerAgent) loginSubmit(page *rod.Page, cursor *human.Cursor, account *models.Account) error {
	logger.Debug().Int("agent_id", a.ID).Msg("waiting for Log in button to become clickable")

	loginBtn, err := page.Timeout(loginBtnTimeout).Element(`button[role="button"][class*="_loginButton_"]:not([disabled])`)
	if err != nil {
		logger.Error().Int("agent_id", a.ID).Err(err).Msg("Log in button did not become clickable in time")
		return fmt.Errorf("login: Log in button timed out: %w", err)
	}

	logger.Info().Int("agent_id", a.ID).Msg("Log in button is now clickable, clicking it")

	if err := cursor.Click(loginBtn); err != nil {
		return fmt.Errorf("login: click Log in button: %w", err)
	}

	logger.Info().
		Int("agent_id", a.ID).
		Str("email", account.Email).
		Msg("login submitted")

	return nil
}

func (a *CrawlerAgent) loginAwaitHomePage(page *rod.Page) error {
	homeURL := config.Get().Crawler.Endpoints.HomeUrl

	logger.Debug().Int("agent_id", a.ID).Str("expected_url", homeURL).Msg("waiting for redirect to home page")

	deadline := time.Now().Add(loginHomeTimeout)
	for time.Now().Before(deadline) {
		info, err := page.Info()
		if err == nil && info != nil && strings.HasPrefix(info.URL, homeURL) {
			logger.Info().Int("agent_id", a.ID).Str("url", info.URL).Msg("home page loaded after login")
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
		Int("agent_id", a.ID).
		Str("current_url", currentURL).
		Str("expected_url", homeURL).
		Msg("home page did not load after login")

	return fmt.Errorf("login: home page did not load within %s (current: %s)", loginHomeTimeout, currentURL)
}
