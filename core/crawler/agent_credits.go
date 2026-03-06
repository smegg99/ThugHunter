// core/crawler/agent_credits.go
package crawler

import (
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/go-rod/rod"

	"smegg.me/thughunter/common/config"
	"smegg.me/thughunter/common/logger"
)

const (
	creditsPageTimeout    = 30 * time.Second
	creditsProgressBarSel = `[data-testid="credit-usage-progress-bar"]`
)

func (a *CrawlerAgent) UpdateCredits() (*CrawlerCredits, error) {
	if !a.loggedIn {
		return nil, fmt.Errorf("update credits: agent is not logged in")
	}

	logger.Info().Int("agent_id", a.ID).Msg("updating credits")

	page, err := a.creditsOpenPage()
	if err != nil {
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
		Int("agent_id", a.ID).
		Int("current", credits.Current).
		Int("total", credits.Total).
		Msg("credits parsed")

	if a.account != nil {
		a.account.CreditsAmount = uint(credits.Current)
	}

	return credits, nil
}

func (a *CrawlerAgent) creditsOpenPage() (*rod.Page, error) {
	page, _, err := a.newTab(config.Get().Crawler.Endpoints.SettingsBillingEndpoint)
	if err != nil {
		return nil, fmt.Errorf("update credits: %w", err)
	}
	return page, nil
}

func (a *CrawlerAgent) creditsAwaitProgressBar(page *rod.Page) error {
	logger.Debug().Int("agent_id", a.ID).Msg("waiting for credit usage progress bar")

	_, err := page.Timeout(creditsPageTimeout).Element(creditsProgressBarSel)
	if err != nil {
		return fmt.Errorf("update credits: credit progress bar not found: %w", err)
	}

	logger.Debug().Int("agent_id", a.ID).Msg("credit usage progress bar found")
	return nil
}

func (a *CrawlerAgent) creditsExtractHeading(page *rod.Page) (string, error) {
	logger.Debug().Int("agent_id", a.ID).Msg("extracting credit heading text")

	result, err := page.Eval(`() => {
		const bar = document.querySelector('[data-testid="credit-usage-progress-bar"]');
		if (!bar) return '';
		const card = bar.parentElement;
		const h3 = card ? card.querySelector('h3') : null;
		return h3 ? h3.textContent.trim() : '';
	}`)
	if err != nil {
		return "", fmt.Errorf("update credits: read credit heading: %w", err)
	}

	text := result.Value.Str()
	logger.Debug().Int("agent_id", a.ID).Str("text", text).Msg("credit heading extracted")
	return text, nil
}

var creditsRe = regexp.MustCompile(`(\d+)\s*/\s*(\d+)\s*available`)

func parseCredits(text string) (*CrawlerCredits, error) {
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

	return &CrawlerCredits{
		Current: current,
		Total:   total,
	}, nil
}
