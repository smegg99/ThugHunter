// core/crawler/agent.go
package crawler

import (
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"

	"smegg.me/thughunter/common/logger"
	"smegg.me/thughunter/core/human"
	"smegg.me/thughunter/core/models"
	"smegg.me/thughunter/core/unrevealed"
)

type CrawlerAgent struct {
	ID       int
	ctx      *rod.Browser
	loggedIn bool
	account  *models.Account
}

func newCrawlerAgent(id int, browser *rod.Browser) (*CrawlerAgent, error) {
	logger.Debug().Int("agent_id", id).Msg("creating agent")
	return &CrawlerAgent{ID: id, ctx: browser}, nil
}

func (a *CrawlerAgent) newTab(url string) (*rod.Page, *human.Cursor, error) {
	logger.Debug().Int("agent_id", a.ID).Str("url", url).Msg("opening new tab")

	var page *rod.Page
	pages, err := a.ctx.Pages()
	if err == nil {
		for _, p := range pages {
			info, _ := p.Info()
			if info != nil && (info.URL == "" || info.URL == "about:blank" ||
				info.URL == "chrome://newtab/" || info.URL == "chrome://new-tab-page/") {
				page = p
				break
			}
		}
	}
	if page == nil {
		page, err = a.ctx.Page(proto.TargetCreateTarget{URL: "about:blank"})
		if err != nil {
			return nil, nil, fmt.Errorf("new page: %w", err)
		}
	}

	if err := unrevealed.Stealth(page); err != nil {
		page.MustClose()
		return nil, nil, fmt.Errorf("apply stealth: %w", err)
	}

	if err := page.Navigate(url); err != nil {
		page.MustClose()
		return nil, nil, fmt.Errorf("navigate to %s: %w", url, err)
	}

	if err := page.WaitLoad(); err != nil {
		page.MustClose()
		return nil, nil, fmt.Errorf("wait load %s: %w", url, err)
	}

	time.Sleep(time.Duration(800+rand.IntN(1200)) * time.Millisecond)

	logger.Debug().Int("agent_id", a.ID).Str("url", url).Msg("page loaded")

	cursor := human.New(page, func(c *human.Config) {
		c.Direct = false
		c.Hesitation = 0.01
		c.MicroPause = 0.01
		c.Steadiness = 0.9
		c.ClickHold = [2]int{25, 60}
		c.ClickDwell = [2]int{50, 120}
		c.TypeDelay = [2]int{15, 55}
		c.ThinkPause = 0.01
	})
	return page, cursor, nil
}

func (a *CrawlerAgent) IsLoggedIn() bool {
	return a.loggedIn
}

func (a *CrawlerAgent) Account() *models.Account {
	return a.account
}

func (a *CrawlerAgent) Close() error {
	logger.Debug().Int("agent_id", a.ID).Msg("closing agent")
	logger.Info().Int("agent_id", a.ID).Msg("crawler agent closed")
	return nil
}
