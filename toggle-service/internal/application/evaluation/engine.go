package evaluation

import (
	"encoding/json"
	"net"
	"strings"

	"gorm.io/datatypes"

	"github.com/JIeeiroSst/toggle-service/internal/domain/model"
)

type Engine struct{}

func NewEngine() *Engine { return &Engine{} }

func (e *Engine) Evaluate(flagKey string, ffEnv *model.FeatureFlagEnvironment, strategies []model.ActivationStrategy, ctx model.EvaluationContext) bool {
	if ffEnv == nil || !ffEnv.Enabled {
		return false
	}
	if len(strategies) == 0 {
		return false
	}
	for _, s := range strategies {
		if e.evaluateStrategy(flagKey, s, ctx) {
			return true
		}
	}
	return false
}

func (e *Engine) evaluateStrategy(flagKey string, s model.ActivationStrategy, ctx model.EvaluationContext) bool {
	if !constraintsPass(s.Constraints, ctx) {
		return false
	}
	switch s.StrategyType {
	case model.StrategyDefault:
		return true
	case model.StrategyFlexibleRollout:
		return evaluateFlexibleRollout(flagKey, s.Parameters, ctx)
	case model.StrategyUserWithID:
		return evaluateUserWithID(s.Parameters, ctx)
	case model.StrategyRemoteAddress:
		return evaluateRemoteAddress(s.Parameters, ctx)
	default:
		return false
	}
}

func constraintsPass(constraints []model.Constraint, ctx model.EvaluationContext) bool {
	for _, c := range constraints {
		if !constraintPasses(c, ctx) {
			return false
		}
	}
	return true
}

func constraintPasses(c model.Constraint, ctx model.EvaluationContext) bool {
	fieldValue, ok := ctx.FieldValue(c.ContextField)

	var values []string
	_ = json.Unmarshal(c.Values, &values)

	switch c.Operator {
	case model.OperatorIn:
		return ok && containsString(values, fieldValue, c.CaseInsensitive)
	case model.OperatorNotIn:
		return !ok || !containsString(values, fieldValue, c.CaseInsensitive)
	case model.OperatorStrContains:
		if !ok {
			return false
		}
		for _, v := range values {
			if c.CaseInsensitive {
				if strings.Contains(strings.ToLower(fieldValue), strings.ToLower(v)) {
					return true
				}
			} else if strings.Contains(fieldValue, v) {
				return true
			}
		}
		return false
	default:
		return false
	}
}

func containsString(values []string, target string, caseInsensitive bool) bool {
	for _, v := range values {
		if caseInsensitive {
			if strings.EqualFold(v, target) {
				return true
			}
		} else if v == target {
			return true
		}
	}
	return false
}

func evaluateUserWithID(params datatypes.JSON, ctx model.EvaluationContext) bool {
	var p model.UserWithIDParams
	_ = json.Unmarshal(params, &p)
	if ctx.UserID == "" {
		return false
	}
	for _, id := range p.UserIDs {
		if id == ctx.UserID {
			return true
		}
	}
	return false
}

func evaluateRemoteAddress(params datatypes.JSON, ctx model.EvaluationContext) bool {
	var p model.RemoteAddressParams
	_ = json.Unmarshal(params, &p)
	if ctx.RemoteAddress == "" {
		return false
	}
	ip := net.ParseIP(ctx.RemoteAddress)
	for _, entry := range p.IPs {
		if strings.Contains(entry, "/") {
			if _, cidr, err := net.ParseCIDR(entry); err == nil && ip != nil && cidr.Contains(ip) {
				return true
			}
		} else if entry == ctx.RemoteAddress {
			return true
		}
	}
	return false
}

func evaluateFlexibleRollout(flagKey string, params datatypes.JSON, ctx model.EvaluationContext) bool {
	var p model.FlexibleRolloutParams
	_ = json.Unmarshal(params, &p)
	if p.Percentage <= 0 {
		return false
	}
	if p.Percentage >= 100 {
		return true
	}

	stickinessValue := stickinessValueFor(p.Stickiness, ctx)
	if stickinessValue == "" {
		return false
	}
	return stickinessHash(flagKey, stickinessValue) < uint32(p.Percentage)*100
}

func stickinessValueFor(stickiness string, ctx model.EvaluationContext) string {
	switch stickiness {
	case "sessionId":
		return ctx.SessionID
	case "random":
		return randomStickiness()
	default: // "userId" or unset
		if ctx.UserID != "" {
			return ctx.UserID
		}
		return ctx.SessionID
	}
}
