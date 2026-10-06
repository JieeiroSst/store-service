package domain

import (
	"sort"
	"strings"
)

type Engine struct {
	ID           string     `json:"id"`
	Group        string     `json:"group"`
	Name         string     `json:"name"`
	Required     [][]string `json:"required,omitempty"`
	Discontinued string     `json:"discontinued,omitempty"`
}

func (e Engine) Validate(p Params) error {
	if e.Discontinued != "" {
		return Invalid("engine %s is discontinued by SerpApi: %s", e.ID, e.Discontinued)
	}
	var missing []string
	for _, alts := range e.Required {
		found := false
		for _, name := range alts {
			if strings.TrimSpace(p[name]) != "" {
				found = true
				break
			}
		}
		if !found {
			missing = append(missing, strings.Join(alts, " or "))
		}
	}
	if len(missing) > 0 {
		return Invalid("engine %s requires parameter(s): %s", e.ID, strings.Join(missing, ", "))
	}
	return nil
}

type Catalog struct {
	list []Engine
	byID map[string]Engine
}

func NewCatalog(engines []Engine) *Catalog {
	c := &Catalog{byID: make(map[string]Engine, len(engines))}
	for _, e := range engines {
		if _, dup := c.byID[e.ID]; dup {
			continue
		}
		c.byID[e.ID] = e
		c.list = append(c.list, e)
	}
	sort.SliceStable(c.list, func(i, j int) bool {
		if c.list[i].Group != c.list[j].Group {
			return c.list[i].Group < c.list[j].Group
		}
		return c.list[i].ID < c.list[j].ID
	})
	return c
}

func (c *Catalog) Get(id string) (Engine, bool) {
	e, ok := c.byID[id]
	return e, ok
}

func (c *Catalog) List(group string) []Engine {
	if group == "" {
		return append([]Engine(nil), c.list...)
	}
	var out []Engine
	for _, e := range c.list {
		if strings.EqualFold(e.Group, group) {
			out = append(out, e)
		}
	}
	return out
}

func ValidEngineID(id string) bool {
	if id == "" || len(id) > 64 {
		return false
	}
	for _, r := range id {
		if !(r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '_') {
			return false
		}
	}
	return true
}
