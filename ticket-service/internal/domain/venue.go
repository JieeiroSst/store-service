package domain

import (
	"fmt"
	"math"
	"time"
)

type Venue struct {
	ID          int64
	OwnerID     int64
	Name        string
	City        string
	Address     string
	Description string
	Shared      bool
	Map         VenueMap
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Seats       int
	Sections    int
}

type VenueTemplate struct {
	Name        string
	Description string
	Params      []TemplateParam
	build       func(p map[string]int) VenueMap
}

type TemplateParam struct {
	Name              string
	Description       string
	Default, Min, Max int
}

var venueTemplates = []VenueTemplate{
	{
		Name: "theatre", Description: "Fan-shaped theatre: rows curve around a stage at the front, each row a little longer than the one before; an optional balcony behind.",
		Params: []TemplateParam{
			{"rows", "rows in the stalls", 12, 1, 40}, {"first_row_seats", "seats in the first row", 14, 2, 60},
			{"growth", "seats added per row", 2, 0, 6}, {"balcony_rows", "rows in the balcony (0: none)", 0, 0, 20},
		},
		build: func(p map[string]int) VenueMap {
			const r0, rowPitch, margin = 240.0, 34.0, 40.0
			rows := p["rows"]
			counts := make([]int, rows)
			for i := range counts {
				counts[i] = p["first_row_seats"] + p["growth"]*i
			}
			rEnd := r0 + float64(rows)*rowPitch
			if p["balcony_rows"] > 0 {
				rEnd += 50 + float64(p["balcony_rows"])*rowPitch
			}
			half := rEnd * math.Cos(rad(35))
			w := 2*half + 2*margin
			cy := 70.0
			m := VenueMap{Width: math.Ceil(w), Height: math.Ceil(cy + rEnd + margin),
				Stages: []StageShape{{Label: "STAGE", X: w/2 - 110, Y: 10, W: 220, H: 45}}}
			m.Sections = append(m.Sections, SectionSpec{Key: "ORCH", Name: "Stalls", Layout: "arc", Center: Point{w / 2, cy},
				Radius: r0, RowSeats: counts, AngleStart: 35, AngleEnd: 145})
			if b := p["balcony_rows"]; b > 0 {
				m.Sections = append(m.Sections, SectionSpec{Key: "BALC", Name: "Balcony", Layout: "arc", Center: Point{w / 2, cy},
					Radius: r0 + float64(rows)*rowPitch + 50, Rows: b, AngleStart: 35, AngleEnd: 145, RowStart: RowLabel(0)})
			}
			return m
		},
	},
	{
		Name: "hall", Description: "Conference or concert hall: straight rows facing a stage at one end, with an optional centre aisle.",
		Params: []TemplateParam{
			{"rows", "rows", 10, 1, 100}, {"seats_per_row", "seats in each row", 20, 1, 100}, {"aisle_after", "centre aisle after this seat (0: none)", 10, 0, 100},
		},
		build: func(p map[string]int) VenueMap {
			const pitch, rowPitch, margin = 30.0, 34.0, 40.0
			n := p["seats_per_row"]
			width := float64(n-1) * pitch
			if p["aisle_after"] > 0 && p["aisle_after"] < n {
				width += pitch
			}
			w := width + 2*margin
			sec := SectionSpec{Key: "MAIN", Name: "Main floor", Layout: "grid", Origin: Point{margin, 110}, Rows: p["rows"], SeatsPerRow: n}
			if p["aisle_after"] > 0 && p["aisle_after"] < n {
				sec.AisleAfter = []int{p["aisle_after"]}
			}
			return VenueMap{Width: math.Ceil(w), Height: math.Ceil(110 + float64(p["rows"]-1)*rowPitch + margin),
				Stages:   []StageShape{{Label: "STAGE", X: margin, Y: 20, W: width, H: 50}},
				Sections: []SectionSpec{sec}}
		},
	},
	{
		Name: "arena", Description: "Arena with a stage in the middle: four blocks (N, S, E, W) face it from every side.",
		Params: []TemplateParam{{"rows", "rows in each block", 8, 1, 50}, {"seats_per_row", "seats in each row", 14, 1, 60}},
		build: func(p map[string]int) VenueMap {
			const pitch, rowPitch, gap, margin = 30.0, 34.0, 30.0, 40.0
			bw, bd := float64(p["seats_per_row"]-1)*pitch, float64(p["rows"]-1)*rowPitch
			sw, sh := math.Max(bw*0.7, 120), math.Max(bw*0.45, 80)
			hx := math.Max(sw/2, bw/2)
			halfX := hx + gap + bd
			halfY := math.Max(sh/2+gap+bd, bw/2)
			w, h := 2*halfX+2*margin, 2*halfY+2*margin
			cx, cy := w/2, h/2
			block := func(key, name string, ox, oy, rot float64) SectionSpec {
				return SectionSpec{Key: key, Name: name, Layout: "grid", Origin: Point{ox, oy}, Rows: p["rows"], SeatsPerRow: p["seats_per_row"], Rotation: rot}
			}
			return VenueMap{Width: math.Ceil(w), Height: math.Ceil(h),
				Stages: []StageShape{{Label: "STAGE", Shape: "ellipse", X: cx - sw/2, Y: cy - sh/2, W: sw, H: sh}},
				Sections: []SectionSpec{
					block("S", "South", cx-bw/2, cy+sh/2+gap, 0),
					block("N", "North", cx+bw/2, cy-sh/2-gap, 180),
					block("E", "East", cx+hx+gap, cy+bw/2, -90),
					block("W", "West", cx-hx-gap, cy-bw/2, 90),
				}}
		},
	},
}

func VenueTemplates() []VenueTemplate { return venueTemplates }

func BuildVenueTemplate(name string, params map[string]int) (VenueMap, error) {
	for _, t := range venueTemplates {
		if t.Name != name {
			continue
		}
		p := map[string]int{}
		for _, pi := range t.Params {
			p[pi.Name] = pi.Default
		}
		for k, v := range params {
			var info *TemplateParam
			for i := range t.Params {
				if t.Params[i].Name == k {
					info = &t.Params[i]
				}
			}
			if info == nil {
				return VenueMap{}, fmt.Errorf("%w: template %q has no parameter %q", ErrInvalid, name, k)
			}
			if v < info.Min || v > info.Max {
				return VenueMap{}, fmt.Errorf("%w: %s must be between %d and %d", ErrInvalid, k, info.Min, info.Max)
			}
			p[k] = v
		}
		return t.build(p), nil
	}
	return VenueMap{}, fmt.Errorf("%w: unknown template %q", ErrInvalid, name)
}
