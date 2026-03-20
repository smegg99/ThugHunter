// core/scraper/agent_register.go
package scraper

import (
	"context"
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/input"

	"github.com/smegg99/human"
	"smegg.me/thughunter/common/config"
	"smegg.me/thughunter/common/i18n"
	"smegg.me/thughunter/common/logger"
	"smegg.me/thughunter/core/catcher"
	"smegg.me/thughunter/core/models"
	"smegg.me/thughunter/core/repositories"
)

const (
	registerPageScanBaseMs     = 50
	registerPageScanJitterMs   = 100
	registerPostTermsBaseMs    = 50
	registerPostTermsJitterMs  = 100
	registerVerifyBtnTimeout   = 25 * time.Second
	registerVerifyPageTimeout  = 25 * time.Second
	registerOTPInputTimeout    = 15 * time.Second
	registerContinueBtnTimeout = 15 * time.Second
	registerCodeTimeout        = 15 * time.Minute
)

// Register creates a new account, fills the form, and verifies the email.
func (a *ScraperAgent) Register(ctx context.Context) (*models.Account, error) {
	logger.Info().Str("agent", a.Name).Msg("registering account")

	a.SetStatusText(i18n.T("agent.creatingTempAccount"))
	account, err := a.registerCreateTempAccount()
	if err != nil {
		return nil, err
	}

	mc, err := catcher.New()
	if err != nil {
		return nil, fmt.Errorf("register: start catcher: %w", err)
	}
	defer mc.Close()

	codeCh := a.registerListenForCode(mc, account.Email)

	a.SetStatusText(i18n.T("agent.openingRegistrationPage"))
	page, cursor, err := a.registerOpenPage(ctx)
	if err != nil {
		return nil, err
	}

	a.SetStatusText(i18n.T("agent.fillingRegistrationForm"))
	if err := a.registerFillForm(page, cursor, account); err != nil {
		return nil, err
	}

	a.SetStatusText(i18n.T("agent.acceptingTerms"))
	if err := a.registerAcceptTerms(ctx, page, cursor); err != nil {
		return nil, err
	}

	a.SetStatusText(i18n.T("agent.submittingRegistration"))
	if err := a.registerSubmit(page, cursor, account); err != nil {
		return nil, err
	}

	a.SetStatusText(i18n.T("agent.waitingForVerificationPage"))
	if err := a.registerAwaitVerificationPage(page, account); err != nil {
		return nil, err
	}

	a.SetStatusText(i18n.T("agent.waitingForVerificationCode"))
	if err := a.registerEnterCode(ctx, page, cursor, account, codeCh); err != nil {
		return nil, err
	}

	a.SetStatusText(i18n.T("agent.savingRegisteredAccount"))
	if err := a.registerCreateAccount(account); err != nil {
		return nil, err
	}

	logger.Info().Str("agent", a.Name).Str("email", account.Email).Msg("registered account saved to DB")
	return account, nil
}

type registerCodeResult struct {
	code string
	err  error
}

// registerCreateAccount saves the newly registered account to the database.
func (s *ScraperAgent) registerCreateAccount(account *models.Account) error {
	accounts := repositories.GetAccountRepository()
	if err := accounts.Create(account); err != nil {
		logger.Error().Err(err).Str("agent", s.Name).Str("email", account.Email).Msg("failed to save registered account to DB")
		return fmt.Errorf("register: save account: %w", err)
	}
	return nil
}

// registerCreateTempAccount creates a temporary account with random credentials.
func (a *ScraperAgent) registerCreateTempAccount() (*models.Account, error) {
	accountID := fmt.Sprintf("%s-%d", a.Name, time.Now().UnixMilli())
	account, err := newAccountFromTemplates(accountID)
	if err != nil {
		return nil, fmt.Errorf("register: create account: %w", err)
	}
	return account, nil
}

// registerListenForCode starts a goroutine that waits for the verification code.
func (a *ScraperAgent) registerListenForCode(mc *catcher.Catcher, email string) <-chan registerCodeResult {
	ch := make(chan registerCodeResult, 1)
	go func() {
		code, err := mc.WaitForCode(email, registerCodeTimeout)
		ch <- registerCodeResult{code, err}
	}()
	return ch
}

// registerOpenPage opens the registration page.
func (a *ScraperAgent) registerOpenPage(ctx context.Context) (*rod.Page, *human.Cursor, error) {
	ep := config.Get().Scraper.Endpoints
	page, cursor, err := a.openPageWithJitter(ctx, ep.RegisterEndpoint, registerPageScanBaseMs, registerPageScanJitterMs)
	if err != nil {
		return nil, nil, fmt.Errorf("register: %w", err)
	}
	return page, cursor, nil
}

// registerFillForm fills the registration form fields with the account data.
func (a *ScraperAgent) registerFillForm(page *rod.Page, cursor *human.Cursor, account *models.Account) error {
	return a.fillFormFields(page, cursor, "register", []formField{
		{"#first_name", account.FirstName, "first name"},
		{"#last_name", account.LastName, "last name"},
		{"#identifier", account.Email, "email"},
		{"#org_name", account.Organization, "organization"},
		{"#password", account.Password, "password"},
		{"#confirm-password", account.Password, "confirm password"},
	})
}

// registerAcceptTerms ticks the terms checkbox and presses Tab.
func (a *ScraperAgent) registerAcceptTerms(ctx context.Context, page *rod.Page, cursor *human.Cursor) error {
	logger.Debug().Str("agent", a.Name).Msg("ticking terms checkbox")
	termsEl, err := page.Element("#terms-and-conditions")
	if err != nil {
		return fmt.Errorf("register: terms checkbox: %w: %w", ErrElementNotFound, err)
	}
	if err := cursor.Click(termsEl); err != nil {
		return fmt.Errorf("register: click terms checkbox: %w", err)
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(time.Duration(registerPostTermsBaseMs+rand.IntN(registerPostTermsJitterMs)) * time.Millisecond):
	}
	cursor.PressKey(input.Tab)
	return nil
}

// registerSubmit clicks the "Verify Email" button.
func (a *ScraperAgent) registerSubmit(page *rod.Page, cursor *human.Cursor, account *models.Account) error {
	logger.Debug().Str("agent", a.Name).Msg("waiting for Verify Email button to become clickable")

	verifyBtn, err := awaitElement(page, `button[type="submit"][class*="_registerButton_"]:not([disabled])`, registerVerifyBtnTimeout, "register: Verify Email button")
	if err != nil {
		logger.Error().Str("agent", a.Name).Err(err).Msg("Verify Email button did not become clickable in time")
		return err
	}

	logger.Info().Str("agent", a.Name).Msg("Verify Email button is now clickable, clicking it")

	if err := cursor.Click(verifyBtn); err != nil {
		return fmt.Errorf("register: click Verify Email button: %w", err)
	}

	logger.Info().
		Str("agent", a.Name).
		Str("email", account.Email).
		Str("first_name", account.FirstName).
		Str("last_name", account.LastName).
		Msg("registration submitted, Verify Email clicked")

	return nil
}

// registerAwaitVerificationPage waits for the email verification page to load.
func (a *ScraperAgent) registerAwaitVerificationPage(page *rod.Page, account *models.Account) error {
	logger.Debug().Str("agent", a.Name).Msg("waiting for email verification page to load")

	if _, err := awaitElement(page, `div[aria-label="Verify check email page"]`, registerVerifyPageTimeout, "register: verification page"); err != nil {
		logger.Error().Str("agent", a.Name).Err(err).Msg("email verification page did not appear in time")
		return err
	}

	logger.Info().Str("agent", a.Name).Msg("email verification page loaded, waiting for OTP code input")

	if _, err := awaitElement(page, `input[data-radix-otp-input][data-radix-index="0"]`, registerOTPInputTimeout, "register: OTP code input"); err != nil {
		logger.Error().Str("agent", a.Name).Err(err).Msg("OTP code input not found")
		return err
	}

	if _, err := awaitElement(page, `button[type="submit"][class*="_continueButton_"]`, registerContinueBtnTimeout, "register: Continue button"); err != nil {
		logger.Error().Str("agent", a.Name).Err(err).Msg("Continue button not found")
		return err
	}

	logger.Info().
		Str("agent", a.Name).
		Str("email", account.Email).
		Msg("email verification page ready, awaiting 6-digit code entry")

	return nil
}

// registerEnterCode enters and submits the verification code.
func (a *ScraperAgent) registerEnterCode(ctx context.Context, page *rod.Page, cursor *human.Cursor, account *models.Account, codeCh <-chan registerCodeResult) error {
	logger.Debug().Str("agent", a.Name).Msg("waiting for verification code from catcher")

	var cr registerCodeResult
	select {
	case <-ctx.Done():
		return ctx.Err()
	case cr = <-codeCh:
	}
	if cr.err != nil {
		return fmt.Errorf("register: %w: %w", ErrVerificationFailed, cr.err)
	}

	logger.Debug().Str("agent", a.Name).Str("code", cr.code).Msg("entering verification code")

	otpInput, err := page.Element(`input[data-radix-otp-input][data-radix-index="0"]`)
	if err != nil {
		return fmt.Errorf("register: OTP input: %w: %w", ErrElementNotFound, err)
	}
	if err := cursor.Click(otpInput); err != nil {
		return fmt.Errorf("register: click OTP input: %w", err)
	}
	cursor.Type(cr.code)

	continueBtn, err := page.Element(`button[type="submit"][class*="_continueButton_"]`)
	if err != nil {
		return fmt.Errorf("register: Continue button: %w: %w", ErrElementNotFound, err)
	}
	if err := cursor.Click(continueBtn); err != nil {
		return fmt.Errorf("register: click Continue button: %w", err)
	}

	logger.Info().
		Str("agent", a.Name).
		Str("email", account.Email).
		Msg("verification code entered and submitted")

	return nil
}
