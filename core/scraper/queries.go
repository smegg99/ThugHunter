// core/scraper/queries.go
package scraper

import "smegg.me/thughunter/common/templating"

type QueryString string

const (
	BaseVNCQueryString                QueryString = "host.services.vnc.security_types.value:\"1\""
	BaseVNCByCountryQueryString       QueryString = "host.services.vnc.security_types.value:\"1\" and host.location.country:\"{{.COUNTRY_NAME}}\""
	BaseVNCNativeQueryString          QueryString = "host.operating_system.product: \"Linux\" and not host.services.software.type: \"VIRTUAL_MACHINE\" and not host.services.software.type: \"CONTAINER\" and not host.autonomous_system.name: \"AWS\" and not host.autonomous_system.name: \"GCP\" and not host.autonomous_system.name: \"AZURE\" and host.services.vnc.security_types.name: \"None\""
	BaseVNCNativeByCountryQueryString QueryString = BaseVNCNativeQueryString + " and host.location.country:\"{{.COUNTRY_NAME}}\""
	BaseCameraNoAuthQueryString       QueryString = "host.services: (hardware.type: \"CAMERA\" and not endpoints.http.headers.key: \"Authorization\")"
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

// AllCountries lists every Country constant for query generation.
var AllCountries = []Country{
	CountryUnitedStates, CountryGermany, CountrySouthKorea, CountryFrance,
	CountryUnitedKingdom, CountryChina, CountryItaly, CountryJapan,
	CountryBrazil, CountryCanada, CountryHongKong, CountryIndia,
	CountryNetherlands, CountryAustralia, CountrySingapore, CountryRussia,
	CountrySpain, CountryDenmark, CountryPoland, CountryTaiwan,
	CountrySweden, CountryArgentina, CountryIreland, CountryThailand,
	CountrySouthAfrica, CountryMorocco, CountryVietnam, CountryBelgium,
	CountryPakistan, CountryIndonesia, CountryTurkey, CountryMexico,
	CountrySwitzerland, CountryRomania, CountryNewZealand, CountryFinland,
	CountryColombia, CountryAustria, CountryUkraine, CountryMalaysia,
	CountryHungary, CountryEgypt, CountryCzechRepublic, CountryBulgaria,
	CountryGreece, CountryPortugal, CountryTunisia, CountryChile,
	CountrySaudiArabia, CountryAlgeria, CountryUnitedArabEmirates, CountryNorway,
	CountryIsrael, CountryPhilippines, CountryKazakhstan, CountryVenezuela,
	CountrySerbia, CountryBangladesh, CountryEcuador, CountryCroatia,
	CountryLatvia, CountryIran, CountryPanama, CountryPeru,
	CountryMoldova, CountryLithuania, CountrySenegal, CountryPuertoRico,
	CountrySlovakia, CountryCostaRica, CountryBosniaAndHerzegovina, CountryBelarus,
	CountryOman, CountryKuwait, CountryUruguay, CountrySlovenia,
	CountryEstonia, CountryNigeria, CountryKenya, CountryGeorgia,
	CountryBurkinaFaso, CountryReunion, CountryTrinidadAndTobago, CountryJamaica,
	CountryIceland, CountryDominicanRepublic, CountryCambodia, CountryPalestinianTerritory,
	CountryBolivia, CountryLuxembourg, CountryIvoryCoast, CountryMauritius,
	CountryMartinique, CountryCyprus, CountryBahrain, CountryParaguay,
	CountryYemen, CountrySriLanka, CountryIraq, CountryGuadeloupe,
}

// ResolveQueryForCountry renders a per-country query template.
func ResolveQueryForCountry(query QueryString, country Country) (string, error) {
	type QueryTemplate struct {
		COUNTRY_NAME string
	}
	return templating.Resolve(string(query), QueryTemplate{
		COUNTRY_NAME: country.String(),
	})
}

// ResolveVNCSearchQueryForCountry renders the per-country VNC query template.
func ResolveVNCSearchQueryForCountry(country Country) (string, error) {
	return ResolveQueryForCountry(BaseVNCByCountryQueryString, country)
}

// Continent groups countries by geographic region.
type Continent string

const (
	ContinentEurope       Continent = "Europe"
	ContinentAsia         Continent = "Asia"
	ContinentNorthAmerica Continent = "North America"
	ContinentSouthAmerica Continent = "South America"
	ContinentAfrica       Continent = "Africa"
	ContinentOceania      Continent = "Oceania"
	ContinentMiddleEast   Continent = "Middle East"
)

// AllContinents lists every defined continent.
var AllContinents = []Continent{
	ContinentEurope, ContinentAsia, ContinentNorthAmerica,
	ContinentSouthAmerica, ContinentAfrica, ContinentOceania,
	ContinentMiddleEast,
}

// CountriesByContinent maps each continent to its countries.
var CountriesByContinent = map[Continent][]Country{
	ContinentEurope: {
		CountryGermany, CountryFrance, CountryUnitedKingdom, CountryItaly,
		CountryNetherlands, CountrySpain, CountryDenmark, CountryPoland,
		CountrySweden, CountryIreland, CountryBelgium, CountrySwitzerland,
		CountryRomania, CountryFinland, CountryAustria, CountryUkraine,
		CountryHungary, CountryCzechRepublic, CountryBulgaria, CountryGreece,
		CountryPortugal, CountryNorway, CountrySerbia, CountryCroatia,
		CountryLatvia, CountryMoldova, CountryLithuania, CountrySlovakia,
		CountryBosniaAndHerzegovina, CountryBelarus, CountrySlovenia,
		CountryEstonia, CountryIceland, CountryLuxembourg, CountryRussia,
		CountryCyprus, CountryGeorgia,
	},
	ContinentAsia: {
		CountrySouthKorea, CountryChina, CountryJapan, CountryHongKong,
		CountryIndia, CountrySingapore, CountryTaiwan, CountryThailand,
		CountryVietnam, CountryPakistan, CountryIndonesia, CountryMalaysia,
		CountryPhilippines, CountryKazakhstan, CountryBangladesh, CountryCambodia,
		CountrySriLanka,
	},
	ContinentNorthAmerica: {
		CountryUnitedStates, CountryCanada, CountryMexico, CountryPanama,
		CountryCostaRica, CountryPuertoRico, CountryJamaica,
		CountryDominicanRepublic, CountryTrinidadAndTobago,
		CountryGuadeloupe, CountryMartinique,
	},
	ContinentSouthAmerica: {
		CountryBrazil, CountryArgentina, CountryColombia, CountryChile,
		CountryVenezuela, CountryEcuador, CountryPeru, CountryUruguay,
		CountryBolivia, CountryParaguay,
	},
	ContinentAfrica: {
		CountrySouthAfrica, CountryMorocco, CountryEgypt, CountryTunisia,
		CountryAlgeria, CountrySenegal, CountryNigeria, CountryKenya,
		CountryBurkinaFaso, CountryIvoryCoast, CountryMauritius, CountryReunion,
	},
	ContinentOceania: {
		CountryAustralia, CountryNewZealand,
	},
	ContinentMiddleEast: {
		CountryTurkey, CountrySaudiArabia, CountryUnitedArabEmirates,
		CountryIsrael, CountryIran, CountryOman, CountryKuwait,
		CountryBahrain, CountryPalestinianTerritory, CountryYemen, CountryIraq,
	},
}
