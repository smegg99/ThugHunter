// core/scraper/agent_search.go
package scraper

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/go-rod/rod"

	"smegg.me/thughunter/common/config"
	"smegg.me/thughunter/common/logger"
	"smegg.me/thughunter/core/human"
	"smegg.me/thughunter/core/models"
)

const searchResultsTimeout = 60 * time.Second

func (a *ScraperAgent) Scrape(queryString string) ([]*models.Host, error) {
	if !a.IsLoggedIn() {
		return nil, fmt.Errorf("scrape: agent is not logged in")
	}

	logger.Info().Str("agent", a.Name).Str("query", queryString).Msg("scraping search results")

	a.SetStatus(AgentStatusBusy)

	page, cursor, err := a.searchOpenPage(queryString)
	if err != nil {
		a.SetStatus(AgentStatusError)
		return nil, err
	}
	defer page.MustClose()

	_ = cursor

	if err := a.searchAwaitResults(page); err != nil {
		a.SetStatus(AgentStatusError)
		return nil, err
	}

	hosts, err := a.searchExtractHosts(page)
	if err != nil {
		a.SetStatus(AgentStatusError)
		return nil, err
	}

	logger.Info().Str("agent", a.Name).Int("count", len(hosts)).Msg("hosts extracted from search results")

	a.SetStatus(AgentStatusIdle)
	return hosts, nil
}

func (a *ScraperAgent) searchOpenPage(queryString string) (*rod.Page, *human.Cursor, error) {
	searchURL := config.Get().Scraper.Endpoints.SearchEndpoint

	u, err := url.Parse(searchURL)
	if err != nil {
		return nil, nil, fmt.Errorf("search: parse endpoint URL: %w", err)
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

// searchAwaitResults waits for the search results and duration indicators to
// appear on the page. It uses stable aria-label and data attributes rather than
// CSS class selectors which contain hashes that change on every site rebuild.
func (a *ScraperAgent) searchAwaitResults(page *rod.Page) error {
	logger.Debug().Str("agent", a.Name).Msg("waiting for search results to load")

	if _, err := page.Timeout(searchResultsTimeout).Element(`[data-search-results="true"]`); err != nil {
		return fmt.Errorf("search: results container did not appear: %w", err)
	}

	if _, err := page.Timeout(searchResultsTimeout).Element(`h4[aria-label$=" results"]`); err != nil {
		return fmt.Errorf("search: result count did not appear: %w", err)
	}

	if _, err := page.Timeout(searchResultsTimeout).Element(`h4[aria-label^="The query took"]`); err != nil {
		return fmt.Errorf("search: query duration did not appear: %w", err)
	}

	logger.Debug().Str("agent", a.Name).Msg("search results loaded")
	return nil
}

func (a *ScraperAgent) searchExtractHosts(page *rod.Page) ([]*models.Host, error) {
	logger.Debug().Str("agent", a.Name).Msg("extracting host data from search results")

	cards, err := page.Elements(`[data-testid="hostDetailsCard"]`)
	if err != nil {
		return nil, fmt.Errorf("search: find host cards: %w", err)
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

// searchParseHostCard extracts all host data from a single hostDetailsCard
// element using go-rod's native API. Only stable data-testid, aria-label,
// title attributes and element structure are used — no hashed CSS classes.
func (a *ScraperAgent) searchParseHostCard(card *rod.Element) (*models.Host, error) {
	ip, err := searchExtractIP(card)
	if err != nil {
		return nil, fmt.Errorf("extract IP: %w", err)
	}

	details := searchExtractDetailTable(card)
	labels := searchExtractLabels(card)
	services := searchExtractTagList(card, "Services")
	software := searchExtractTagList(card, "Software")

	host := models.NewHost(ip, details["Location"], details["OS"], details["Network (AS)"])

	host.Labels = make(map[string]string, len(labels))
	for _, l := range labels {
		host.Labels[l] = l
	}

	host.Services = make(map[string]string, len(services))
	for _, s := range services {
		host.Services[s] = s
	}

	host.Software = make(map[string]string, len(software))
	for _, s := range software {
		host.Software[s] = s
	}

	return host, nil
}

// searchExtractIP gets the IP from the "View X.X.X.X Details" link's title attribute.
func searchExtractIP(card *rod.Element) (string, error) {
	link, err := card.Element(`a[title^="View "][title$=" Details"]`)
	if err != nil {
		return "", fmt.Errorf("IP link not found: %w", err)
	}

	title, err := link.Attribute("title")
	if err != nil || title == nil {
		return "", fmt.Errorf("IP link has no title")
	}

	ip := strings.TrimPrefix(*title, "View ")
	ip = strings.TrimSuffix(ip, " Details")
	return ip, nil
}

// searchExtractDetailTable reads the key-value detail rows (OS, Network, Location)
// from the table inside the card. Uses td elements within tr rows.
func searchExtractDetailTable(card *rod.Element) map[string]string {
	details := make(map[string]string)

	rows, err := card.Elements(`table tbody tr`)
	if err != nil {
		return details
	}

	for _, row := range rows {
		cells, err := row.Elements(`td`)
		if err != nil || len(cells) < 2 {
			continue
		}

		label, err := cells[0].Text()
		if err != nil {
			continue
		}

		value, err := cells[1].Text()
		if err != nil {
			continue
		}

		details[strings.TrimSpace(label)] = strings.TrimSpace(value)
	}

	return details
}

// searchExtractLabels collects label text from the data-testid="label-list" links.
func searchExtractLabels(card *rod.Element) []string {
	links, err := card.Elements(`[data-testid="label-list"]`)
	if err != nil {
		return nil
	}

	var labels []string
	for _, link := range links {
		text, err := link.Text()
		if err != nil {
			continue
		}
		if t := strings.TrimSpace(text); t != "" {
			labels = append(labels, t)
		}
	}
	return labels
}

// searchExtractTagList extracts service or software names from the tag list
// section identified by its heading text (e.g. "Services (10)" or "Software (3)").
// It reads the title attribute from the <a> links inside the list container,
// which hold stable values like "80 / HTTP" or "F5 Nginx".
func searchExtractTagList(card *rod.Element, heading string) []string {
	// Find the span whose text starts with the heading (e.g. "Services (10)").
	spans, err := card.Elements(`span`)
	if err != nil {
		return nil
	}

	for _, span := range spans {
		text, err := span.Text()
		if err != nil {
			continue
		}
		text = strings.TrimSpace(text)
		if !strings.HasPrefix(text, heading+" (") {
			continue
		}

		parent, err := span.Parent()
		if err != nil {
			continue
		}

		links, err := parent.Elements(`a[title]`)
		if err != nil {
			continue
		}

		var items []string
		for _, link := range links {
			title, err := link.Attribute("title")
			if err != nil || title == nil || *title == "" {
				continue
			}
			items = append(items, *title)
		}
		return items
	}

	return nil
}
