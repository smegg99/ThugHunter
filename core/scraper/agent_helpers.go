// core/scraper/agent_helpers.go
package scraper

import (
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/go-rod/rod"

	"smegg.me/thughunter/common/logger"
	"smegg.me/thughunter/core/human"
)

// formField describes a single form input to fill.
type formField struct {
	selector string
	value    string
	label    string
}

// runTask sets the agent to busy, executes fn, and transitions to idle or error.
func (a *ScraperAgent) runTask(fn func() error) error {
	a.SetStatus(AgentStatusBusy)
	if err := fn(); err != nil {
		a.SetStatus(AgentStatusError)
		return err
	}
	a.SetStatus(AgentStatusIdle)
	return nil
}

// runTaskResult is like runTask but for functions returning a value alongside an error.
func runTaskResult[T any](a *ScraperAgent, fn func() (T, error)) (T, error) {
	a.SetStatus(AgentStatusBusy)
	result, err := fn()
	if err != nil {
		a.SetStatus(AgentStatusError)
		var zero T
		return zero, err
	}
	a.SetStatus(AgentStatusIdle)
	return result, nil
}

// fillFormFields fills a list of form fields on the page using the cursor.
func (a *ScraperAgent) fillFormFields(page *rod.Page, cursor *human.Cursor, context string, fields []formField) error {
	for _, f := range fields {
		logger.Debug().Str("agent", a.Name).Str("field", f.label).Msg("filling field")
		el, err := page.Element(f.selector)
		if err != nil {
			return fmt.Errorf("%s: %s: %w: %w", context, f.label, ErrElementNotFound, err)
		}
		if err := cursor.ClickAndType(el, f.value); err != nil {
			return fmt.Errorf("%s: fill %s: %w", context, f.label, err)
		}
	}
	return nil
}

// openPageWithJitter opens a new tab and waits a random duration to simulate page scanning.
func (a *ScraperAgent) openPageWithJitter(url string, baseMs, jitterMs int) (*rod.Page, *human.Cursor, error) {
	page, cursor, err := a.newTab(url)
	if err != nil {
		return nil, nil, err
	}
	time.Sleep(time.Duration(baseMs+rand.IntN(jitterMs)) * time.Millisecond)
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
