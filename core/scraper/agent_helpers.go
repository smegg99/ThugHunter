// core/scraper/agent_helpers.go
package scraper

import (
	"context"
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/go-rod/rod"

	"github.com/smegg99/human"
	"smegg.me/thughunter/common/logger"
)

// formField describes a single form input to fill.
type formField struct {
	selector string
	value    string
	label    string
}

const formFieldTimeout = 30 * time.Second

// fillFormFields fills a list of form fields on the page using the cursor.
// Each element lookup is bounded by formFieldTimeout to prevent indefinite hangs
// (e.g. when a login page redirects to the homepage and the form never appears).
func (a *ScraperAgent) fillFormFields(page *rod.Page, cursor *human.Cursor, context string, fields []formField) error {
	for _, f := range fields {
		logger.Debug().Str("agent", a.Name).Str("field", f.label).Msg("filling field")
		el, err := awaitElement(page, f.selector, formFieldTimeout, fmt.Sprintf("%s: %s", context, f.label))
		if err != nil {
			return err
		}
		if err := cursor.ClickAndType(el, f.value); err != nil {
			return fmt.Errorf("%s: fill %s: %w", context, f.label, err)
		}
	}
	return nil
}

// openPageWithJitter opens a new tab and waits a random duration to simulate scanning.
func (a *ScraperAgent) openPageWithJitter(ctx context.Context, url string, baseMs, jitterMs int) (*rod.Page, *human.Cursor, error) {
	page, cursor, err := a.newTab(ctx, url)
	if err != nil {
		return nil, nil, err
	}
	select {
	case <-ctx.Done():
		return nil, nil, ctx.Err()
	case <-time.After(time.Duration(baseMs+rand.IntN(jitterMs)) * time.Millisecond):
	}
	return page, cursor, nil
}

// awaitElement waits for an element matching selector to appear within the timeout.
func awaitElement(page *rod.Page, selector string, timeout time.Duration, context string) (*rod.Element, error) {
	el, err := page.Timeout(timeout).Element(selector)
	if err != nil {
		return nil, fmt.Errorf("%s: %w: %w", context, ErrTimeout, err)
	}
	return el, nil
}
