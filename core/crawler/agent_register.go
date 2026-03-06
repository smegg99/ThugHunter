// core/crawler/agent_register.go
package crawler

import (
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/input"

	"smegg.me/thughunter/common/config"
	"smegg.me/thughunter/common/logger"
	"smegg.me/thughunter/core/catcher"
	"smegg.me/thughunter/core/human"
	"smegg.me/thughunter/core/models"
)

const (
	registerPageScanBaseMs     = 100
	registerPageScanJitterMs   = 200
	registerPostTermsBaseMs    = 100
	registerPostTermsJitterMs  = 200
	registerVerifyBtnTimeout   = 30 * time.Second
	registerVerifyPageTimeout  = 30 * time.Second
	registerOTPInputTimeout    = 10 * time.Second
	registerContinueBtnTimeout = 10 * time.Second
	registerCodeTimeout        = 2 * time.Minute
)

func (a *CrawlerAgent) Register() (*models.Account, error) {
	logger.Info().Int("agent_id", a.ID).Msg("registering account")

	account, err := a.registerCreateAccount()
	if err != nil {
		return nil, err
	}

	mc, err := catcher.New()
	if err != nil {
		return nil, fmt.Errorf("register: start catcher: %w", err)
	}
	defer mc.Close()

	codeCh := a.registerListenForCode(mc, account.Email)

	page, cursor, err := a.registerOpenPage()
	if err != nil {
		return nil, err
	}

	if err := a.registerFillForm(page, cursor, account); err != nil {
		return nil, err
	}

	if err := a.registerAcceptTerms(page, cursor); err != nil {
		return nil, err
	}

	if err := a.registerSubmit(page, cursor, account); err != nil {
		return nil, err
	}

	if err := a.registerAwaitVerificationPage(page, account); err != nil {
		return nil, err
	}

	if err := a.registerEnterCode(page, cursor, account, codeCh); err != nil {
		return nil, err
	}

	return account, nil
}

func (a *CrawlerAgent) registerCreateAccount() (*models.Account, error) {
	accountID := fmt.Sprintf("%d%d", a.ID, time.Now().UnixMilli())
	account, err := newAccountFromTemplates(accountID)
	if err != nil {
		return nil, fmt.Errorf("register: create account: %w", err)
	}
	return account, nil
}

type registerCodeResult struct {
	code string
	err  error
}

func (a *CrawlerAgent) registerListenForCode(mc *catcher.Catcher, email string) <-chan registerCodeResult {
	ch := make(chan registerCodeResult, 1)
	go func() {
		code, err := mc.WaitForCode(email, registerCodeTimeout)
		ch <- registerCodeResult{code, err}
	}()
	return ch
}

func (a *CrawlerAgent) registerOpenPage() (*rod.Page, *human.Cursor, error) {
	ep := config.Get().Crawler.Endpoints
	page, cursor, err := a.newTab(ep.RegisterEndpoint)
	if err != nil {
		return nil, nil, fmt.Errorf("register: %w", err)
	}
	time.Sleep(time.Duration(registerPageScanBaseMs+rand.IntN(registerPageScanJitterMs)) * time.Millisecond)
	return page, cursor, nil
}

func (a *CrawlerAgent) registerFillForm(page *rod.Page, cursor *human.Cursor, account *models.Account) error {
	fields := []struct {
		selector string
		value    string
		label    string
	}{
		{"#first_name", account.FirstName, "first name"},
		{"#last_name", account.LastName, "last name"},
		{"#identifier", account.Email, "email"},
		{"#org_name", account.Organization, "organization"},
		{"#password", account.Password, "password"},
		{"#confirm-password", account.Password, "confirm password"},
	}

	for _, f := range fields {
		logger.Debug().Int("agent_id", a.ID).Str("field", f.label).Msg("filling field")
		el, err := page.Element(f.selector)
		if err != nil {
			return fmt.Errorf("register: find %s: %w", f.label, err)
		}
		if err := cursor.ClickAndType(el, f.value); err != nil {
			return fmt.Errorf("register: fill %s: %w", f.label, err)
		}
	}
	return nil
}

func (a *CrawlerAgent) registerAcceptTerms(page *rod.Page, cursor *human.Cursor) error {
	logger.Debug().Int("agent_id", a.ID).Msg("ticking terms checkbox")
	termsEl, err := page.Element("#terms-and-conditions")
	if err != nil {
		return fmt.Errorf("register: find terms checkbox: %w", err)
	}
	if err := cursor.Click(termsEl); err != nil {
		return fmt.Errorf("register: click terms checkbox: %w", err)
	}
	time.Sleep(time.Duration(registerPostTermsBaseMs+rand.IntN(registerPostTermsJitterMs)) * time.Millisecond)
	cursor.PressKey(input.Tab)
	return nil
}

func (a *CrawlerAgent) registerSubmit(page *rod.Page, cursor *human.Cursor, account *models.Account) error {
	logger.Debug().Int("agent_id", a.ID).Msg("waiting for Verify Email button to become clickable")

	verifyBtn, err := page.Timeout(registerVerifyBtnTimeout).Element(`button[type="submit"][class*="_registerButton_"]:not([disabled])`)
	if err != nil {
		logger.Error().Int("agent_id", a.ID).Err(err).Msg("Verify Email button did not become clickable in time")
		return fmt.Errorf("register: Verify Email button timed out: %w", err)
	}

	logger.Info().Int("agent_id", a.ID).Msg("Verify Email button is now clickable, clicking it")

	if err := cursor.Click(verifyBtn); err != nil {
		return fmt.Errorf("register: click Verify Email button: %w", err)
	}

	logger.Info().
		Int("agent_id", a.ID).
		Str("email", account.Email).
		Str("first_name", account.FirstName).
		Str("last_name", account.LastName).
		Msg("registration submitted, Verify Email clicked")

	return nil
}

func (a *CrawlerAgent) registerAwaitVerificationPage(page *rod.Page, account *models.Account) error {
	logger.Debug().Int("agent_id", a.ID).Msg("waiting for email verification page to load")

	if _, err := page.Timeout(registerVerifyPageTimeout).Element(`div[aria-label="Verify check email page"]`); err != nil {
		logger.Error().Int("agent_id", a.ID).Err(err).Msg("email verification page did not appear in time")
		return fmt.Errorf("register: email verification page timed out: %w", err)
	}

	logger.Info().Int("agent_id", a.ID).Msg("email verification page loaded, waiting for OTP code input")

	if _, err := page.Timeout(registerOTPInputTimeout).Element(`input[data-radix-otp-input][data-radix-index="0"]`); err != nil {
		logger.Error().Int("agent_id", a.ID).Err(err).Msg("OTP code input not found")
		return fmt.Errorf("register: OTP code input not found: %w", err)
	}

	if _, err := page.Timeout(registerContinueBtnTimeout).Element(`button[type="submit"][class*="_continueButton_"]`); err != nil {
		logger.Error().Int("agent_id", a.ID).Err(err).Msg("Continue button not found")
		return fmt.Errorf("register: Continue button not found: %w", err)
	}

	logger.Info().
		Int("agent_id", a.ID).
		Str("email", account.Email).
		Msg("email verification page ready, awaiting 6-digit code entry")

	return nil
}

func (a *CrawlerAgent) registerEnterCode(page *rod.Page, cursor *human.Cursor, account *models.Account, codeCh <-chan registerCodeResult) error {
	logger.Debug().Int("agent_id", a.ID).Msg("waiting for verification code from catcher")

	cr := <-codeCh
	if cr.err != nil {
		return fmt.Errorf("register: wait for code: %w", cr.err)
	}

	logger.Debug().Int("agent_id", a.ID).Str("code", cr.code).Msg("entering verification code")

	otpInput, err := page.Element(`input[data-radix-otp-input][data-radix-index="0"]`)
	if err != nil {
		return fmt.Errorf("register: find OTP input: %w", err)
	}
	if err := cursor.Click(otpInput); err != nil {
		return fmt.Errorf("register: click OTP input: %w", err)
	}
	cursor.Type(cr.code)

	continueBtn, err := page.Element(`button[type="submit"][class*="_continueButton_"]`)
	if err != nil {
		return fmt.Errorf("register: find Continue button: %w", err)
	}
	if err := cursor.Click(continueBtn); err != nil {
		return fmt.Errorf("register: click Continue button: %w", err)
	}

	logger.Info().
		Int("agent_id", a.ID).
		Str("email", account.Email).
		Msg("verification code entered and submitted")

	return nil
}
