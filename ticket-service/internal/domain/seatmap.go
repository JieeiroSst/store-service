package domain

import (
	"fmt"
	"math"
)

type SeatMapLayout struct {
	Width    float64        `json:"width"`
	Height   float64        `json:"height"`
	Stage    *MapRect       `json:"stage,omitempty"`
	Stages   []StageShape   `json:"stages,omitempty"`
	Sections []SectionShape `json:"sections,omitempty"`
}

type MapRect struct {
	Label string  `json:"label,omitempty"`
	X     float64 `json:"x"`
	Y     float64 `json:"y"`
	W     float64 `json:"w"`
	H     float64 `json:"h"`
}

type SectionShape struct {
	Key          string       `json:"key,omitempty"`
	Name         string       `json:"name"`
	Label        string       `json:"label,omitempty"`
	Color        string       `json:"color,omitempty"`
	Points       [][2]float64 `json:"points"`
	TicketTypeID int64        `json:"ticket_type_id,omitempty"`
}

type SeatPosition struct {
	Section string
	Row     string
	Number  int
	X, Y    float64
}

type SeatMap struct {
	Layout SeatMapLayout
	Seats  []Seat
}

const (
	maxMapSize     = 100_000
	maxMapSections = 100
	maxMapPoints   = 500
)

func finite(v ...float64) bool {
	for _, x := range v {
		if math.IsNaN(x) || math.IsInf(x, 0) {
			return false
		}
	}
	return true
}

func (l SeatMapLayout) Validate() error {
	if !finite(l.Width, l.Height) || l.Width <= 0 || l.Height <= 0 || l.Width > maxMapSize || l.Height > maxMapSize {
		return fmt.Errorf("%w: width and height must be between 0 and %d", ErrInvalid, maxMapSize)
	}
	in := func(x, y float64) bool { return finite(x, y) && x >= 0 && y >= 0 && x <= l.Width && y <= l.Height }
	if s := l.Stage; s != nil && (!in(s.X, s.Y) || !in(s.X+s.W, s.Y+s.H) || s.W < 0 || s.H < 0) {
		return fmt.Errorf("%w: the stage must lie inside the canvas", ErrInvalid)
	}
	if len(l.Stages) > maxVenueStages {
		return fmt.Errorf("%w: at most %d stages", ErrInvalid, maxVenueStages)
	}
	for _, st := range l.Stages {
		if err := st.validate(l.Width, l.Height); err != nil {
			return err
		}
	}
	if len(l.Sections) > maxMapSections {
		return fmt.Errorf("%w: at most %d sections", ErrInvalid, maxMapSections)
	}
	for _, sec := range l.Sections {
		if sec.Name == "" || len(sec.Name) > 60 || len(sec.Label) > 60 || len(sec.Color) > 20 {
			return fmt.Errorf("%w: every section needs a name (up to 60 characters)", ErrInvalid)
		}
		if len(sec.Points) < 3 || len(sec.Points) > maxMapPoints {
			return fmt.Errorf("%w: section %q needs 3 to %d points", ErrInvalid, sec.Name, maxMapPoints)
		}
		for _, p := range sec.Points {
			if !in(p[0], p[1]) {
				return fmt.Errorf("%w: section %q has a point outside the canvas", ErrInvalid, sec.Name)
			}
		}
	}
	return nil
}

type EventSeatMap struct {
	Layout SeatMapLayout
	Types  []TicketType
	Seats  []Seat
}

type SeatGroup struct {
	TicketTypeID int64
	Seats        []GeneratedSeat
}
