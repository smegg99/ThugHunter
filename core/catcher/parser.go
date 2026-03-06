// core/catcher/parser.go
package catcher

import "regexp"

var codeRegex = regexp.MustCompile(`\b(\d{6})\b`)

func extractVerificationCode(body string) string {
	match := codeRegex.FindStringSubmatch(body)
	if len(match) >= 2 {
		return match[1]
	}
	return ""
}
