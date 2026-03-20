// core/scraper/agent_page.go
package scraper

import (
	"context"
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"github.com/smegg99/human"
	"github.com/smegg99/unrevealed"
	"smegg.me/thughunter/common/logger"
)

// newTab navigates the agent's persistent tab to url, creating one on first
// call. Returns the context-bound page and a human-like cursor.
func (a *ScraperAgent) newTab(ctx context.Context, url string) (*rod.Page, *human.Cursor, error) {
	if err := a.ensureBrowser(ctx); err != nil {
		return nil, nil, err
	}

	logger.Debug().Str("agent", a.Name).Str("url", url).Msg("navigating tab")

	page, err := a.reuseOrCreatePage()
	if err != nil {
		return nil, nil, err
	}

	if !a.pageReady {
		a.injectPageCleanupScript(page)
		a.pageReady = true
	}

	ctxPage := page.Context(ctx)

	if err := a.navigatePage(ctx, ctxPage, url); err != nil {
		return nil, nil, err
	}

	cursor := human.New(ctxPage, func(c *human.Config) {
		c.Direct = false
		c.Hesitation = 0.1
		c.MicroPause = 0.1
		c.Steadiness = 1.0
		c.ClickHold = [2]int{80, 120}
		c.ClickDwell = [2]int{60, 120}
		c.TypeDelay = [2]int{80, 120}
		c.ThinkPause = 0.3
	})
	return ctxPage, cursor, nil
}

// reuseOrCreatePage returns the agent's persistent page, creating one if needed.
func (a *ScraperAgent) reuseOrCreatePage() (*rod.Page, error) {
	if a.page != nil {
		return a.page, nil
	}

	pages, err := a.browser.Pages()
	if err == nil {
		for _, p := range pages {
			info, _ := p.Info()
			if info != nil && (info.URL == "" || info.URL == "about:blank" ||
				info.URL == "chrome://newtab/" || info.URL == "chrome://new-tab-page/") {
				a.page = p
				logger.Debug().Str("agent", a.Name).Msg("reusing existing blank page")
				return a.page, nil
			}
		}
	}

	page, err := a.browser.Page(proto.TargetCreateTarget{URL: "about:blank"})
	if err != nil {
		if len(pages) > 0 {
			logger.Warn().Str("agent", a.Name).Msg("failed to create new tab, reusing existing page")
			a.page = pages[0]
			return a.page, nil
		}
		return nil, fmt.Errorf("new page: %w", err)
	}
	a.page = page
	return a.page, nil
}

// navigatePage applies stealth and navigates with a random post-load delay.
func (a *ScraperAgent) navigatePage(ctx context.Context, page *rod.Page, url string) error {
	if err := unrevealed.Stealth(page); err != nil {
		return fmt.Errorf("apply stealth: %w", err)
	}

	if err := page.Navigate(url); err != nil {
		return fmt.Errorf("%w: %s: %w", ErrNavigationFailed, url, err)
	}

	if err := page.WaitLoad(); err != nil {
		return fmt.Errorf("%w: wait load %s: %w", ErrNavigationFailed, url, err)
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(time.Duration(1000+rand.IntN(2000)) * time.Millisecond):
	}

	logger.Debug().Str("agent", a.Name).Str("url", url).Msg("page loaded")
	return nil
}

// ClearSession removes all cookies and storage so the next login starts
// from a clean slate without closing the browser.
func (a *ScraperAgent) ClearSession() error {
	if a.page == nil {
		a.account = nil
		a.estimatedCredits = 0
		return nil
	}

	logger.Debug().Str("agent", a.Name).Msg("clearing session")

	resetPage := func() {
		a.page = nil
		a.pageReady = false
		a.account = nil
		a.estimatedCredits = 0
	}

	defer func() {
		if r := recover(); r != nil {
			logger.Warn().Str("agent", a.Name).Interface("panic", r).Msg("recovered panic during session clear, resetting page")
			resetPage()
		}
	}()

	if err := a.page.Navigate("about:blank"); err != nil {
		logger.Warn().Err(err).Str("agent", a.Name).Msg("navigate about:blank failed, resetting page")
		resetPage()
		return nil
	}

	if err := a.page.SetCookies(nil); err != nil {
		logger.Warn().Err(err).Str("agent", a.Name).Msg("failed to clear cookies, resetting page")
		resetPage()
		return nil
	}

	_, _ = a.page.Eval(`() => {
		try { localStorage.clear(); } catch(e) {}
		try { sessionStorage.clear(); } catch(e) {}
	}`)

	a.account = nil
	a.estimatedCredits = 0

	logger.Debug().Str("agent", a.Name).Msg("session cleared")
	return nil
}

// injectPageCleanupScript starts the periodic DOM cleanup loop.
func (a *ScraperAgent) injectPageCleanupScript(page *rod.Page) {
	go a.runPageCleanupLoop(page)
}

// runPageCleanupLoop periodically removes blocking UI elements.
func (a *ScraperAgent) runPageCleanupLoop(page *rod.Page) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		if a.page != page || a.status == AgentStatusOffline {
			return
		}
		a.cleanupPageElements(page)
	}
}

// cleanupPageElements removes overlays, modals, focus guards, and
// pointer-events restrictions from the page.
func (a *ScraperAgent) cleanupPageElements(page *rod.Page) {
	defer func() {
		if r := recover(); r != nil {
			logger.Debug().Str("agent", a.Name).Interface("recover", r).Msg("recovered during page cleanup")
		}
	}()

	body, err := page.Element("body")
	if err == nil && body != nil {
		_, err = page.Eval(`() => {
			const body = document.body;
			if (body) {
				body.style.removeProperty('pointer-events');
				body.removeAttribute('data-scroll-locked');
				body.style.removeProperty('overflow');
				body.style.removeProperty('padding-right');
			}
		}`)
		if err != nil {
			logger.Debug().Str("agent", a.Name).Err(err).Msg("failed to clean body restrictions")
		}
	}

	removeSelectors := []string{
		`[class*="_overlay_"][data-state="open"]`,
		`[role="dialog"][data-state="open"]`,
		`[data-radix-focus-guard]`,
		`[class*="_toastViewport_"]`,
		`#pendo-base`,
	}
	for _, sel := range removeSelectors {
		if elems, err := page.Elements(sel); err == nil {
			for _, el := range elems {
				_ = el.Remove()
			}
		}
	}
}
