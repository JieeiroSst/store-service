package application

import (
	"encoding/json"
	"math/rand"
	"strings"

	"github.com/JIeeiroSst/admanagement-service/internal/domain/model"
)

func matchesTargeting(rules []model.AdTargetingRule, target model.TargetContext) bool {
	for _, rule := range rules {
		if !matchesRule(rule, target) {
			return false
		}
	}
	return true
}

func matchesRule(rule model.AdTargetingRule, target model.TargetContext) bool {
	switch rule.RuleType {
	case model.TargetingRuleTypeCountry:
		return matchString(rule, target.Country)
	case model.TargetingRuleTypeDevice:
		return matchString(rule, target.Device)
	case model.TargetingRuleTypeGender:
		return matchString(rule, target.Gender)
	case model.TargetingRuleTypeAge:
		return matchNumber(rule, float64(target.Age))
	case model.TargetingRuleTypeTime:
		return matchNumber(rule, float64(target.HourOfDay))
	default:
		return true
	}
}

func matchString(rule model.AdTargetingRule, value string) bool {
	switch rule.RuleOperator {
	case model.TargetingOperatorIn:
		var options []string
		if err := json.Unmarshal(rule.RuleValue, &options); err != nil {
			return false
		}
		for _, opt := range options {
			if strings.EqualFold(opt, value) {
				return true
			}
		}
		return false
	default: // equals
		var expected string
		if err := json.Unmarshal(rule.RuleValue, &expected); err != nil {
			return false
		}
		return strings.EqualFold(expected, value)
	}
}

func matchNumber(rule model.AdTargetingRule, value float64) bool {
	switch rule.RuleOperator {
	case model.TargetingOperatorIn:
		var options []float64
		if err := json.Unmarshal(rule.RuleValue, &options); err != nil {
			return false
		}
		for _, opt := range options {
			if opt == value {
				return true
			}
		}
		return false
	case model.TargetingOperatorBetween:
		var bounds [2]float64
		if err := json.Unmarshal(rule.RuleValue, &bounds); err != nil {
			return false
		}
		return value >= bounds[0] && value <= bounds[1]
	case model.TargetingOperatorGreater:
		var bound float64
		if err := json.Unmarshal(rule.RuleValue, &bound); err != nil {
			return false
		}
		return value > bound
	case model.TargetingOperatorLess:
		var bound float64
		if err := json.Unmarshal(rule.RuleValue, &bound); err != nil {
			return false
		}
		return value < bound
	default: // equals
		var bound float64
		if err := json.Unmarshal(rule.RuleValue, &bound); err != nil {
			return false
		}
		return value == bound
	}
}

func pickWeighted(candidates []model.Ad, weights []int) model.Ad {
	total := 0
	for _, w := range weights {
		total += w
	}
	if total <= 0 {
		return candidates[0]
	}

	r := rand.Intn(total)
	for i, w := range weights {
		if r < w {
			return candidates[i]
		}
		r -= w
	}
	return candidates[len(candidates)-1]
}
