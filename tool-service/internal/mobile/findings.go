package mobile

const (
	High   = "high"
	Medium = "medium"
	Low    = "low"
	Info   = "info"
)

type Finding struct {
	ID       string `json:"id"`
	Severity string `json:"severity"`
	Title    string `json:"title"`
	Detail   string `json:"detail,omitempty"`
}

var severityRank = map[string]int{High: 0, Medium: 1, Low: 2, Info: 3}

type collector struct{ list []Finding }

func (c *collector) add(id, sev, title, detail string) {
	c.list = append(c.list, Finding{id, sev, title, detail})
}

func (c *collector) sorted() []Finding {
	out := append([]Finding(nil), c.list...)
	for i := 1; i < len(out); i++ {
		for j := i; j > 0 && severityRank[out[j].Severity] < severityRank[out[j-1].Severity]; j-- {
			out[j], out[j-1] = out[j-1], out[j]
		}
	}
	return out
}
