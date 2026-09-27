package handler

import (
	"testing"

	"github.com/JIeeiroSst/tool-service/internal/learning"
	"github.com/JIeeiroSst/tool-service/internal/ollama"
)

func addReviewed(t *testing.T, s *learning.Store, model string, good, bad int) {
	t.Helper()
	for i := 0; i < good+bad; i++ {
		e := &learning.Experience{Kind: learning.KindTestCases, Model: model}
		if err := s.Add(e); err != nil {
			t.Fatal(err)
		}
		v := learning.VerdictBad
		if i < good {
			v = learning.VerdictGood
		}
		if err := s.Feedback(e.ID, v, "", nil); err != nil {
			t.Fatal(err)
		}
	}
}

func TestGuardRollback(t *testing.T) {
	cases := []struct {
		name         string
		baseG, baseB int
		newG, newB   int
		wantModel    string
	}{
		{"worse model is rolled back", 9, 1, 5, 5, "base"},
		{"better model is kept", 5, 5, 9, 1, "evolved"},
		{"too few reviews is kept", 9, 1, 0, 3, "evolved"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mem, err := learning.Open(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			addReviewed(t, mem, "base", tc.baseG, tc.baseB)
			addReviewed(t, mem, "evolved", tc.newG, tc.newB)
			ai := ollama.NewClient("http://unused", "evolved", 0)
			h := New(ai, mem, LearnConfig{BaseModel: "base"}, nil)
			h.guardRollback()
			if got := ai.Model(); got != tc.wantModel {
				t.Fatalf("model = %s, want %s", got, tc.wantModel)
			}
		})
	}
}
