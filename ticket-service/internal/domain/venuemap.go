package domain

import (
	"fmt"
	"math"
	"regexp"
	"strings"
)

func RowLabel(n int) string {
	var b []byte
	for n++; n > 0; n = (n - 1) / 26 {
		b = append([]byte{byte('A' + (n-1)%26)}, b...)
	}
	return string(b)
}

func RowIndex(label string) int {
	n := 0
	for _, c := range label {
		n = n*26 + int(c-'A') + 1
	}
	return n - 1
}

type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type StageShape struct {
	Label  string       `json:"label,omitempty"`
	Shape  string       `json:"shape,omitempty"`
	X      float64      `json:"x"`
	Y      float64      `json:"y"`
	W      float64      `json:"w"`
	H      float64      `json:"h"`
	Points [][2]float64 `json:"points,omitempty"`
}

type SectionSpec struct {
	Key         string       `json:"key"`
	Name        string       `json:"name"`
	Label       string       `json:"label,omitempty"`
	Color       string       `json:"color,omitempty"`
	Layout      string       `json:"layout"`
	Outline     [][2]float64 `json:"outline,omitempty"`
	Rows        int          `json:"rows,omitempty"`
	SeatsPerRow int          `json:"seats_per_row,omitempty"`
	RowSeats    []int        `json:"row_seats,omitempty"`
	RowStart    string       `json:"row_start,omitempty"`
	NumberFrom  string       `json:"number_from,omitempty"`
	Skip        []string     `json:"skip,omitempty"`
	Accessible  []string     `json:"accessible,omitempty"`
	Origin      Point        `json:"origin"`
	SeatPitch   float64      `json:"seat_pitch,omitempty"`
	RowPitch    float64      `json:"row_pitch,omitempty"`
	Align       string       `json:"align,omitempty"`
	AisleAfter  []int        `json:"aisle_after,omitempty"`
	Rotation    float64      `json:"rotation,omitempty"`
	Center      Point        `json:"center"`
	Radius      float64      `json:"radius,omitempty"`
	AngleStart  float64      `json:"angle_start,omitempty"`
	AngleEnd    float64      `json:"angle_end,omitempty"`
}

type VenueMap struct {
	Width    float64       `json:"width"`
	Height   float64       `json:"height"`
	Stages   []StageShape  `json:"stages,omitempty"`
	Sections []SectionSpec `json:"sections"`
}

type GeneratedSeat struct {
	Section    string
	Row        string
	Number     int
	X, Y       float64
	Accessible bool
}

const (
	MaxVenueSeats     = 50_000
	maxVenueSections  = 100
	maxRowsPerSection = 200
	maxSeatsPerRow    = 300
	maxVenueStages    = 10
	defaultSeatPitch  = 30.0
	defaultRowPitch   = 34.0
)

var sectionKeyRE = regexp.MustCompile(`^[A-Za-z0-9_-]{1,20}$`)

func bad(format string, a ...any) error {
	return fmt.Errorf("%w: "+format, append([]any{ErrInvalid}, a...)...)
}

func seatKey(row string, n int) string { return fmt.Sprintf("%s:%d", strings.ToUpper(row), n) }

func parseSeatKeys(list []string) (map[string]bool, error) {
	out := make(map[string]bool, len(list))
	for _, k := range list {
		row, num, ok := strings.Cut(k, ":")
		var n int
		if _, err := fmt.Sscanf(num, "%d", &n); !ok || row == "" || err != nil || n < 1 {
			return nil, bad("%q is not a seat: write it as row:number, like B:7", k)
		}
		out[seatKey(row, n)] = true
	}
	return out, nil
}

func (s SectionSpec) rowSeats() ([]int, error) {
	if len(s.RowSeats) > 0 {
		return s.RowSeats, nil
	}
	if s.Rows <= 0 || s.SeatsPerRow <= 0 {
		return nil, bad("section %q needs rows and seats_per_row, or row_seats", s.Key)
	}
	out := make([]int, s.Rows)
	for i := range out {
		out[i] = s.SeatsPerRow
	}
	return out, nil
}

func rad(deg float64) float64 { return deg * math.Pi / 180 }

func (m VenueMap) Generate() ([]GeneratedSeat, error) {
	if !finite(m.Width, m.Height) || m.Width <= 0 || m.Height <= 0 || m.Width > maxMapSize || m.Height > maxMapSize {
		return nil, bad("width and height must be between 0 and %d", maxMapSize)
	}
	if len(m.Sections) == 0 || len(m.Sections) > maxVenueSections {
		return nil, bad("a venue has between 1 and %d sections", maxVenueSections)
	}
	if len(m.Stages) > maxVenueStages {
		return nil, bad("at most %d stages", maxVenueStages)
	}
	for i, st := range m.Stages {
		if err := st.validate(m.Width, m.Height); err != nil {
			return nil, fmt.Errorf("stage %d: %w", i+1, err)
		}
	}
	seen := map[string]bool{}
	var all []GeneratedSeat
	for _, sec := range m.Sections {
		if !sectionKeyRE.MatchString(sec.Key) {
			return nil, bad("a section key is 1-20 letters, digits, - or _ (got %q)", sec.Key)
		}
		if seen[sec.Key] {
			return nil, bad("section key %q is used twice", sec.Key)
		}
		seen[sec.Key] = true
		if len(sec.Name) > 60 || len(sec.Label) > 60 || len(sec.Color) > 20 {
			return nil, bad("section %q: name, label or color is too long", sec.Key)
		}
		seats, err := sec.generate()
		if err != nil {
			return nil, err
		}
		if len(all)+len(seats) > MaxVenueSeats {
			return nil, bad("a venue has at most %d seats", MaxVenueSeats)
		}
		for _, s := range seats {
			if !finite(s.X, s.Y) || s.X < 0 || s.Y < 0 || s.X > m.Width || s.Y > m.Height {
				return nil, bad("section %q has seats outside the canvas (%s-%d at %.0f,%.0f): make the canvas bigger or move the section", sec.Key, s.Row, s.Number, s.X, s.Y)
			}
		}
		if len(sec.Outline) > 0 {
			if len(sec.Outline) < 3 || len(sec.Outline) > maxMapPoints {
				return nil, bad("section %q: an outline needs 3 to %d points", sec.Key, maxMapPoints)
			}
			for _, p := range sec.Outline {
				if !finite(p[0], p[1]) || p[0] < 0 || p[1] < 0 || p[0] > m.Width || p[1] > m.Height {
					return nil, bad("section %q has an outline point outside the canvas", sec.Key)
				}
			}
		}
		all = append(all, seats...)
	}
	return all, nil
}

func (st StageShape) validate(w, h float64) error {
	in := func(x, y float64) bool { return finite(x, y) && x >= 0 && y >= 0 && x <= w && y <= h }
	switch st.Shape {
	case "", "rect", "ellipse":
		if st.W < 0 || st.H < 0 || !in(st.X, st.Y) || !in(st.X+st.W, st.Y+st.H) {
			return bad("the stage must lie inside the canvas")
		}
	case "polygon":
		if len(st.Points) < 3 || len(st.Points) > maxMapPoints {
			return bad("a polygon stage needs 3 to %d points", maxMapPoints)
		}
		for _, p := range st.Points {
			if !in(p[0], p[1]) {
				return bad("the stage must lie inside the canvas")
			}
		}
	default:
		return bad("stage shape must be rect, ellipse or polygon")
	}
	if len(st.Label) > 60 {
		return bad("stage label is too long")
	}
	return nil
}

func (s SectionSpec) generate() ([]GeneratedSeat, error) {
	counts, err := s.rowSeats()
	if err != nil && s.Layout == "grid" {
		return nil, err
	}
	skip, err2 := parseSeatKeys(s.Skip)
	if err2 != nil {
		return nil, err2
	}
	acc, err2 := parseSeatKeys(s.Accessible)
	if err2 != nil {
		return nil, err2
	}
	first := 0
	if s.RowStart != "" {
		r := strings.ToUpper(s.RowStart)
		if strings.Trim(r, "ABCDEFGHIJKLMNOPQRSTUVWXYZ") != "" {
			return nil, bad("section %q: row_start must be letters, like A or AA", s.Key)
		}
		first = RowIndex(r)
	}
	if s.NumberFrom != "" && s.NumberFrom != "left" && s.NumberFrom != "right" {
		return nil, bad("section %q: number_from must be left or right", s.Key)
	}
	rev := s.NumberFrom == "right"

	var out []GeneratedSeat
	emit := func(row, k, n int, x, y float64) {
		label := RowLabel(first + row)
		num := k
		if rev {
			num = n - k + 1
		}
		if skip[seatKey(label, num)] {
			return
		}
		out = append(out, GeneratedSeat{Section: s.Key, Row: label, Number: num, X: round2(x), Y: round2(y), Accessible: acc[seatKey(label, num)]})
	}

	switch s.Layout {
	case "grid":
		if len(counts) > maxRowsPerSection {
			return nil, bad("section %q has too many rows (at most %d)", s.Key, maxRowsPerSection)
		}
		pitch, rowPitch := orDefault(s.SeatPitch, defaultSeatPitch), orDefault(s.RowPitch, defaultRowPitch)
		if pitch < 5 || pitch > 200 || rowPitch < 5 || rowPitch > 300 {
			return nil, bad("section %q: seat_pitch and row_pitch must be between 5 and 200", s.Key)
		}
		aisle := map[int]bool{}
		for _, a := range s.AisleAfter {
			aisle[a] = true
		}
		xAt := func(k int) float64 {
			x := float64(k-1) * pitch
			for a := range aisle {
				if a < k {
					x += pitch
				}
			}
			return x
		}
		widest := 0.0
		for _, n := range counts {
			if n < 1 || n > maxSeatsPerRow {
				return nil, bad("section %q: a row has between 1 and %d seats", s.Key, maxSeatsPerRow)
			}
			widest = max(widest, xAt(n))
		}
		sin, cos := math.Sincos(rad(s.Rotation))
		for i, n := range counts {
			off := 0.0
			switch s.Align {
			case "", "center":
				off = (widest - xAt(n)) / 2
			case "left":
			case "right":
				off = widest - xAt(n)
			default:
				return nil, bad("section %q: align must be center, left or right", s.Key)
			}
			for k := 1; k <= n; k++ {
				lx, ly := off+xAt(k), float64(i)*rowPitch
				emit(i, k, n, s.Origin.X+lx*cos-ly*sin, s.Origin.Y+lx*sin+ly*cos)
			}
		}
	case "arc":
		if s.AngleEnd <= s.AngleStart || s.AngleEnd-s.AngleStart > 360 {
			return nil, bad("section %q: angle_end must be greater than angle_start, within 360 degrees", s.Key)
		}
		if s.Radius <= 0 {
			return nil, bad("section %q: an arc needs a radius", s.Key)
		}
		rowPitch := orDefault(s.RowPitch, defaultRowPitch)
		pitch := orDefault(s.SeatPitch, defaultSeatPitch)
		if pitch < 5 || pitch > 200 || rowPitch < 5 || rowPitch > 300 {
			return nil, bad("section %q: seat_pitch and row_pitch must be between 5 and 200", s.Key)
		}
		rows := s.Rows
		if len(s.RowSeats) > 0 {
			rows = len(s.RowSeats)
		}
		if rows < 1 || rows > maxRowsPerSection {
			return nil, bad("section %q: rows must be between 1 and %d", s.Key, maxRowsPerSection)
		}
		span := rad(s.AngleEnd - s.AngleStart)
		for i := 0; i < rows; i++ {
			r := s.Radius + float64(i)*rowPitch
			n := s.SeatsPerRow
			if len(s.RowSeats) > 0 {
				n = s.RowSeats[i]
			}
			if n <= 0 {
				n = int(math.Floor(r * span / pitch))
			}
			if n < 1 || n > maxSeatsPerRow {
				return nil, bad("section %q: row %d would have %d seats (between 1 and %d)", s.Key, i+1, n, maxSeatsPerRow)
			}
			for k := 1; k <= n; k++ {
				a := rad(s.AngleStart) + span*(float64(k)-0.5)/float64(n)
				emit(i, k, n, s.Center.X+r*math.Cos(a), s.Center.Y+r*math.Sin(a))
			}
		}
	default:
		return nil, bad("section %q: layout must be grid or arc", s.Key)
	}
	if len(out) == 0 {
		return nil, bad("section %q has no seats", s.Key)
	}
	return out, nil
}

func orDefault(v, def float64) float64 {
	if v == 0 {
		return def
	}
	return v
}

func round2(v float64) float64 { return math.Round(v*100) / 100 }

func (m VenueMap) View(seats []GeneratedSeat, typeBySection map[string]int64) SeatMapLayout {
	l := SeatMapLayout{Width: m.Width, Height: m.Height}
	for _, st := range m.Stages {
		l.Stages = append(l.Stages, st)
	}
	box := map[string][4]float64{}
	for _, s := range seats {
		b, ok := box[s.Section]
		if !ok {
			b = [4]float64{s.X, s.Y, s.X, s.Y}
		}
		b[0], b[1], b[2], b[3] = min(b[0], s.X), min(b[1], s.Y), max(b[2], s.X), max(b[3], s.Y)
		box[s.Section] = b
	}
	for _, sec := range m.Sections {
		pts := sec.Outline
		if len(pts) == 0 {
			b, ok := box[sec.Key]
			if !ok {
				continue
			}
			pad := orDefault(sec.SeatPitch, defaultSeatPitch) / 2
			x0, y0, x1, y1 := max(0, b[0]-pad), max(0, b[1]-pad), min(m.Width, b[2]+pad), min(m.Height, b[3]+pad)
			pts = [][2]float64{{x0, y0}, {x1, y0}, {x1, y1}, {x0, y1}}
		}
		name := sec.Name
		if name == "" {
			name = sec.Key
		}
		l.Sections = append(l.Sections, SectionShape{Key: sec.Key, Name: name, Label: sec.Label, Color: sec.Color, Points: pts, TicketTypeID: typeBySection[sec.Key]})
	}
	return l
}

// ---- keeping the older single-stage layout working

func (l SeatMapLayout) allStages() []StageShape {
	out := append([]StageShape(nil), l.Stages...)
	if l.Stage != nil {
		out = append(out, StageShape{Label: l.Stage.Label, X: l.Stage.X, Y: l.Stage.Y, W: l.Stage.W, H: l.Stage.H})
	}
	return out
}
