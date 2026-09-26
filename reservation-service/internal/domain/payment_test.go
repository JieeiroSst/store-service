package domain

import "testing"

func TestMinorUnits(t *testing.T) {
	tests := []struct {
		amount, cur string
		want        int64
		wantErr     bool
	}{
		{"300001.500000", "VND", 300002, false},
		{"100000", "VND", 100000, false},
		{"12.345", "USD", 1235, false},
		{"0.10", "usd", 10, false},
		{"abc", "USD", 0, true},
		{"-1", "USD", 0, true},
	}
	for _, tc := range tests {
		got, err := MinorUnits(tc.amount, tc.cur)
		if (err != nil) != tc.wantErr || got != tc.want {
			t.Errorf("MinorUnits(%q,%q) = %d, %v; want %d, err=%v", tc.amount, tc.cur, got, err, tc.want, tc.wantErr)
		}
	}
}
