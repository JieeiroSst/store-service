package domain

import (
	"testing"
	"time"
)

func d(y int, m time.Month, day int) time.Time { return time.Date(y, m, day, 0, 0, 0, 0, time.UTC) }

func TestBuildScheduleMonthly(t *testing.T) {
	// Starts on the 31st: periods must not drift (Feb 28, then back to Mar 31).
	plans, end, err := BuildSchedule(ModelMonth, d(2027, 1, 31), 4, 5)
	if err != nil {
		t.Fatal(err)
	}
	wantStarts := []time.Time{d(2027, 1, 31), d(2027, 2, 28), d(2027, 3, 31), d(2027, 4, 30)}
	wantDue := []time.Time{d(2027, 1, 31), d(2027, 2, 5), d(2027, 3, 5), d(2027, 4, 5)} // first due at signing, then the 5th
	for i, p := range plans {
		if !p.PeriodStart.Equal(wantStarts[i]) || !p.DueDate.Equal(wantDue[i]) || p.PeriodNo != i {
			t.Errorf("plan %d = %+v, want start %v due %v", i, p, wantStarts[i], wantDue[i])
		}
		if i > 0 && !plans[i-1].PeriodEnd.Equal(p.PeriodStart) {
			t.Errorf("periods must be contiguous at %d", i)
		}
	}
	if !end.Equal(d(2027, 5, 31)) {
		t.Errorf("end = %v", end)
	}
}

func TestBuildScheduleWeeklyAndYearly(t *testing.T) {
	// 2027-03-03 is a Wednesday; collect on Fridays (5).
	plans, end, err := BuildSchedule(ModelWeek, d(2027, 3, 3), 3, 5)
	if err != nil {
		t.Fatal(err)
	}
	if !plans[0].DueDate.Equal(d(2027, 3, 3)) || !plans[1].DueDate.Equal(d(2027, 3, 12)) || !plans[2].DueDate.Equal(d(2027, 3, 19)) {
		t.Errorf("weekly due dates: %v %v %v", plans[0].DueDate, plans[1].DueDate, plans[2].DueDate)
	}
	if isoWeekday(plans[1].DueDate) != 5 || !end.Equal(d(2027, 3, 24)) {
		t.Errorf("weekday/end wrong: %v", end)
	}

	plans, end, err = BuildSchedule(ModelYear, d(2028, 2, 29), 2, 0)
	if err != nil || !plans[1].PeriodStart.Equal(d(2029, 2, 28)) || !end.Equal(d(2030, 2, 28)) {
		t.Errorf("yearly from leap day: %+v %v %v", plans, end, err)
	}
}

func TestBuildScheduleRejects(t *testing.T) {
	bad := []struct {
		m       RentalModel
		periods int
		bd      int
	}{
		{ModelDay, 1, 0}, {ModelMonth, 0, 5}, {ModelMonth, 121, 5}, {ModelMonth, 3, 0}, {ModelMonth, 3, 29},
		{ModelWeek, 3, 8}, {ModelYear, 11, 0},
	}
	for _, b := range bad {
		if _, _, err := BuildSchedule(b.m, d(2027, 1, 1), b.periods, b.bd); err == nil {
			t.Errorf("%v %d %d: want error", b.m, b.periods, b.bd)
		}
	}
}

func TestDefaultBillingDay(t *testing.T) {
	if DefaultBillingDay(ModelMonth, d(2027, 1, 31)) != 28 || DefaultBillingDay(ModelMonth, d(2027, 1, 9)) != 9 ||
		DefaultBillingDay(ModelWeek, d(2027, 3, 7)) != 7 || DefaultBillingDay(ModelYear, d(2027, 1, 1)) != 0 {
		t.Fatal("default billing day")
	}
}

func TestRankScore(t *testing.T) {
	// A single 5-star review must not beat many 4.8s.
	if RankScore(5, 1, 1) >= RankScore(4.8, 50, 1) {
		t.Fatal("one review outranks fifty")
	}
	// Tier breaks near ties in the host's favour.
	if RankScore(4.5, 20, 4) <= RankScore(4.5, 20, 1) {
		t.Fatal("tier boost")
	}
	// ...but does not rescue a clearly worse homestay.
	if RankScore(3.0, 100, 4) >= RankScore(4.6, 100, 1) {
		t.Fatal("boost overturned rating")
	}
	if BayesianRating(0, 0) != 3.5 {
		t.Fatal("no reviews = prior")
	}
}
