// core/catcher/client.go
package catcher

import (
	"bytes"
	"fmt"
	"io"

	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
	"github.com/emersion/go-message/mail"

	"smegg.me/thughunter/common/logger"
)

type imapClient struct {
	c    *imapclient.Client
	mbox string
}

func dialIMAP(host string, port int, username, password, mbox string, useTLS bool) (*imapClient, error) {
	addr := fmt.Sprintf("%s:%d", host, port)

	var c *imapclient.Client
	var err error
	if useTLS {
		c, err = imapclient.DialTLS(addr, nil)
	} else {
		c, err = imapclient.DialInsecure(addr, nil)
	}
	if err != nil {
		return nil, fmt.Errorf("dial %s: %w", addr, err)
	}

	if err := c.Login(username, password).Wait(); err != nil {
		c.Close()
		return nil, fmt.Errorf("login: %w", err)
	}

	if _, err := c.Select(mbox, nil).Wait(); err != nil {
		c.Close()
		return nil, fmt.Errorf("select %s: %w", mbox, err)
	}

	logger.Info().Str("host", addr).Str("mbox", mbox).Msg("IMAP connected")
	return &imapClient{c: c, mbox: mbox}, nil
}

type fetchedMail struct {
	UID  imap.UID
	To   []string
	Body string
}

func (ic *imapClient) fetchUnseenMails(skip map[imap.UID]struct{}) ([]fetchedMail, error) {
	if err := ic.c.Noop().Wait(); err != nil {
		return nil, fmt.Errorf("noop: %w", err)
	}

	criteria := &imap.SearchCriteria{
		NotFlag: []imap.Flag{imap.FlagSeen},
	}
	data, err := ic.c.UIDSearch(criteria, nil).Wait()
	if err != nil {
		return nil, fmt.Errorf("search: %w", err)
	}

	uids := data.AllUIDs()
	if len(uids) == 0 {
		return nil, nil
	}

	var toFetch []imap.UID
	for _, uid := range uids {
		if _, ok := skip[uid]; !ok {
			toFetch = append(toFetch, uid)
		}
	}
	if len(toFetch) == 0 {
		return nil, nil
	}

	bodySection := &imap.FetchItemBodySection{Peek: true}
	fetchOptions := &imap.FetchOptions{
		Envelope:    true,
		UID:         true,
		BodySection: []*imap.FetchItemBodySection{bodySection},
	}

	fetchCmd := ic.c.Fetch(imap.UIDSetNum(toFetch...), fetchOptions)

	var results []fetchedMail
	for {
		msg := fetchCmd.Next()
		if msg == nil {
			break
		}

		var fm fetchedMail
		var envelope *imap.Envelope
		var bodyBytes []byte

		for {
			item := msg.Next()
			if item == nil {
				break
			}
			switch d := item.(type) {
			case imapclient.FetchItemDataUID:
				fm.UID = d.UID
			case imapclient.FetchItemDataEnvelope:
				envelope = d.Envelope
			case imapclient.FetchItemDataBodySection:
				bodyBytes, _ = io.ReadAll(d.Literal)
			}
		}

		if envelope != nil {
			for _, addr := range envelope.To {
				fm.To = append(fm.To, addr.Mailbox+"@"+addr.Host)
			}
		}

		if len(bodyBytes) > 0 {
			mr, err := mail.CreateReader(bytes.NewReader(bodyBytes))
			if err == nil {
				fm.Body = extractTextBody(mr)
			} else {
				// Use the raw message bytes as a fallback if parsing fails.
				fm.Body = string(bodyBytes)
			}
		}

		results = append(results, fm)
	}

	if err := fetchCmd.Close(); err != nil {
		return results, fmt.Errorf("fetch: %w", err)
	}

	return results, nil
}

func extractTextBody(mr *mail.Reader) string {
	for {
		p, err := mr.NextPart()
		if err != nil {
			break
		}
		if _, ok := p.Header.(*mail.InlineHeader); ok {
			body, err := io.ReadAll(p.Body)
			if err == nil && len(body) > 0 {
				return string(body)
			}
		}
	}
	return ""
}

func (ic *imapClient) snapshotExistingUIDs() ([]imap.UID, error) {
	criteria := &imap.SearchCriteria{
		NotFlag: []imap.Flag{imap.FlagSeen},
	}
	data, err := ic.c.UIDSearch(criteria, nil).Wait()
	if err != nil {
		return nil, fmt.Errorf("snapshot search: %w", err)
	}
	return data.AllUIDs(), nil
}

func (ic *imapClient) close() error {
	return ic.c.Close()
}
