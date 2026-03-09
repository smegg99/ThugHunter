// core/scraper/agent_scrape.go
package scraper

import (
"fmt"
"net/url"
"time"

"github.com/go-rod/rod"

"smegg.me/thughunter/common/config"
"smegg.me/thughunter/common/logger"
"smegg.me/thughunter/core/human"
"smegg.me/thughunter/core/models"
)

const searchResultsTimeout = 60 * time.Second

// Scrape runs a search query and extracts host data from the results page.
func (a *ScraperAgent) Scrape(queryString string) ([]*models.Host, error) {
	if !a.IsLoggedIn() {
		return nil, fmt.Errorf("scrape: %w", ErrAgentNotLoggedIn)
	}

	logger.Info().Str("agent", a.Name).Str("query", queryString).Msg("scraping search results")

	return runTaskResult(a, func() ([]*models.Host, error) {
		page, _, err := a.searchOpenPage(queryString)
		if err != nil {
			return nil, err
		}
		defer page.MustClose()

		if err := a.searchAwaitResults(page); err != nil {
			return nil, err
		}

		hosts, err := a.searchExtractHosts(page)
		if err != nil {
			return nil, err
		}

		logger.Info().Str("agent", a.Name).Int("count", len(hosts)).Msg("hosts extracted from search results")

		a.UpdateCredits() // Update credits after each search.

		return hosts, nil
	})
}

func (a *ScraperAgent) searchOpenPage(queryString string) (*rod.Page, *human.Cursor, error) {
	searchURL := config.Get().Scraper.Endpoints.SearchEndpoint

	u, err := url.Parse(searchURL)
	if err != nil {
		return nil, nil, fmt.Errorf("search: %w: %w", ErrInvalidURL, err)
	}

	q := u.Query()
	q.Set("q", queryString)
	u.RawQuery = q.Encode()

	logger.Debug().Str("agent", a.Name).Str("url", u.String()).Msg("opening search page")

	page, cursor, err := a.newTab(u.String())
	if err != nil {
		return nil, nil, fmt.Errorf("search: %w", err)
	}
	return page, cursor, nil
}

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
