// core/catcher/catcher.go
package catcher

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/emersion/go-imap/v2"

	"smegg.me/thughunter/common/config"
	"smegg.me/thughunter/common/logger"
)

const defaultPollInterval = 5 * time.Second

type Catcher struct {
	mu        sync.Mutex
	waiters   map[string]chan string
	client    *imapClient
	processed map[imap.UID]struct{}
	stop      chan struct{}
	done      chan struct{}
}

func New() (*Catcher, error) {
	cfg := config.Get().Imap
	cl, err := dialIMAP(cfg.Host, int(cfg.Port), cfg.CatchAllUsername, cfg.CatchAllPassword, cfg.Mbox, cfg.UseTls)
	if err != nil {
		return nil, fmt.Errorf("catcher: %w", err)
	}

	processed := make(map[imap.UID]struct{})

	existing, err := cl.snapshotExistingUIDs()
	if err != nil {
		cl.close()
		return nil, fmt.Errorf("catcher: snapshot existing: %w", err)
	}
	for _, uid := range existing {
		processed[uid] = struct{}{}
	}
	logger.Info().Int("skipped", len(existing)).Msg("catcher: ignored pre-existing unseen emails")

	c := &Catcher{
		waiters:   make(map[string]chan string),
		client:    cl,
		processed: processed,
		stop:      make(chan struct{}),
		done:      make(chan struct{}),
	}

	go c.pollLoop()

	logger.Info().Msg("mail catcher started")
	return c, nil
}

func (c *Catcher) WaitForCode(email string, timeout time.Duration) (string, error) {
	key := strings.ToLower(email)
	ch := make(chan string, 1)

	c.mu.Lock()
	c.waiters[key] = ch
	c.mu.Unlock()

	defer func() {
		c.mu.Lock()
		delete(c.waiters, key)
		c.mu.Unlock()
	}()

	logger.Info().Str("email", email).Dur("timeout", timeout).Msg("waiting for verification code")

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	select {
	case code := <-ch:
		logger.Info().Str("email", email).Str("code", code).Msg("verification code received")
		return code, nil
	case <-ctx.Done():
		return "", fmt.Errorf("catcher: timeout waiting for verification code for %s", email)
	case <-c.stop:
		return "", fmt.Errorf("catcher: stopped while waiting for %s", email)
	}
}

func (c *Catcher) Close() error {
	close(c.stop)
	<-c.done
	return c.client.close()
}

func (c *Catcher) pollLoop() {
	defer close(c.done)

	ticker := time.NewTicker(defaultPollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-c.stop:
			return
		case <-ticker.C:
			c.poll()
		}
	}
}

func (c *Catcher) poll() {
	c.mu.Lock()
	if len(c.waiters) == 0 {
		c.mu.Unlock()
		return
	}
	c.mu.Unlock()

	mails, err := c.client.fetchUnseenMails(c.processed)
	if err != nil {
		logger.Error().Err(err).Msg("catcher: poll failed")
		return
	}

	for _, m := range mails {
		c.dispatch(m)
	}
}

func (c *Catcher) dispatch(m fetchedMail) {
	code := extractVerificationCode(m.Body)
	if code == "" {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	for _, to := range m.To {
		key := strings.ToLower(to)
		if ch, ok := c.waiters[key]; ok {
			logger.Info().
				Str("to", to).
				Str("code", code).
				Uint32("uid", uint32(m.UID)).
				Msg("dispatching verification code")
			ch <- code
			c.processed[m.UID] = struct{}{}
			return
		}
	}

	logger.Debug().
		Strs("to", m.To).
		Uint32("uid", uint32(m.UID)).
		Msg("no waiter for email, skipping")
}
