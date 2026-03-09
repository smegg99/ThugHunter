// core/scraper/agent_scrape_parse.go
package scraper

import (
	"fmt"
	"strings"

	"github.com/go-rod/rod"

	"smegg.me/thughunter/core/models"
)

// searchParseHostCard extracts all host data from a single hostDetailsCard element.
func (a *ScraperAgent) searchParseHostCard(card *rod.Element) (*models.Host, error) {
	ip, err := searchExtractIP(card)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrParseFailed, err)
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
		return "", fmt.Errorf("IP link: %w: %w", ErrElementNotFound, err)
	}

	title, err := link.Attribute("title")
	if err != nil || title == nil {
		return "", fmt.Errorf("IP link: %w: missing title attribute", ErrParseFailed)
	}

	ip := strings.TrimPrefix(*title, "View ")
	ip = strings.TrimSuffix(ip, " Details")
	return ip, nil
}

// searchExtractDetailTable reads the key-value detail rows from the table inside the card.
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
func searchExtractTagList(card *rod.Element, heading string) []string {
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
