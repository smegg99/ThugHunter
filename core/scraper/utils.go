// core/scraper/utils.go
package scraper

import (
	"fmt"
	"math/rand/v2"
	"strings"

	"github.com/brianvoe/gofakeit/v6"
	petname "github.com/dustinkirkland/golang-petname"

	"smegg.me/thughunter/common/config"
	"smegg.me/thughunter/common/logger"
	"smegg.me/thughunter/common/templating"
	"smegg.me/thughunter/core/models"
)

// PetnameWords controls how many words the petname generator uses (e.g. 2 = "bold-frog").
var PetnameWords = 2

// PetnameSeparator is the separator between petname words.
var PetnameSeparator = "-"

const (
	upper    = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	lower    = "abcdefghijklmnopqrstuvwxyz"
	digits   = "0123456789"
	specials = `~!@#$%^&*()_+{}":;[]'`
	allChars = upper + lower + digits + specials
)

// generatePetname returns a random petname with the configured word count and separator.
func generatePetname() string {
	return petname.Generate(PetnameWords, PetnameSeparator)
}

// generateRandomNonsense returns a 12-char string with at least one
// uppercase, lowercase, digit, and special character.
func generateRandomNonsense() string {
	const length = 8
	result := make([]byte, 0, length)

	result = append(result, upper[rand.IntN(len(upper))])
	result = append(result, lower[rand.IntN(len(lower))])
	result = append(result, digits[rand.IntN(len(digits))])
	result = append(result, specials[rand.IntN(len(specials))])

	for len(result) < length {
		result = append(result, allChars[rand.IntN(len(allChars))])
	}

	rand.Shuffle(len(result), func(i, j int) { result[i], result[j] = result[j], result[i] })
	return string(result)
}

// templateData holds the variables available for template substitution.
type templateData struct {
	ACCOUNT_ID      string
	RANDOM_NONSENSE string

	FIRST_NAME    string
	LAST_NAME     string
	LC_FIRST_NAME string
	LC_LAST_NAME  string
	FULL_NAME     string
	LC_FULL_NAME  string
	USERNAME      string
	COMPANY       string
	CITY          string
	COUNTRY       string
	JOB_TITLE     string
	BUZZWORD      string
	DOMAIN        string
	DIGITS_1      string
	DIGITS_2      string
	DIGITS_3      string
	DIGITS_4      string
	DIGITS_5      string
	DIGITS_6      string
}

// newAccountFromTemplates creates an Account by resolving config templates.
func newAccountFromTemplates(accountID string) (*models.Account, error) {
	tpl := config.Get().Scraper.Agents.Templates
	nonsense := generateRandomNonsense()

	faker := gofakeit.New(0) // 0 = random seed from crypto/rand
	firstName := faker.FirstName()
	lastName := faker.LastName()
	fullName := fmt.Sprintf("%s %s", firstName, lastName)

	data := templateData{
		ACCOUNT_ID:      accountID,
		RANDOM_NONSENSE: nonsense,

		FIRST_NAME:    firstName,
		LAST_NAME:     lastName,
		LC_FIRST_NAME: strings.ToLower(firstName),
		LC_LAST_NAME:  strings.ToLower(lastName),
		FULL_NAME:     fullName,
		LC_FULL_NAME:  strings.ToLower(fullName),
		USERNAME:      strings.ToLower(firstName[:1] + lastName + fmt.Sprintf("%d", faker.IntRange(1, 99))),
		COMPANY:       faker.Company(),
		CITY:          faker.City(),
		COUNTRY:       faker.Country(),
		JOB_TITLE:     faker.JobTitle(),
		BUZZWORD:      faker.BuzzWord(),
		DOMAIN:        faker.DomainName(),
		DIGITS_1:      fmt.Sprintf("%d", faker.IntRange(0, 9)),
		DIGITS_2:      fmt.Sprintf("%02d", faker.IntRange(0, 99)),
		DIGITS_3:      fmt.Sprintf("%03d", faker.IntRange(0, 999)),
		DIGITS_4:      fmt.Sprintf("%04d", faker.IntRange(0, 9999)),
		DIGITS_5:      fmt.Sprintf("%05d", faker.IntRange(0, 99999)),
		DIGITS_6:      fmt.Sprintf("%06d", faker.IntRange(0, 999999)),
	}

	logger.Debug().
		Str("account_id", accountID).
		Str("random_nonsense", nonsense).
		Str("first_name", firstName).
		Str("last_name", lastName).
		Msg("resolving account templates")

	type field struct {
		name string
		tpl  string
		dst  *string
	}

	var email, password, organization string
	fields := []field{
		{"email", tpl.EmailTemplate, &email},
		{"password", tpl.PasswordTemplate, &password},
		{"first_name", tpl.FirstNameTemplate, &firstName},
		{"last_name", tpl.LastNameTemplate, &lastName},
		{"organization", tpl.OrganizationTemplate, &organization},
	}
	for _, f := range fields {
		val, err := templating.Resolve(f.tpl, data)
		if err != nil {
			return nil, fmt.Errorf("resolve %s template: %w", f.name, err)
		}
		*f.dst = val
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
