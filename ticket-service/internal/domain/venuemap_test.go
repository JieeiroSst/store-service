package domain

import (
	"errors"
	"math"
	"testing"
)

func TestRowLabels(t *testing.T) {
	for n, want := range map[int]string{0: "A", 1: "B", 25: "Z", 26: "AA", 27: "AB", 51: "AZ", 52: "BA", 701: "ZZ", 702: "AAA"} {
		if got := RowLabel(n); got != want {
			t.Errorf("RowLabel(%d) = %q, want %q", n, got, want)
		}
		if got := RowIndex(want); got != n {
			t.Errorf("RowIndex(%q) = %d, want %d", want, got, n)
		}
	}
}

func seatAt(seats []GeneratedSeat, section, row string, n int) (GeneratedSeat, bool) {
	for _, s := range seats {
		if s.Section == section && s.Row == row && s.Number == n {
			return s, true
		}
	}
	return GeneratedSeat{}, false
}

func gen(t *testing.T, m VenueMap) []GeneratedSeat {
	t.Helper()
	seats, err := m.Generate()
	if err != nil {
		t.Fatal(err)
	}
	return seats
}

func TestGridRowsOfDifferentLengthsAreCentred(t *testing.T) {
	m := VenueMap{Width: 500, Height: 300, Sections: []SectionSpec{{Key: "A", Layout: "grid", Origin: Point{100, 50}, RowSeats: []int{3, 5, 4}}}}
	seats := gen(t, m)
	if len(seats) != 12 {
		t.Fatalf("%d seats", len(seats))
	}
	// the widest row (5 seats) spans x = 100..220; a row of 3 is centred on it: 130, 160, 190
	for k, x := range []float64{130, 160, 190} {
		if s, ok := seatAt(seats, "A", "A", k+1); !ok || s.X != x || s.Y != 50 {
			t.Errorf("A-%d = %+v (%v), want x=%v y=50", k+1, s, ok, x)
		}
	}
	if s, _ := seatAt(seats, "A", "B", 1); s.X != 100 || s.Y != 84 {
		t.Errorf("B-1 = %+v", s)
	}
	m.Sections[0].Align = "left"
	if s, _ := seatAt(gen(t, m), "A", "A", 1); s.X != 100 {
		t.Errorf("left-aligned A-1 at %v", s.X)
	}
	m.Sections[0].Align = "right"
	if s, _ := seatAt(gen(t, m), "A", "A", 3); s.X != 220 {
		t.Errorf("right-aligned A-3 at %v", s.X)
	}
}

func TestGridAislesSkipAccessibleAndNumbering(t *testing.T) {
	m := VenueMap{Width: 500, Height: 300, Sections: []SectionSpec{{
		Key: "L", Layout: "grid", Origin: Point{10, 10}, Rows: 2, SeatsPerRow: 6, AisleAfter: []int{3},
		Skip: []string{"c:2"}, Accessible: []string{"D:1"}, RowStart: "C", NumberFrom: "right"}}}
	seats := gen(t, m)
	if len(seats) != 11 {
		t.Fatalf("%d seats, want 11 (one skipped)", len(seats))
	}
	// rows start at C; numbered from the right, so the leftmost seat is number 6
	if left, ok := seatAt(seats, "L", "D", 6); !ok || left.X != 10 {
		t.Errorf("D-6 (the leftmost) = %+v %v", left, ok)
	}
	// a skipped seat is absent and its number is not reused
	if _, ok := seatAt(seats, "L", "C", 2); ok {
		t.Error("C-2 should be skipped")
	}
	if _, ok := seatAt(seats, "L", "C", 1); !ok {
		t.Error("C-1 must exist")
	}
	// the aisle: visual positions 3 and 4 (numbers 4 and 3) are two pitches apart, not one
	s3, _ := seatAt(seats, "L", "D", 4)
	s4, _ := seatAt(seats, "L", "D", 3)
	if d := s4.X - s3.X; d != 60 {
		t.Errorf("across the aisle: %v, want 60", d)
	}
	if d1, _ := seatAt(seats, "L", "D", 1); !d1.Accessible {
		t.Error("D-1 is a wheelchair place")
	}
	acc := 0
	for _, s := range seats {
		if s.Accessible {
			acc++
		}
	}
	if acc != 1 {
		t.Errorf("%d accessible seats, want 1", acc)
	}
}

func TestGridRotation(t *testing.T) {
	// rows extend to the right when the block is turned -90 degrees; seats run upward
	m := VenueMap{Width: 400, Height: 400, Sections: []SectionSpec{{Key: "E", Layout: "grid", Origin: Point{100, 300}, Rows: 2, SeatsPerRow: 3, Rotation: -90}}}
	seats := gen(t, m)
	a1, _ := seatAt(seats, "E", "A", 1)
	a3, _ := seatAt(seats, "E", "A", 3)
	b1, _ := seatAt(seats, "E", "B", 1)
	if a1.X != 100 || a1.Y != 300 || a3.X != 100 || a3.Y != 240 || b1.X != 134 || b1.Y != 300 {
		t.Fatalf("A-1 %+v A-3 %+v B-1 %+v", a1, a3, b1)
	}
}

func TestArcRowsWidenWithRadius(t *testing.T) {
	m := VenueMap{Width: 900, Height: 700, Sections: []SectionSpec{{Key: "F", Layout: "arc", Center: Point{450, 50}, Radius: 240, Rows: 4, AngleStart: 35, AngleEnd: 145}}}
	seats := gen(t, m)
	per := map[string]int{}
	for _, s := range seats {
		per[s.Row]++
	}
	if !(per["A"] < per["B"] && per["B"] < per["C"] && per["C"] < per["D"]) {
		t.Fatalf("rows on a fan get longer towards the back: %v", per)
	}
	for _, s := range seats { // every seat is on its row's circle
		r := 240 + float64(RowIndex(s.Row))*34
		if d := math.Hypot(s.X-450, s.Y-50); math.Abs(d-r) > 0.02 {
			t.Fatalf("%s-%d is %.2f from the focus, want %.2f", s.Row, s.Number, d, r)
		}
	}
	// seats in a row are at least about a seat pitch apart
	for _, row := range []string{"A", "D"} {
		var prev *GeneratedSeat
		for i := range seats {
			if seats[i].Row != row {
				continue
			}
			if prev != nil && math.Hypot(seats[i].X-prev.X, seats[i].Y-prev.Y) < 29 {
				t.Fatalf("row %s is too tight at seat %d", row, seats[i].Number)
			}
			prev = &seats[i]
		}
	}
	m.Sections[0].SeatsPerRow = 10
	if len(gen(t, m)) != 40 {
		t.Fatal("fixed seats per row")
	}
	m.Sections[0].SeatsPerRow, m.Sections[0].RowSeats = 0, []int{5, 7}
	if len(gen(t, m)) != 12 {
		t.Fatal("row_seats")
	}
}

func manySections(n int) []SectionSpec {
	var out []SectionSpec
	for i := 0; i < n; i++ {
		out = append(out, SectionSpec{Key: RowLabel(i), Layout: "grid", Origin: Point{float64(i) * 1500, 10}, Rows: 100, SeatsPerRow: 100})
	}
	return out
}

func TestGenerateRefusesWhatCannotBeDrawn(t *testing.T) {
	ok := SectionSpec{Key: "A", Layout: "grid", Origin: Point{10, 10}, Rows: 2, SeatsPerRow: 3}
	cases := map[string]VenueMap{
		"no canvas":          {Sections: []SectionSpec{ok}},
		"no sections":        {Width: 100, Height: 100},
		"duplicate keys":     {Width: 500, Height: 500, Sections: []SectionSpec{ok, ok}},
		"bad key":            {Width: 500, Height: 500, Sections: []SectionSpec{{Key: "a b", Layout: "grid", Rows: 1, SeatsPerRow: 1}}},
		"unknown layout":     {Width: 500, Height: 500, Sections: []SectionSpec{{Key: "A", Layout: "spiral", Rows: 1, SeatsPerRow: 1}}},
		"no rows":            {Width: 500, Height: 500, Sections: []SectionSpec{{Key: "A", Layout: "grid"}}},
		"seats off canvas":   {Width: 60, Height: 60, Sections: []SectionSpec{ok}},
		"bad skip":           {Width: 500, Height: 500, Sections: []SectionSpec{{Key: "A", Layout: "grid", Rows: 1, SeatsPerRow: 3, Skip: []string{"x"}}}},
		"bad pitch":          {Width: 500, Height: 500, Sections: []SectionSpec{{Key: "A", Layout: "grid", Rows: 1, SeatsPerRow: 3, SeatPitch: 1}}},
		"bad align":          {Width: 500, Height: 500, Sections: []SectionSpec{{Key: "A", Layout: "grid", Rows: 1, SeatsPerRow: 3, Align: "diagonal"}}},
		"arc without angle":  {Width: 500, Height: 500, Sections: []SectionSpec{{Key: "A", Layout: "arc", Radius: 100, Rows: 1, SeatsPerRow: 3}}},
		"arc without radius": {Width: 500, Height: 500, Sections: []SectionSpec{{Key: "A", Layout: "arc", AngleEnd: 90, Rows: 1, SeatsPerRow: 3}}},
		"row too long":       {Width: 1e5, Height: 1e5, Sections: []SectionSpec{{Key: "A", Layout: "grid", Rows: 1, SeatsPerRow: 301}}},
		"skip everything":    {Width: 500, Height: 500, Sections: []SectionSpec{{Key: "A", Layout: "grid", Origin: Point{1, 1}, Rows: 1, SeatsPerRow: 1, Skip: []string{"A:1"}}}},
		"stage outside":      {Width: 500, Height: 500, Stages: []StageShape{{X: 450, Y: 0, W: 100, H: 10}}, Sections: []SectionSpec{ok}},
		"bad stage shape":    {Width: 500, Height: 500, Stages: []StageShape{{Shape: "star"}}, Sections: []SectionSpec{ok}},
		"outline outside":    {Width: 500, Height: 500, Sections: []SectionSpec{{Key: "A", Layout: "grid", Origin: Point{10, 10}, Rows: 1, SeatsPerRow: 3, Outline: [][2]float64{{0, 0}, {10, 0}, {900, 5}}}}},
		"too many seats":     {Width: 1e5, Height: 1e5, Sections: manySections(6)},
	}
	for name, m := range cases {
		if _, err := m.Generate(); !errors.Is(err, ErrInvalid) {
			t.Errorf("%s: %v", name, err)
		}
	}
}

func TestTemplatesProduceDrawableVenues(t *testing.T) {
	want := map[string]int{"theatre": 300, "hall": 200, "arena": 448}
	for _, tpl := range VenueTemplates() {
		m, err := BuildVenueTemplate(tpl.Name, nil)
		if err != nil {
			t.Fatalf("%s: %v", tpl.Name, err)
		}
		seats, err := m.Generate()
		if err != nil {
			t.Fatalf("%s: %v", tpl.Name, err)
		}
		if len(seats) != want[tpl.Name] {
			t.Errorf("%s makes %d seats, want %d", tpl.Name, len(seats), want[tpl.Name])
		}
		// no two seats on top of each other, so a client can draw them
		cell := map[[2]int]GeneratedSeat{}
		for _, s := range seats {
			k := [2]int{int(s.X / 10), int(s.Y / 10)}
			for dx := -3; dx <= 3; dx++ {
				for dy := -3; dy <= 3; dy++ {
					if o, ok := cell[[2]int{k[0] + dx, k[1] + dy}]; ok && math.Hypot(o.X-s.X, o.Y-s.Y) < 20 {
						t.Fatalf("%s: %s %s-%d and %s %s-%d overlap", tpl.Name, o.Section, o.Row, o.Number, s.Section, s.Row, s.Number)
					}
				}
			}
			cell[k] = s
		}
		v := m.View(seats, nil)
		if len(v.Sections) != len(m.Sections) || len(v.Stages) == 0 {
			t.Errorf("%s: view has %d sections, %d stages", tpl.Name, len(v.Sections), len(v.Stages))
		}
	}
	// a theatre's seats are never closer than a seat pitch, in any row
	tm, _ := BuildVenueTemplate("theatre", nil)
	ts, _ := tm.Generate()
	for i := 1; i < len(ts); i++ {
		if ts[i].Row == ts[i-1].Row && math.Hypot(ts[i].X-ts[i-1].X, ts[i].Y-ts[i-1].Y) < 30 {
			t.Fatalf("theatre row %s is too tight at seat %d", ts[i].Row, ts[i].Number)
		}
	}
	// the templates differ in what they have
	th, _ := BuildVenueTemplate("theatre", map[string]int{"balcony_rows": 3, "rows": 6})
	if len(th.Sections) != 2 {
		t.Error("a theatre with a balcony has two sections")
	}
	if _, err := th.Generate(); err != nil {
		t.Errorf("theatre with a balcony: %v", err)
	}
	ar, _ := BuildVenueTemplate("arena", map[string]int{"rows": 3, "seats_per_row": 5})
	seats, err := ar.Generate()
	if err != nil || len(seats) != 60 || ar.Stages[0].Shape != "ellipse" {
		t.Errorf("arena: %d seats, stage %v (%v)", len(seats), ar.Stages[0].Shape, err)
	}
	// the four blocks of an arena sit on four sides of the stage
	side := func(key string) GeneratedSeat { s, _ := seatAt(seats, key, "A", 1); return s }
	cx, cy := ar.Width/2, ar.Height/2
	if !(side("S").Y > cy && side("N").Y < cy && side("E").X > cx && side("W").X < cx) {
		t.Errorf("arena blocks: S %+v N %+v E %+v W %+v centre %v,%v", side("S"), side("N"), side("E"), side("W"), cx, cy)
	}
	for _, bad := range []struct {
		name string
		p    map[string]int
	}{{"nope", nil}, {"hall", map[string]int{"rows": 0}}, {"hall", map[string]int{"colour": 1}}, {"hall", map[string]int{"rows": 1000}}} {
		if _, err := BuildVenueTemplate(bad.name, bad.p); !errors.Is(err, ErrInvalid) {
			t.Errorf("%s %v: %v", bad.name, bad.p, err)
		}
	}
}

func TestViewGivesEverySectionAnOutlineAndItsTicketType(t *testing.T) {
	m := VenueMap{Width: 500, Height: 300, Sections: []SectionSpec{
		{Key: "A", Name: "Front", Layout: "grid", Origin: Point{100, 50}, Rows: 2, SeatsPerRow: 4},
		{Key: "B", Layout: "grid", Origin: Point{100, 200}, Rows: 1, SeatsPerRow: 2, Outline: [][2]float64{{90, 190}, {200, 190}, {200, 240}}}}}
	seats := gen(t, m)
	v := m.View(seats, map[string]int64{"A": 7})
	if len(v.Sections) != 2 || v.Sections[0].TicketTypeID != 7 || v.Sections[1].TicketTypeID != 0 || v.Sections[1].Name != "B" {
		t.Fatalf("%+v", v.Sections)
	}
	box := v.Sections[0].Points // 100..190 x 50..84, padded by half a pitch
	if len(box) != 4 || box[0] != [2]float64{85, 35} || box[2] != [2]float64{205, 99} {
		t.Fatalf("outline of a section without one: %v", box)
	}
	if len(v.Sections[1].Points) != 3 {
		t.Fatal("an explicit outline is kept")
	}
}

func TestGenerateIsDeterministic(t *testing.T) {
	m, _ := BuildVenueTemplate("theatre", map[string]int{"balcony_rows": 4})
	a, b := gen(t, m), gen(t, m)
	for i := range a {
		if a[i] != b[i] {
			t.Fatal("the same blueprint produced different seats")
		}
	}
}
