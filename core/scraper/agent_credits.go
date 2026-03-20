// core/scraper/agent_credits.go
package scraper

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-rod/rod"

	"smegg.me/thughunter/common/config"
	"smegg.me/thughunter/common/logger"
)

type ScraperCredits struct {
	Current    int        `json:"current"`
	MaxCredits int        `json:"max_credits"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
}

const (
	creditsPageTimeout       = 30 * time.Second
	creditsProgressBarSel    = `[data-testid="credit-usage-progress-bar"]`
	creditsExpirationDateSel = `[data-testid="fmt-timestamp-date"]`
)

var creditsRe = regexp.MustCompile(`(\d+)\s*/\s*(\d+)\s*available`)

// UpdateCredits fetches the credit balance from the billing page.
// The caller must ensure the agent is already logged in.
func (a *ScraperAgent) UpdateCredits(ctx context.Context) (*ScraperCredits, error) {
	logger.Info().Str("agent", a.Name).Msg("updating credits")

	page, err := a.creditsOpenPage(ctx)
	if err != nil {
		return nil, err
	}
	if err := a.creditsAwaitProgressBar(page); err != nil {
		return nil, err
	}

	text, err := a.creditsExtractHeading(page)
	if err != nil {
		return nil, err
	}

	credits, err := parseCredits(text)
	if err != nil {
		return nil, fmt.Errorf("update credits: %w", err)
	}

	expiresAt, err := a.creditsExtractExpirationDate(page)
	if err != nil {
		logger.Warn().Err(err).Str("agent", a.Name).Msg("failed to extract credit expiration date")
	} else {
		credits.ExpiresAt = expiresAt
	}

	logger.Info().
		Str("agent", a.Name).
		Int("current", credits.Current).
		Int("max_credits", credits.MaxCredits).
		Msg("credits parsed")

	a.estimatedCredits = credits.Current

	if a.account != nil {
		a.account.CreditsAmount = uint(credits.Current)
		a.account.CreditsExpireAt = credits.ExpiresAt
		now := time.Now()
		a.account.RefreshedCreditsAt = &now

		if uint(credits.Current) < CreditsAmountPerQuery {
			a.account.RanOutOfCreditsAt = &now
			logger.Warn().
				Str("agent", a.Name).
				Str("email", a.account.Email).
				Int("credits", credits.Current).
				Msg("account ran out of credits")
		}
	}

	return credits, nil
}

func (a *ScraperAgent) creditsOpenPage(ctx context.Context) (*rod.Page, error) {
	page, _, err := a.newTab(ctx, config.Get().Scraper.Endpoints.SettingsBillingEndpoint)
	if err != nil {
		return nil, fmt.Errorf("update credits: %w", err)
	}
	return page, nil
}

func (a *ScraperAgent) creditsAwaitProgressBar(page *rod.Page) error {
	logger.Debug().Str("agent", a.Name).Msg("waiting for credit usage progress bar")

	if _, err := awaitElement(page, creditsProgressBarSel, creditsPageTimeout, "update credits: progress bar"); err != nil {
		return err
	}

	logger.Debug().Str("agent", a.Name).Msg("credit usage progress bar found")
	return nil
}

func (a *ScraperAgent) creditsExtractHeading(page *rod.Page) (string, error) {
	logger.Debug().Str("agent", a.Name).Msg("extracting credit heading text")

	bar, err := page.Element(creditsProgressBarSel)
	if err != nil {
		return "", fmt.Errorf("update credits: progress bar: %w: %w", ErrElementNotFound, err)
	}

	card, err := bar.Parent()
	if err != nil {
		return "", fmt.Errorf("update credits: parent card: %w: %w", ErrElementNotFound, err)
	}

	h3, err := card.Element("h3")
	if err != nil {
		return "", fmt.Errorf("update credits: heading: %w: %w", ErrElementNotFound, err)
	}

	text, err := h3.Text()
	if err != nil {
		return "", fmt.Errorf("update credits: heading text: %w: %w", ErrParseFailed, err)
	}

	text = strings.TrimSpace(text)
	logger.Debug().Str("agent", a.Name).Str("text", text).Msg("credit heading extracted")
	return text, nil
}

// creditsExtractExpirationDate parses the expiration date from the billing page.
func (a *ScraperAgent) creditsExtractExpirationDate(page *rod.Page) (*time.Time, error) {
	logger.Debug().Str("agent", a.Name).Msg("extracting credit expiration date")

	el, err := page.Element(creditsExpirationDateSel)
	if err != nil {
		return nil, fmt.Errorf("update credits: expiration date: %w: %w", ErrElementNotFound, err)
	}

	text, err := el.Text()
	if err != nil {
		return nil, fmt.Errorf("update credits: expiration date text: %w: %w", ErrParseFailed, err)
	}

	text = strings.TrimSpace(text)
	parsed, err := time.Parse("Jan 2, 2006", text)
	if err != nil {
		return nil, fmt.Errorf("update credits: parse expiration date %q: %w", text, err)
	}

	logger.Debug().Str("agent", a.Name).Time("expires_at", parsed).Msg("credit expiration date extracted")
	return &parsed, nil
}

func parseCredits(text string) (*ScraperCredits, error) {
	matches := creditsRe.FindStringSubmatch(text)
	if len(matches) != 3 {
		return nil, fmt.Errorf("%w: unexpected format %q", ErrParseFailed, text)
	}

	current, err := strconv.Atoi(matches[1])
	if err != nil {
		return nil, fmt.Errorf("%w: current credits: %w", ErrParseFailed, err)
	}

	maxCredits, err := strconv.Atoi(matches[2])
	if err != nil {
		return nil, fmt.Errorf("%w: max credits: %w", ErrParseFailed, err)
	}

	return &ScraperCredits{
		Current:    current,
		MaxCredits: maxCredits,
	}, nil
}
