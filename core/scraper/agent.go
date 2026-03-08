// core/scraper/agent.go
package scraper

import (
	"context"
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"

	"smegg.me/thughunter/common/config"
	"smegg.me/thughunter/common/logger"
	"smegg.me/thughunter/core/human"
	"smegg.me/thughunter/core/models"
	"smegg.me/thughunter/core/unrevealed"

	petname "github.com/dustinkirkland/golang-petname"
)

type AgentStatus string

const (
	AgentStatusOffline AgentStatus = "offline" // When the agent is not logged in or has not been initialized
	AgentStatusIdle    AgentStatus = "idle"    // When the agent is logged in but not currently performing any tasks
	AgentStatusBusy    AgentStatus = "busy"    // When the agent is actively performing a task, such as logging in, searching, or scraping
	AgentStatusError   AgentStatus = "error"   // When the agent has encountered an error that prevents it from functioning properly
)

// PetnameWords controls how many words the petname generator uses (e.g. 2 = "bold-frog").
var PetnameWords = 2

// PetnameSeparator is the separator between petname words.
var PetnameSeparator = "-"

type ScraperAgent struct {
	Name     string
	status   AgentStatus
	browser  *unrevealed.Browser
	account  *models.Account
}

// generatePetname returns a random petname with the configured word count and separator.
func generatePetname() string {
	return petname.Generate(PetnameWords, PetnameSeparator)
}

func newScraperAgent(name string) *ScraperAgent {
	logger.Debug().Str("agent", name).Msg("creating agent")
	return &ScraperAgent{
		Name:   name,
		status: AgentStatusOffline,
	}
}

func (a *ScraperAgent) ensureBrowser() error {
	if a.browser != nil {
		return nil
	}

	logger.Debug().Str("agent", a.Name).Msg("launching browser for agent")

	cfg := config.Get()
	browser, err := unrevealed.New(context.Background(), unrevealed.Config{
		ChromePath: cfg.Scraper.BrowserBinaryPath,
		Headless:   false,
	})
	if err != nil {
		a.SetStatus(AgentStatusError)
		return fmt.Errorf("agent %s: launch browser: %w", a.Name, err)
	}

	a.browser = browser
	logger.Debug().Str("agent", a.Name).Msg("browser launched for agent")
	return nil
}

func (a *ScraperAgent) newTab(url string) (*rod.Page, *human.Cursor, error) {
	if err := a.ensureBrowser(); err != nil {
		return nil, nil, err
	}

	logger.Debug().Str("agent", a.Name).Str("url", url).Msg("opening new tab")

	var page *rod.Page
	pages, err := a.browser.Pages()
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
		page, err = a.browser.Page(proto.TargetCreateTarget{URL: "about:blank"})
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

	logger.Debug().Str("agent", a.Name).Str("url", url).Msg("page loaded")

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

func (a *ScraperAgent) Status() AgentStatus {
	return a.status
}

func (a *ScraperAgent) IsIdle() bool {
	return a.status == AgentStatusIdle
}

func (a *ScraperAgent) IsBusy() bool {
	return a.status == AgentStatusBusy
}

func (a *ScraperAgent) IsOffline() bool {
	return a.status == AgentStatusOffline
}

func (a *ScraperAgent) IsError() bool {
	return a.status == AgentStatusError
}

func (a *ScraperAgent) SetStatus(status AgentStatus) {
	switch status {
	case AgentStatusIdle, AgentStatusBusy, AgentStatusOffline:
		logger.Debug().Str("agent", a.Name).Str("status", string(status)).Msg("setting agent status")
	case AgentStatusError:
		logger.Error().Str("agent", a.Name).Str("status", string(status)).Msg("agent entered error status")
		a.SetStatus(status)
		return
	default:
		logger.Warn().Str("agent", a.Name).Str("status", string(status)).Msg("attempted to set invalid agent status")
		return
	}
	a.SetStatus(status)
}

func (a *ScraperAgent) IsLoggedIn() bool {
	return a.account != nil && a.account.IsValid() && (a.IsBusy() || a.IsIdle())
}

func (a *ScraperAgent) Account() *models.Account {
	return a.account
}

func (a *ScraperAgent) Close() error {
	logger.Debug().Str("agent", a.Name).Msg("closing agent")

	a.SetStatus(AgentStatusOffline)

	if a.browser != nil {
		if err := a.browser.Close(); err != nil {
			logger.Error().Err(err).Str("agent", a.Name).Msg("error closing agent browser")
			return err
		}
		a.browser = nil
	}

	logger.Info().Str("agent", a.Name).Msg("scraper agent closed")
	return nil
}
