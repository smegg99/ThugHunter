// core/scraper/agent_register.go
package scraper

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
	"smegg.me/thughunter/core/repositories"
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

// Register creates a new account, fills the registration form, and verifies the email.
func (a *ScraperAgent) Register() (*models.Account, error) {
	logger.Info().Str("agent", a.Name).Msg("registering account")

	return runTaskResult(a, func() (*models.Account, error) {
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

		accounts := repositories.GetAccountRepository()
		if err := accounts.Create(account); err != nil {
			logger.Error().Err(err).Str("agent", a.Name).Str("email", account.Email).Msg("failed to save registered account to DB")
			return nil, fmt.Errorf("register: save account: %w", err)
		}

		logger.Info().Str("agent", a.Name).Str("email", account.Email).Msg("registered account saved to DB")
		return account, nil
	})
}

type registerCodeResult struct {
	code string
	err  error
}

func (a *ScraperAgent) registerCreateAccount() (*models.Account, error) {
	accountID := fmt.Sprintf("%s-%d", a.Name, time.Now().UnixMilli())
	account, err := newAccountFromTemplates(accountID)
	if err != nil {
		return nil, fmt.Errorf("register: create account: %w", err)
	}
	return account, nil
}

func (a *ScraperAgent) registerListenForCode(mc *catcher.Catcher, email string) <-chan registerCodeResult {
	ch := make(chan registerCodeResult, 1)
	go func() {
		code, err := mc.WaitForCode(email, registerCodeTimeout)
		ch <- registerCodeResult{code, err}
	}()
	return ch
}

func (a *ScraperAgent) registerOpenPage() (*rod.Page, *human.Cursor, error) {
	ep := config.Get().Scraper.Endpoints
	page, cursor, err := a.openPageWithJitter(ep.RegisterEndpoint, registerPageScanBaseMs, registerPageScanJitterMs)
	if err != nil {
		return nil, nil, fmt.Errorf("register: %w", err)
	}
	return page, cursor, nil
}

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

func (a *ScraperAgent) registerAcceptTerms(page *rod.Page, cursor *human.Cursor) error {
	logger.Debug().Str("agent", a.Name).Msg("ticking terms checkbox")
	termsEl, err := page.Element("#terms-and-conditions")
	if err != nil {
		return fmt.Errorf("register: terms checkbox: %w: %w", ErrElementNotFound, err)
	}
	if err := cursor.Click(termsEl); err != nil {
		return fmt.Errorf("register: click terms checkbox: %w", err)
	}
	time.Sleep(time.Duration(registerPostTermsBaseMs+rand.IntN(registerPostTermsJitterMs)) * time.Millisecond)
	cursor.PressKey(input.Tab)
	return nil
}

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

func (a *ScraperAgent) registerEnterCode(page *rod.Page, cursor *human.Cursor, account *models.Account, codeCh <-chan registerCodeResult) error {
	logger.Debug().Str("agent", a.Name).Msg("waiting for verification code from catcher")

	cr := <-codeCh
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
