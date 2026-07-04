package geoip

import "strings"

// CountryName maps ISO 3166-1 alpha-2 codes to English display names.
func CountryName(code string) string {
	code = strings.ToUpper(strings.TrimSpace(code))
	if code == "" {
		return ""
	}
	if name, ok := countryNames[code]; ok {
		return name
	}
	return code
}

// FormatLocation builds a human label for dashboards (region override, else country).
func FormatLocation(region, countryCode string) string {
	region = strings.TrimSpace(region)
	if region != "" {
		return region
	}
	if name := CountryName(countryCode); name != "" {
		return name
	}
	return ""
}

// countryNames — common fleet regions; unknown codes fall back to the code itself.
var countryNames = map[string]string{
	"AE": "United Arab Emirates",
	"CN": "China",
	"DE": "Germany",
	"FR": "France",
	"GB": "United Kingdom",
	"HK": "Hong Kong",
	"IR": "Iran",
	"JP": "Japan",
	"NL": "Netherlands",
	"RU": "Russia",
	"SG": "Singapore",
	"TR": "Turkey",
	"US": "United States",
}
