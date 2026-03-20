// core/scraper/agent_login.go
package scraper

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-rod/rod"

	"github.com/smegg99/human"
	"smegg.me/thughunter/common/config"
	"smegg.me/thughunter/common/logger"
	"smegg.me/thughunter/core/models"
)

const (
	loginPageScanBaseMs   = 50
	loginPageScanJitterMs = 100
	loginBtnTimeout       = 45 * time.Second
	loginHomeTimeout      = 45 * time.Second
	loginHomePollInterval = 100 * time.Millisecond
	loginClickMaxRetries  = 5
	loginClickRetryDelay  = 2 * time.Second

	// Selector for the error-message banner shown after a failed login (e.g. "This account is not active yet").
	loginErrorMessageSel = `[data-testid="error-message"]`

	// Selector for the avatar dropdown that appears only when the user is fully logged in.
	loginAvatarSel = `[class*="_avatarContainer_"]`
)

// Login authenticates the agent with the given account.
func (a *ScraperAgent) Login(ctx context.Context, account *models.Account) error {
	if account == nil {
		if a.account == nil {
			return fmt.Errorf("login: %w", ErrNoAccountProvided)
		}
		logger.Debug().Str("agent", a.Name).Str("email", a.account.Email).Msg("no account provided, using agent's current account")
		account = a.account
	}

	logger.Info().Str("agent", a.Name).Str("email", account.Email).Msg("logging in")

	page, cursor, err := a.loginOpenPage(ctx)
	if err != nil {
		return err
	}

	// If the login page redirected to the homepage, an old session is
	// still active. Clear the session and open the login page again.
	if a.loginDetectRedirect(page) {
		logger.Warn().Str("agent", a.Name).Msg("login page redirected to home, clearing stale session")
		_ = a.ClearSession()
		page, cursor, err = a.loginOpenPage(ctx)
		if err != nil {
			return err
		}
		if a.loginDetectRedirect(page) {
			return fmt.Errorf("login: %w: redirected to home after session clear", ErrLoginFailed)
		}
	}

	if err := a.loginFillFields(page, cursor, account); err != nil {
		return err
	}
	if err := a.loginSubmit(page, cursor, account); err != nil {
		return err
	}
	if err := a.loginAwaitSuccess(ctx, page); err != nil {
		return err
	}

	a.account = account
	a.estimatedCredits = int(account.CreditsAmount)
	logger.Info().Str("agent", a.Name).Str("email", account.Email).Msg("agent is now logged in")

	return nil
}

func (a *ScraperAgent) loginOpenPage(ctx context.Context) (*rod.Page, *human.Cursor, error) {
	ep := config.Get().Scraper.Endpoints
	page, cursor, err := a.openPageWithJitter(ctx, ep.LoginEndpoint, loginPageScanBaseMs, loginPageScanJitterMs)
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

	var loginBtn *rod.Element
	var err error

	for attempt := 1; attempt <= loginClickMaxRetries; attempt++ {
		loginBtn, err = awaitElement(page, `button[role="button"][class*="_loginButton_"]:not([disabled])`, loginBtnTimeout, "login: Log in button")
		if err != nil {
			logger.Error().Str("agent", a.Name).Err(err).Msg("Log in button did not become clickable in time")
			return err
		}

		logger.Info().Str("agent", a.Name).Int("attempt", attempt).Msg("Log in button is now clickable, clicking it")

		if err = cursor.Click(loginBtn); err != nil {
			logger.Warn().Str("agent", a.Name).Err(err).Int("attempt", attempt).Msg("click on Log in button failed")
			if attempt < loginClickMaxRetries {
				time.Sleep(loginClickRetryDelay)
				continue
			}
			return fmt.Errorf("login: click Log in button after %d attempts: %w", loginClickMaxRetries, err)
		}

		logger.Info().
			Str("agent", a.Name).
			Str("email", account.Email).
			Msg("login submitted")
		return nil
	}

	return fmt.Errorf("login: click Log in button: %w", err)
}

// censysCookieURLs are queried explicitly so we catch auth cookies regardless
// of which subdomain the page is currently on during the login redirect chain.
var censysCookieURLs = []string{
	"https://censys.io/",
	"https://accounts.censys.io/",
	"https://platform.censys.io/",
}

func (a *ScraperAgent) loginAwaitSuccess(ctx context.Context, page *rod.Page) error {
	homeURL := config.Get().Scraper.Endpoints.HomeEndpoint
	logger.Debug().Str("agent", a.Name).Msg("waiting for login (avatar, auth cookie, or home redirect)")

	deadline := time.Now().Add(loginHomeTimeout)
	for time.Now().Before(deadline) {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		// Avatar dropdown visible (user is fully logged in).
		if has, _, _ := page.Has(loginAvatarSel); has {
			logger.Info().Str("agent", a.Name).Msg("avatar dropdown detected, login successful")
			return nil
		}

		// Auth cookie on any censys domain.
		if a.hasAuthCookie(page) {
			logger.Info().Str("agent", a.Name).Msg("auth cookie detected, login successful")
			return nil
		}

		// URL redirect to home page.
		if info, err := page.Info(); err == nil && info != nil && strings.HasPrefix(info.URL, homeURL) {
			logger.Info().Str("agent", a.Name).Str("url", info.URL).Msg("home page loaded after login")
			return nil
		}

		// Has error banner (e.g. "account not active" or "Incorrect email or password").
		if has, el, _ := page.Has(loginErrorMessageSel); has && el != nil {
			if text, _ := el.Text(); text != "" {
				switch {
				case strings.Contains(text, "not active"):
					logger.Warn().Str("agent", a.Name).Str("message", text).Msg("account not active banner detected")
					return fmt.Errorf("login: %w", ErrAccountNotActive)
				case strings.Contains(text, "Incorrect"):
					logger.Warn().Str("agent", a.Name).Str("message", text).Msg("invalid credentials banner detected")
					return fmt.Errorf("login: %w", ErrInvalidCredentials)
				}
			}
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(loginHomePollInterval):
		}
	}

	currentURL := ""
	if info, _ := page.Info(); info != nil {
		currentURL = info.URL
	}
	logger.Error().Str("agent", a.Name).Str("current_url", currentURL).Msg("login not detected within timeout")
	return fmt.Errorf("login: %w: timed out after %s (current: %s)", ErrLoginFailed, loginHomeTimeout, currentURL)
}

// loginDetectRedirect checks whether the login page was redirected to the
// home page, which happens when the browser still carries a valid session.
func (a *ScraperAgent) loginDetectRedirect(page *rod.Page) bool {
	homeURL := config.Get().Scraper.Endpoints.HomeEndpoint
	info, err := page.Info()
	if err != nil {
		return false
	}
	return info != nil && strings.HasPrefix(info.URL, homeURL)
}

// hasAuthCookie checks all censys domains for a cookie indicating an active session.
func (a *ScraperAgent) hasAuthCookie(page *rod.Page) bool {
	cookies, err := page.Cookies(censysCookieURLs)
	if err != nil {
		return false
	}
	for _, c := range cookies {
		name := strings.ToLower(c.Name)
		if strings.HasPrefix(name, "auth") {
			return true
		}
	}
	return false
}
