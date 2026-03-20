// core/catcher/validate.go
package catcher

import (
	"fmt"
	"io"
	"strings"

	"github.com/emersion/go-imap/v2/imapclient"
)

// ValidationError is a structured error returned by ValidateCredentials.
// Code is a machine-readable key for frontend localisation.
type ValidationError struct {
	Code string `json:"code"`
}

func (e *ValidationError) Error() string {
	return e.Code
}

// ValidateCredentials connects to the IMAP server, authenticates, and selects
// the mailbox. It returns nil if everything succeeds, or a *ValidationError
// indicating which step failed.
func ValidateCredentials(host string, port int, username, password, mbox string, useTLS bool) error {
	addr := fmt.Sprintf("%s:%d", host, port)

	var c *imapclient.Client
	var err error
	if useTLS {
		c, err = imapclient.DialTLS(addr, nil)
	} else {
		c, err = imapclient.DialInsecure(addr, nil)
	}
	if err != nil {
		if !useTLS {
			return &ValidationError{Code: "connection_failed_try_tls"}
		}
		return &ValidationError{Code: "connection_failed"}
	}
	defer c.Close()

	if err := c.Login(username, password).Wait(); err != nil {
		// EOF or connection reset during login on a non-TLS connection
		// almost always means the server requires TLS.
		if !useTLS && isConnectionLost(err) {
			return &ValidationError{Code: "login_failed_try_tls"}
		}
		return &ValidationError{Code: "login_failed"}
	}

	if _, err := c.Select(mbox, nil).Wait(); err != nil {
		return &ValidationError{Code: "mailbox_not_found"}
	}

	return nil
}

func isConnectionLost(err error) bool {
	if err == io.EOF || err == io.ErrUnexpectedEOF {
		return true
	}
	s := err.Error()
	return strings.Contains(s, "EOF") || strings.Contains(s, "connection reset") || strings.Contains(s, "broken pipe")
}
