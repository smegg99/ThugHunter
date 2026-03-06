// core/crawler/agent_search.go
package crawler

import (
	"fmt"

	"github.com/go-rod/rod"

	"smegg.me/thughunter/common/config"
	"smegg.me/thughunter/common/logger"
	"smegg.me/thughunter/core/human"
)

func (a *CrawlerAgent) Search(query string) error {
	if !a.loggedIn {
		return fmt.Errorf("search: agent is not logged in")
	}

	logger.Info().Int("agent_id", a.ID).Str("query", query).Msg("searching")

	page, cursor, err := a.searchOpenPage()
	if err != nil {
		return err
	}
	defer page.MustClose()

	_ = cursor // TODO: implement search interaction

	return nil
}

func (a *CrawlerAgent) ScrapeSearch() error {
	if !a.loggedIn {
		return fmt.Errorf("scrape search: agent is not logged in")
	}

	logger.Info().Int("agent_id", a.ID).Msg("scraping search results")

	page, cursor, err := a.searchOpenPage()
	if err != nil {
		return err
	}
	defer page.MustClose()

	_ = cursor // TODO: implement search scraping

	return nil
}

func (a *CrawlerAgent) searchOpenPage() (*rod.Page, *human.Cursor, error) {
	page, cursor, err := a.newTab(config.Get().Crawler.Endpoints.SearchEndpoint)
	if err != nil {
		return nil, nil, fmt.Errorf("search: %w", err)
	}
	return page, cursor, nil
}
