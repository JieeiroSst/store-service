package domain

import (
	"crypto/rand"
	"net/url"
)

const shortCodeAlphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789_-"

const templateSlugAlphabet = "abcdefghijklmnopqrstuvwxyz0123456789"

func randomString(alphabet string, length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	out := make([]byte, length)
	n := len(alphabet)
	for i, v := range b {
		out[i] = alphabet[int(v)%n]
	}
	return string(out), nil
}

func GenerateShortCode(length int) (string, error) {
	if length <= 0 {
		length = 8
	}
	return randomString(shortCodeAlphabet, length)
}

func GenerateTemplateSlug() (string, error) {
	return randomString(templateSlugAlphabet, 8)
}

func BuildRedirectURL(originalURL string, utm UTMParameters) string {
	if originalURL == "" {
		return ""
	}
	u, err := url.Parse(originalURL)
	if err != nil {
		return originalURL
	}
	q := u.Query()
	setIfNonEmpty := func(key, val string) {
		if val != "" {
			q.Set(key, val)
		}
	}
	setIfNonEmpty("utm_source", utm.Source)
	setIfNonEmpty("utm_medium", utm.Medium)
	setIfNonEmpty("utm_campaign", utm.Campaign)
	setIfNonEmpty("utm_term", utm.Term)
	setIfNonEmpty("utm_content", utm.Content)
	u.RawQuery = q.Encode()
	return u.String()
}

type GeoLocation struct {
	CountryCode *string
	CountryName *string
	Region      *string
	City        *string
	Latitude    *float64
	Longitude   *float64
	Timezone    *string
}

var countryNames = map[string]string{
	"US": "United States", "GB": "United Kingdom", "CA": "Canada", "AU": "Australia",
	"DE": "Germany", "FR": "France", "IT": "Italy", "ES": "Spain", "NL": "Netherlands",
	"SE": "Sweden", "NO": "Norway", "DK": "Denmark", "FI": "Finland", "PL": "Poland",
	"BR": "Brazil", "MX": "Mexico", "AR": "Argentina", "IN": "India", "CN": "China",
	"JP": "Japan", "KR": "South Korea", "SG": "Singapore", "MY": "Malaysia",
	"TH": "Thailand", "ID": "Indonesia", "PH": "Philippines", "VN": "Vietnam",
	"ZA": "South Africa", "EG": "Egypt", "NG": "Nigeria", "KE": "Kenya", "RU": "Russia",
	"TR": "Turkey", "AE": "United Arab Emirates", "SA": "Saudi Arabia", "IL": "Israel",
	"NZ": "New Zealand", "IE": "Ireland", "CH": "Switzerland", "AT": "Austria",
	"BE": "Belgium", "PT": "Portugal", "GR": "Greece", "CZ": "Czech Republic",
	"HU": "Hungary", "RO": "Romania",
}

func CountryName(code string) string {
	if name, ok := countryNames[code]; ok {
		return name
	}
	return code
}
