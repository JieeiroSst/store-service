package evaluation

import (
	"encoding/json"
	"fmt"
	"testing"

	"gorm.io/datatypes"

	"github.com/JIeeiroSst/toggle-service/internal/domain/model"
)

func mustJSON(t *testing.T, v any) datatypes.JSON {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return datatypes.JSON(b)
}

func enabledEnv() *model.FeatureFlagEnvironment {
	return &model.FeatureFlagEnvironment{Enabled: true}
}

func TestEvaluate_DisabledEnvironment(t *testing.T) {
	e := NewEngine()
	strategies := []model.ActivationStrategy{{StrategyType: model.StrategyDefault}}
	got := e.Evaluate("flag", &model.FeatureFlagEnvironment{Enabled: false}, strategies, model.EvaluationContext{})
	if got {
		t.Fatal("expected disabled environment to yield false regardless of strategies")
	}
}

func TestEvaluate_NoStrategies(t *testing.T) {
	e := NewEngine()
	got := e.Evaluate("flag", enabledEnv(), nil, model.EvaluationContext{})
	if got {
		t.Fatal("expected no strategies to yield false")
	}
}

func TestEvaluate_DefaultStrategy(t *testing.T) {
	e := NewEngine()
	strategies := []model.ActivationStrategy{{StrategyType: model.StrategyDefault}}
	if !e.Evaluate("flag", enabledEnv(), strategies, model.EvaluationContext{}) {
		t.Fatal("expected default strategy to always enable")
	}
}

func TestEvaluate_UserWithID(t *testing.T) {
	e := NewEngine()
	strategies := []model.ActivationStrategy{{
		StrategyType: model.StrategyUserWithID,
		Parameters:   mustJSON(t, model.UserWithIDParams{UserIDs: []string{"42", "99"}}),
	}}

	if !e.Evaluate("flag", enabledEnv(), strategies, model.EvaluationContext{UserID: "42"}) {
		t.Error("expected matching userId to enable")
	}
	if e.Evaluate("flag", enabledEnv(), strategies, model.EvaluationContext{UserID: "7"}) {
		t.Error("expected non-matching userId to disable")
	}
	if e.Evaluate("flag", enabledEnv(), strategies, model.EvaluationContext{}) {
		t.Error("expected empty userId to disable")
	}
}

func TestEvaluate_RemoteAddress(t *testing.T) {
	e := NewEngine()
	strategies := []model.ActivationStrategy{{
		StrategyType: model.StrategyRemoteAddress,
		Parameters:   mustJSON(t, model.RemoteAddressParams{IPs: []string{"127.0.0.1", "10.0.0.0/8"}}),
	}}

	cases := []struct {
		ip   string
		want bool
	}{
		{"127.0.0.1", true},
		{"10.1.2.3", true}, // inside 10.0.0.0/8
		{"192.168.1.1", false},
		{"", false},
	}
	for _, c := range cases {
		got := e.Evaluate("flag", enabledEnv(), strategies, model.EvaluationContext{RemoteAddress: c.ip})
		if got != c.want {
			t.Errorf("ip %q: got %v, want %v", c.ip, got, c.want)
		}
	}
}

func TestEvaluate_FlexibleRolloutBoundaries(t *testing.T) {
	e := NewEngine()

	zero := []model.ActivationStrategy{{
		StrategyType: model.StrategyFlexibleRollout,
		Parameters:   mustJSON(t, model.FlexibleRolloutParams{Percentage: 0, Stickiness: "userId"}),
	}}
	if e.Evaluate("flag", enabledEnv(), zero, model.EvaluationContext{UserID: "u1"}) {
		t.Error("0% rollout should never enable")
	}

	hundred := []model.ActivationStrategy{{
		StrategyType: model.StrategyFlexibleRollout,
		Parameters:   mustJSON(t, model.FlexibleRolloutParams{Percentage: 100, Stickiness: "userId"}),
	}}
	if !e.Evaluate("flag", enabledEnv(), hundred, model.EvaluationContext{UserID: "u1"}) {
		t.Error("100% rollout should always enable")
	}
}

func TestEvaluate_FlexibleRolloutDeterministicStickiness(t *testing.T) {
	e := NewEngine()
	strategies := []model.ActivationStrategy{{
		StrategyType: model.StrategyFlexibleRollout,
		Parameters:   mustJSON(t, model.FlexibleRolloutParams{Percentage: 50, Stickiness: "userId"}),
	}}
	ctx := model.EvaluationContext{UserID: "stable-user"}

	first := e.Evaluate("flag", enabledEnv(), strategies, ctx)
	for i := 0; i < 20; i++ {
		if got := e.Evaluate("flag", enabledEnv(), strategies, ctx); got != first {
			t.Fatalf("expected deterministic result for same userId+flag, got %v then %v", first, got)
		}
	}
}

func TestEvaluate_FlexibleRolloutDistribution(t *testing.T) {
	e := NewEngine()
	strategies := []model.ActivationStrategy{{
		StrategyType: model.StrategyFlexibleRollout,
		Parameters:   mustJSON(t, model.FlexibleRolloutParams{Percentage: 50, Stickiness: "userId"}),
	}}

	enabledCount := 0
	const total = 2000
	for i := 0; i < total; i++ {
		ctx := model.EvaluationContext{UserID: fmt.Sprintf("user-%d", i)}
		if e.Evaluate("flag", enabledEnv(), strategies, ctx) {
			enabledCount++
		}
	}
	ratio := float64(enabledCount) / float64(total)
	if ratio < 0.4 || ratio > 0.6 {
		t.Fatalf("expected roughly 50%% enabled across many users, got %.2f%%", ratio*100)
	}
}

func TestEvaluate_ConstraintOperators(t *testing.T) {
	e := NewEngine()

	inStrategy := []model.ActivationStrategy{{
		StrategyType: model.StrategyDefault,
		Constraints: []model.Constraint{{
			ContextField: "country",
			Operator:     model.OperatorIn,
			Values:       mustJSON(t, []string{"VN", "US"}),
		}},
	}}
	if !e.Evaluate("flag", enabledEnv(), inStrategy, model.EvaluationContext{Properties: map[string]string{"country": "VN"}}) {
		t.Error("IN: expected VN to match [VN, US]")
	}
	if e.Evaluate("flag", enabledEnv(), inStrategy, model.EvaluationContext{Properties: map[string]string{"country": "FR"}}) {
		t.Error("IN: expected FR to not match [VN, US]")
	}

	notInStrategy := []model.ActivationStrategy{{
		StrategyType: model.StrategyDefault,
		Constraints: []model.Constraint{{
			ContextField: "country",
			Operator:     model.OperatorNotIn,
			Values:       mustJSON(t, []string{"VN", "US"}),
		}},
	}}
	if e.Evaluate("flag", enabledEnv(), notInStrategy, model.EvaluationContext{Properties: map[string]string{"country": "VN"}}) {
		t.Error("NOT_IN: expected VN to be excluded")
	}
	if !e.Evaluate("flag", enabledEnv(), notInStrategy, model.EvaluationContext{Properties: map[string]string{"country": "FR"}}) {
		t.Error("NOT_IN: expected FR to pass")
	}

	containsStrategy := []model.ActivationStrategy{{
		StrategyType: model.StrategyDefault,
		Constraints: []model.Constraint{{
			ContextField:    "email",
			Operator:        model.OperatorStrContains,
			Values:          mustJSON(t, []string{"@ACME.COM"}),
			CaseInsensitive: true,
		}},
	}}
	if !e.Evaluate("flag", enabledEnv(), containsStrategy, model.EvaluationContext{Properties: map[string]string{"email": "a@acme.com"}}) {
		t.Error("STR_CONTAINS (case-insensitive): expected match")
	}
	if e.Evaluate("flag", enabledEnv(), containsStrategy, model.EvaluationContext{Properties: map[string]string{"email": "a@other.com"}}) {
		t.Error("STR_CONTAINS: expected no match for different domain")
	}
}

func TestEvaluate_ConstraintsAndAcrossStrategiesOr(t *testing.T) {
	e := NewEngine()

	// Strategy 1 requires country=VN (fails for FR) AND is default otherwise-enable.
	// Strategy 2 is unconditional default. OR across strategies means the flag
	// should still be enabled even though strategy 1's constraint fails.
	strategies := []model.ActivationStrategy{
		{
			StrategyType: model.StrategyDefault,
			Constraints: []model.Constraint{{
				ContextField: "country",
				Operator:     model.OperatorIn,
				Values:       mustJSON(t, []string{"VN"}),
			}},
		},
		{StrategyType: model.StrategyDefault},
	}

	if !e.Evaluate("flag", enabledEnv(), strategies, model.EvaluationContext{Properties: map[string]string{"country": "FR"}}) {
		t.Error("expected OR-across-strategies: unconditional second strategy should still enable")
	}
}
