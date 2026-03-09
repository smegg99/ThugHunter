// core/scraper/queries.go
package scraper

import "smegg.me/thughunter/common/templating"

type QueryString string

const (
	BaseVNCQueryString          QueryString = "host.services.vnc.security_types.value:\"1\""
	BaseVNCByCountryQueryString QueryString = "host.services.vnc.security_types.value:\"1\" and host.location.country:\"{{.COUNTRY_NAME}}\""
)

type Country string

// All the countries that are available in the Censys Search.
const (
	CountryUnitedStates         Country = "United States"
	CountryGermany              Country = "Germany"
	CountrySouthKorea           Country = "South Korea"
	CountryFrance               Country = "France"
	CountryUnitedKingdom        Country = "United Kingdom"
	CountryChina                Country = "China"
	CountryItaly                Country = "Italy"
	CountryJapan                Country = "Japan"
	CountryBrazil               Country = "Brazil"
	CountryCanada               Country = "Canada"
	CountryHongKong             Country = "Hong Kong"
	CountryIndia                Country = "India"
	CountryNetherlands          Country = "Netherlands"
	CountryAustralia            Country = "Australia"
	CountrySingapore            Country = "Singapore"
	CountryRussia               Country = "Russia"
	CountrySpain                Country = "Spain"
	CountryDenmark              Country = "Denmark"
	CountryPoland               Country = "Poland"
	CountryTaiwan               Country = "Taiwan"
	CountrySweden               Country = "Sweden"
	CountryArgentina            Country = "Argentina"
	CountryIreland              Country = "Ireland"
	CountryThailand             Country = "Thailand"
	CountrySouthAfrica          Country = "South Africa"
	CountryMorocco              Country = "Morocco"
	CountryVietnam              Country = "Vietnam"
	CountryBelgium              Country = "Belgium"
	CountryPakistan             Country = "Pakistan"
	CountryIndonesia            Country = "Indonesia"
	CountryTurkey               Country = "Turkey"
	CountryMexico               Country = "Mexico"
	CountrySwitzerland          Country = "Switzerland"
	CountryRomania              Country = "Romania"
	CountryNewZealand           Country = "New Zealand"
	CountryFinland              Country = "Finland"
	CountryColombia             Country = "Colombia"
	CountryAustria              Country = "Austria"
	CountryUkraine              Country = "Ukraine"
	CountryMalaysia             Country = "Malaysia"
	CountryHungary              Country = "Hungary"
	CountryEgypt                Country = "Egypt"
	CountryCzechRepublic        Country = "Czech Republic"
	CountryBulgaria             Country = "Bulgaria"
	CountryGreece               Country = "Greece"
	CountryPortugal             Country = "Portugal"
	CountryTunisia              Country = "Tunisia"
	CountryChile                Country = "Chile"
	CountrySaudiArabia          Country = "Saudi Arabia"
	CountryAlgeria              Country = "Algeria"
	CountryUnitedArabEmirates   Country = "United Arab Emirates"
	CountryNorway               Country = "Norway"
	CountryIsrael               Country = "Israel"
	CountryPhilippines          Country = "Philippines"
	CountryKazakhstan           Country = "Kazakhstan"
	CountryVenezuela            Country = "Venezuela"
	CountrySerbia               Country = "Serbia"
	CountryBangladesh           Country = "Bangladesh"
	CountryEcuador              Country = "Ecuador"
	CountryCroatia              Country = "Croatia"
	CountryLatvia               Country = "Latvia"
	CountryIran                 Country = "Iran"
	CountryPanama               Country = "Panama"
	CountryPeru                 Country = "Peru"
	CountryMoldova              Country = "Moldova"
	CountryLithuania            Country = "Lithuania"
	CountrySenegal              Country = "Senegal"
	CountryPuertoRico           Country = "Puerto Rico"
	CountrySlovakia             Country = "Slovakia"
	CountryCostaRica            Country = "Costa Rica"
	CountryBosniaAndHerzegovina Country = "Bosnia and Herzegovina"
	CountryBelarus              Country = "Belarus"
	CountryOman                 Country = "Oman"
	CountryKuwait               Country = "Kuwait"
	CountryUruguay              Country = "Uruguay"
	CountrySlovenia             Country = "Slovenia"
	CountryEstonia              Country = "Estonia"
	CountryNigeria              Country = "Nigeria"
	CountryKenya                Country = "Kenya"
	CountryGeorgia              Country = "Georgia"
	CountryBurkinaFaso          Country = "Burkina Faso"
	CountryReunion              Country = "Reunion"
	CountryTrinidadAndTobago    Country = "Trinidad and Tobago"
	CountryJamaica              Country = "Jamaica"
	CountryIceland              Country = "Iceland"
	CountryDominicanRepublic    Country = "Dominican Republic"
	CountryCambodia             Country = "Cambodia"
	CountryPalestinianTerritory Country = "Palestinian Territory"
	CountryBolivia              Country = "Bolivia"
	CountryLuxembourg           Country = "Luxembourg"
	CountryIvoryCoast           Country = "Ivory Coast"
	CountryMauritius            Country = "Mauritius"
	CountryMartinique           Country = "Martinique"
	CountryCyprus               Country = "Cyprus"
	CountryBahrain              Country = "Bahrain"
	CountryParaguay             Country = "Paraguay"
	CountryYemen                Country = "Yemen"
	CountrySriLanka             Country = "Sri Lanka"
	CountryIraq                 Country = "Iraq"
	CountryGuadeloupe           Country = "Guadeloupe"
)

func (c Country) String() string {
	return string(c)
}

type OperatingSystem string

const (
	OperatingSystemWindows OperatingSystem = "Windows"
	OperatingSystemLinux   OperatingSystem = "Linux"
	OperatingSystemUnix   OperatingSystem = "Unix"
)

func (os OperatingSystem) String() string {
	return string(os)
}

// Resolve renders the QueryString template with the given data.
func ResolveVNCSearchQueryForCountry(country Country) (string, error) {
	type QueryTemplate struct {
		COUNTRY_NAME string
	}
	return templating.Resolve(string(BaseVNCByCountryQueryString), QueryTemplate{
		COUNTRY_NAME: country.String(),
	})
}
