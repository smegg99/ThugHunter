// core/crawler/utils.go
package crawler

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
	"text/template"

	"github.com/brianvoe/gofakeit/v6"

	"smegg.me/thughunter/common/config"
	"smegg.me/thughunter/common/logger"
	"smegg.me/thughunter/core/models"
)

const (
	upper    = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	lower    = "abcdefghijklmnopqrstuvwxyz"
	digits   = "0123456789"
	specials = `~!@#$%^&*()_+{}":;[]'`
	allChars = upper + lower + digits + specials
)

func generateRandomNonsense() string {
	length := 12
	result := make([]byte, 0, length)

	result = append(result, randomChar(upper))
	result = append(result, randomChar(lower))
	result = append(result, randomChar(digits))
	result = append(result, randomChar(specials))

	for len(result) < length {
		result = append(result, randomChar(allChars))
	}

	shuffle(result)
	return string(result)
}

func randomChar(set string) byte {
	n := big.NewInt(int64(len(set)))
	i, _ := rand.Int(rand.Reader, n)
	return set[i.Int64()]
}

func shuffle(b []byte) {
	for i := len(b) - 1; i > 0; i-- {
		j := randInt(i + 1)
		b[i], b[j] = b[j], b[i]
	}
}

func randInt(max int) int {
	n, _ := rand.Int(rand.Reader, big.NewInt(int64(max)))
	return int(n.Int64())
}

// templateData holds the variables available for template substitution.
type templateData struct {
	ACCOUNT_ID      string
	RANDOM_NONSENSE string

	// Fake identity fields (generated via gofakeit)
	FIRST_NAME string
	LAST_NAME  string
	FULL_NAME  string
	USERNAME   string
	COMPANY    string
	CITY       string
	COUNTRY    string
	JOB_TITLE  string
	BUZZWORD   string
	DOMAIN     string
	DIGITS_4   string // 4 random digits, e.g. "0742"
	DIGITS_6   string // 6 random digits
}

// resolveTemplate renders a Go text/template string with the given data.
func resolveTemplate(tmplStr string, data templateData) (string, error) {
	t, err := template.New("").Parse(tmplStr)
	if err != nil {
		return "", fmt.Errorf("parse template %q: %w", tmplStr, err)
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute template %q: %w", tmplStr, err)
	}
	return buf.String(), nil
}

// newAccountFromTemplates creates a Account by resolving the config
// templates with a unique account ID.
func newAccountFromTemplates(accountID string) (*models.Account, error) {
	tpl := config.Get().Crawler.Agents.Templates
	nonsense := generateRandomNonsense()

	// Generate fake identity data
	faker := gofakeit.New(0) // 0 = random seed from crypto/rand
	firstName := faker.FirstName()
	lastName := faker.LastName()

	data := templateData{
		ACCOUNT_ID:      accountID,
		RANDOM_NONSENSE: nonsense,

		FIRST_NAME: firstName,
		LAST_NAME:  lastName,
		FULL_NAME:  firstName + " " + lastName,
		USERNAME:   strings.ToLower(firstName[:1] + lastName + fmt.Sprintf("%d", faker.IntRange(1, 99))),
		COMPANY:    faker.Company(),
		CITY:       faker.City(),
		COUNTRY:    faker.Country(),
		JOB_TITLE:  faker.JobTitle(),
		BUZZWORD:   faker.BuzzWord(),
		DOMAIN:     faker.DomainName(),
		DIGITS_4:   fmt.Sprintf("%04d", faker.IntRange(0, 9999)),
		DIGITS_6:   fmt.Sprintf("%06d", faker.IntRange(0, 999999)),
	}

	logger.Debug().
		Str("account_id", accountID).
		Str("random_nonsense", nonsense).
		Str("first_name", firstName).
		Str("last_name", lastName).
		Msg("resolving account templates")

	email, err := resolveTemplate(tpl.EmailTemplate, data)
	if err != nil {
		return nil, fmt.Errorf("resolve email template: %w", err)
	}

	password, err := resolveTemplate(tpl.PasswordTemplate, data)
	if err != nil {
		return nil, fmt.Errorf("resolve password template: %w", err)
	}

	firstName, err = resolveTemplate(tpl.FirstNameTemplate, data)
	if err != nil {
		return nil, fmt.Errorf("resolve first_name template: %w", err)
	}

	lastName, err = resolveTemplate(tpl.LastNameTemplate, data)
	if err != nil {
		return nil, fmt.Errorf("resolve last_name template: %w", err)
	}

	organization, err := resolveTemplate(tpl.OrganizationTemplate, data)
	if err != nil {
		return nil, fmt.Errorf("resolve organization template: %w", err)
	}

	logger.Debug().
		Str("email", email).
		Str("first_name", firstName).
		Str("last_name", lastName).
		Str("organization", organization).
		Msg("account templates resolved")

	return &models.Account{
		Email:        email,
		Password:     password,
		FirstName:    firstName,
		LastName:     lastName,
		Organization: organization,
	}, nil
}
