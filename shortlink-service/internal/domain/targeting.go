package domain

import "strings"

type TargetingContext struct {
	Device          DeviceType
	CountryCode     string // "" when unknown
	PrimaryLanguage string // "" when unknown, already lowercased primary subtag
}

func EvaluateTargeting(rules TargetingRules, ctx TargetingContext) bool {
	if len(rules.Countries) > 0 {
		matched := false
		if ctx.CountryCode != "" {
			target := ctx.CountryCode
			for _, c := range rules.Countries {
				if strings.EqualFold(c, target) {
					matched = true
					break
				}
			}
		}
		if !matched {
			return false
		}
	}

	if len(rules.Devices) > 0 {
		matched := false
		for _, d := range rules.Devices {
			if strings.EqualFold(d, string(ctx.Device)) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	if len(rules.Languages) > 0 {
		matched := false
		if ctx.PrimaryLanguage != "" {
			for _, l := range rules.Languages {
				if strings.EqualFold(l, ctx.PrimaryLanguage) {
					matched = true
					break
				}
			}
		}
		if !matched {
			return false
		}
	}

	return true
}

func PrimaryLanguageFromAcceptLanguage(acceptLanguage string) string {
	first := strings.Split(acceptLanguage, ",")[0]
	first = strings.Split(first, "-")[0]
	return strings.ToLower(strings.TrimSpace(first))
}

func FingerprintLanguageFromAcceptLanguage(acceptLanguage string) string {
	first := strings.Split(acceptLanguage, ",")[0]
	first = strings.Split(first, ";")[0]
	return strings.TrimSpace(first)
}
