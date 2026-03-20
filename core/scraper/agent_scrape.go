// core/scraper/agent_scrape.go
package scraper

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/go-rod/rod"

	"github.com/smegg99/human"
	"smegg.me/thughunter/common/config"
	"smegg.me/thughunter/common/logger"
	"smegg.me/thughunter/core/models"
)

const searchResultsTimeout = 60 * time.Second

// outOfCreditsBannerSel matches the Censys "You have run out of credits" banner.
const outOfCreditsBannerSel = `div[class*="broadcastBanner"][class*="error"]`

// checkOutOfCreditsBanner looks for the credits-exhaustion banner on the current page.
// Returns ErrRanOutOfCredits if found, nil otherwise.
func (a *ScraperAgent) checkOutOfCreditsBanner(page *rod.Page) error {
	has, _, err := page.Has(outOfCreditsBannerSel)
	if err != nil {
		return nil // element lookup failed, not a credit issue
	}
	if has {
		logger.Warn().Str("agent", a.Name).Msg("out-of-credits banner detected")
		return ErrRanOutOfCredits
	}
	return nil
}

// Scrape runs a search query and extracts hosts from the results.
func (a *ScraperAgent) Scrape(ctx context.Context, queryString string) ([]*models.Host, error) {
	if !a.IsLoggedIn() {
		return nil, fmt.Errorf("scrape: %w", ErrAgentNotLoggedIn)
	}

	logger.Info().Str("agent", a.Name).Str("query", queryString).Msg("scraping search results")

	page, _, err := a.searchOpenPage(ctx, queryString)
	if err != nil {
		return nil, err
	}

	// Check for credits banner before waiting for results.
	if err := a.checkOutOfCreditsBanner(page); err != nil {
		return nil, err
	}

	if err := a.searchAwaitResults(page); err != nil {
		return nil, err
	}

	// Check again after results load (banner can appear late).
	if err := a.checkOutOfCreditsBanner(page); err != nil {
		return nil, err
	}
	hosts, err := a.searchExtractHosts(page)
	if err != nil {
		return nil, err
	}

	logger.Info().Str("agent", a.Name).Int("count", len(hosts)).Msg("hosts extracted from search results")

	return hosts, nil
}

// searchOpenPage navigates to the search page with the query parameter set.
func (a *ScraperAgent) searchOpenPage(ctx context.Context, queryString string) (*rod.Page, *human.Cursor, error) {
	searchURL := config.Get().Scraper.Endpoints.SearchEndpoint

	u, err := url.Parse(searchURL)
	if err != nil {
		return nil, nil, fmt.Errorf("search: %w: %w", ErrInvalidURL, err)
	}

	q := u.Query()
	q.Set("q", queryString)
	u.RawQuery = q.Encode()

	logger.Debug().Str("agent", a.Name).Str("url", u.String()).Msg("opening search page")

	page, cursor, err := a.newTab(ctx, u.String())
	if err != nil {
		return nil, nil, fmt.Errorf("search: %w", err)
	}
	return page, cursor, nil
}

// searchAwaitResults waits for the search results page to fully load.
func (a *ScraperAgent) searchAwaitResults(page *rod.Page) error {
	logger.Debug().Str("agent", a.Name).Msg("waiting for search results to load")

	if _, err := awaitElement(page, `[data-search-results="true"]`, searchResultsTimeout, "search: results container"); err != nil {
		return err
	}
	if _, err := awaitElement(page, `h4[aria-label$=" results"]`, searchResultsTimeout, "search: result count"); err != nil {
		return err
	}
	if _, err := awaitElement(page, `h4[aria-label^="The query took"]`, searchResultsTimeout, "search: query duration"); err != nil {
		return err
	}

	logger.Debug().Str("agent", a.Name).Msg("search results loaded")
	return nil
}

// searchExtractHosts finds all host cards on the search results page and extracts structured host data from each card.
func (a *ScraperAgent) searchExtractHosts(page *rod.Page) ([]*models.Host, error) {
	logger.Debug().Str("agent", a.Name).Msg("extracting host data from search results")

	cards, err := page.Elements(`[data-testid="hostDetailsCard"]`)
	if err != nil {
		return nil, fmt.Errorf("search: host cards: %w: %w", ErrElementNotFound, err)
	}

	hosts := make([]*models.Host, 0, len(cards))
	for _, card := range cards {
		host, err := a.searchParseHostCard(card)
		if err != nil {
			logger.Warn().Str("agent", a.Name).Err(err).Msg("skipping host card")
			continue
		}
		hosts = append(hosts, host)
	}

	return hosts, nil
}
