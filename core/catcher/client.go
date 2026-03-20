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

// dialIMAP connects to the IMAP server and selects the mailbox. It returns an authenticated client ready for use.
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

// fetchUnseenMails retrieves unseen emails from the mailbox, skipping any UIDs in the provided skip map. It returns a slice of fetchedMail.
func (ic *imapClient) fetchUnseenMails(skip map[imap.UID]struct{}) ([]fetchedMail, error) {
	if err := ic.c.Noop().Wait(); err != nil {
		return nil, fmt.Errorf("noop: %w", err)
	}

	toFetch, err := ic.searchUnseenUIDs(skip)
	if err != nil {
		return nil, err
	}
	if len(toFetch) == 0 {
		return nil, nil
	}

	return ic.fetchMails(toFetch)
}

// searchUnseenUIDs performs a search for unseen emails and returns their UIDs, excluding any UIDs present in the skip map.
func (ic *imapClient) searchUnseenUIDs(skip map[imap.UID]struct{}) ([]imap.UID, error) {
	criteria := &imap.SearchCriteria{
		NotFlag: []imap.Flag{imap.FlagSeen},
	}
	data, err := ic.c.UIDSearch(criteria, nil).Wait()
	if err != nil {
		return nil, fmt.Errorf("search: %w", err)
	}

	var toFetch []imap.UID
	for _, uid := range data.AllUIDs() {
		if _, ok := skip[uid]; !ok {
			toFetch = append(toFetch, uid)
		}
	}
	return toFetch, nil
}

// fetchMails retrieves the full email data for the given UIDs and returns a slice of fetchedMail.
func (ic *imapClient) fetchMails(uids []imap.UID) ([]fetchedMail, error) {
	bodySection := &imap.FetchItemBodySection{Peek: true}
	fetchOptions := &imap.FetchOptions{
		Envelope:    true,
		UID:         true,
		BodySection: []*imap.FetchItemBodySection{bodySection},
	}

	fetchCmd := ic.c.Fetch(imap.UIDSetNum(uids...), fetchOptions)

	var results []fetchedMail
	for {
		msg := fetchCmd.Next()
		if msg == nil {
			break
		}
		fm := parseFetchedMessage(msg)
		results = append(results, fm)
	}

	if err := fetchCmd.Close(); err != nil {
		return results, fmt.Errorf("fetch: %w", err)
	}
	return results, nil
}

// parseFetchedMessage extracts the UID, recipient addresses, and body text from the fetched message data. It returns a fetchedMail struct with this information.
func parseFetchedMessage(msg *imapclient.FetchMessageData) fetchedMail {
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
		fm.To = extractAddresses(envelope.To)
	}

	fm.Body = parseBody(bodyBytes)
	return fm
}

// extractAddresses converts a slice of imap.Address to a slice of email address strings in the format "mailbox@host".
func extractAddresses(addrs []imap.Address) []string {
	var result []string
	for _, addr := range addrs {
		result = append(result, addr.Mailbox+"@"+addr.Host)
	}
	return result
}

// parseBody attempts to read the email body as a multipart message. If successful, it extracts and returns the text/plain part. If parsing fails, it falls back to returning the raw body bytes as a string.
func parseBody(bodyBytes []byte) string {
	if len(bodyBytes) == 0 {
		return ""
	}
	mr, err := mail.CreateReader(bytes.NewReader(bodyBytes))
	if err != nil {
		return string(bodyBytes)
	}
	return extractTextBody(mr)
}

// extractTextBody traverses the mail reader parts to find and return the first text/plain body content. If no such part is found, it returns an empty string.
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

// snapshotExistingUIDs performs an initial search for unseen emails and returns their UIDs as a map. This is used to avoid processing old emails that were present before the catcher started.
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
