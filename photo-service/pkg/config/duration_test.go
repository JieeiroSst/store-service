package config

import (
	"encoding/json"
	"testing"
	"time"
)

func TestDuration_UnmarshalJSON(t *testing.T) {
	var d Duration
	if err := json.Unmarshal([]byte(`"10s"`), &d); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if time.Duration(d) != 10*time.Second {
		t.Fatalf("expected 10s, got %s", time.Duration(d))
	}
}

func TestDuration_UnmarshalJSON_InvalidString(t *testing.T) {
	var d Duration
	if err := json.Unmarshal([]byte(`"not-a-duration"`), &d); err == nil {
		t.Fatal("expected an error for an invalid duration string")
	}
}

func TestDuration_UnmarshalJSON_RejectsRawNumber(t *testing.T) {
	// Durations must be authored as human strings ("10s"), not raw nanoseconds.
	var d Duration
	if err := json.Unmarshal([]byte(`10`), &d); err == nil {
		t.Fatal("expected an error for a raw numeric duration")
	}
}

func TestDuration_RoundTrip(t *testing.T) {
	want := Duration(90 * time.Minute)
	data, err := json.Marshal(want)
	if err != nil {
		t.Fatalf("unexpected error marshaling: %v", err)
	}

	var got Duration
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unexpected error unmarshaling: %v", err)
	}
	if got != want {
		t.Fatalf("round trip mismatch: got %s, want %s", time.Duration(got), time.Duration(want))
	}
}
