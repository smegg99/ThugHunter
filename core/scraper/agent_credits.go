// core/scraper/agent_credits.go
package scraper

import (
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
	Current int `json:"current"`
	Total   int `json:"total"`
}

const (
	creditsPageTimeout    = 30 * time.Second
	creditsProgressBarSel = `[data-testid="credit-usage-progress-bar"]`
)

func (a *ScraperAgent) UpdateCredits() (*ScraperCredits, error) {
	if !a.IsLoggedIn() {
		logger.Warn().Str("agent", a.Name).Msg("update credits: agent not logged in, logging in first")
		if err := a.Login(nil); err != nil {
			return nil, fmt.Errorf("update credits: login: %w", err)
		}
	}

	logger.Info().Str("agent", a.Name).Msg("updating credits")

	a.SetStatus(AgentStatusBusy)

	page, err := a.creditsOpenPage()
	if err != nil {
		a.SetStatus(AgentStatusError)
		return nil, err
	}
	defer page.MustClose()

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

	logger.Info().
		Str("agent", a.Name).
		Int("current", credits.Current).
		Int("total", credits.Total).
		Msg("credits parsed")

	if a.account != nil {
		a.account.CreditsAmount = uint(credits.Current)
	}

	a.SetStatus(AgentStatusIdle)
	return credits, nil
}

func (a *ScraperAgent) creditsOpenPage() (*rod.Page, error) {
	page, _, err := a.newTab(config.Get().Scraper.Endpoints.SettingsBillingEndpoint)
	if err != nil {
		return nil, fmt.Errorf("update credits: %w", err)
	}
	return page, nil
}

func (a *ScraperAgent) creditsAwaitProgressBar(page *rod.Page) error {
	logger.Debug().Str("agent", a.Name).Msg("waiting for credit usage progress bar")

	_, err := page.Timeout(creditsPageTimeout).Element(creditsProgressBarSel)
	if err != nil {
		return fmt.Errorf("update credits: credit progress bar not found: %w", err)
	}

	logger.Debug().Str("agent", a.Name).Msg("credit usage progress bar found")
	return nil
}

func (a *ScraperAgent) creditsExtractHeading(page *rod.Page) (string, error) {
	logger.Debug().Str("agent", a.Name).Msg("extracting credit heading text")

	bar, err := page.Element(creditsProgressBarSel)
	if err != nil {
		return "", fmt.Errorf("update credits: read credit heading: progress bar not found: %w", err)
	}

	card, err := bar.Parent()
	if err != nil {
		return "", fmt.Errorf("update credits: read credit heading: parent not found: %w", err)
	}

	h3, err := card.Element("h3")
	if err != nil {
		return "", fmt.Errorf("update credits: read credit heading: h3 not found: %w", err)
	}

	text, err := h3.Text()
	if err != nil {
		return "", fmt.Errorf("update credits: read credit heading: text extraction failed: %w", err)
	}

	text = strings.TrimSpace(text)
	logger.Debug().Str("agent", a.Name).Str("text", text).Msg("credit heading extracted")
	return text, nil
}

var creditsRe = regexp.MustCompile(`(\d+)\s*/\s*(\d+)\s*available`)

func parseCredits(text string) (*ScraperCredits, error) {
	matches := creditsRe.FindStringSubmatch(text)
	if len(matches) != 3 {
		return nil, fmt.Errorf("parse credits: unexpected format %q", text)
	}

	current, err := strconv.Atoi(matches[1])
	if err != nil {
		return nil, fmt.Errorf("parse credits current: %w", err)
	}

	total, err := strconv.Atoi(matches[2])
	if err != nil {
		return nil, fmt.Errorf("parse credits total: %w", err)
	}

	return &ScraperCredits{
		Current: current,
		Total:   total,
	}, nil
}
