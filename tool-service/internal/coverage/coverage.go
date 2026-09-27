package coverage

import (
	"regexp"
	"strings"
	"unicode"
)

type Report struct {
	Total     int      `json:"total_requirements"`
	Covered   int      `json:"covered"`
	Percent   float64  `json:"percent"`
	Uncovered []string `json:"uncovered,omitempty"`
}

var markers = []string{
	" must ", " shall ", " should ", " required", " cannot ", " can't ", " only ", " at least ", " at most ", " maximum ", " minimum ",
	" within ", " no more than ", " never ", " always ", " returns ", " return ", " rejects ", " reject ", " allowed", " not allowed",
	" phải ", " không được ", " cần ", " tối đa ", " tối thiểu ", " yêu cầu ", " bắt buộc ", " chỉ được ", " trả về ", " từ chối ", " không cho phép ",
}

var splitter = regexp.MustCompile(`[\n\r]+|[.!?;]\s+`)

func Extract(doc string) []string {
	var out []string
	seen := map[string]bool{}
	for _, part := range splitter.Split(doc, -1) {
		s := strings.TrimSpace(strings.TrimLeft(strings.TrimSpace(part), "-*•0123456789.) \t"))
		if len(s) < 12 || len(s) > 400 {
			continue
		}
		low := " " + strings.ToLower(s) + " "
		hit := false
		for _, m := range markers {
			if strings.Contains(low, m) {
				hit = true
				break
			}
		}
		if !hit {
			continue
		}
		key := norm(s)
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, s)
		if len(out) >= 200 {
			break
		}
	}
	return out
}

func Compute(reqs, quotes []string) Report {
	r := Report{Total: len(reqs)}
	for _, req := range reqs {
		covered := false
		for _, q := range quotes {
			if matches(req, q) {
				covered = true
				break
			}
		}
		if covered {
			r.Covered++
		} else {
			r.Uncovered = append(r.Uncovered, req)
		}
	}
	if r.Total > 0 {
		r.Percent = float64(int(1000*float64(r.Covered)/float64(r.Total)+0.5)) / 10
	}
	return r
}

func matches(req, quote string) bool {
	a, b := norm(req), norm(quote)
	if len(b) < 8 || len(a) < 8 {
		return false
	}
	if strings.Contains(a, b) || strings.Contains(b, a) {
		return true
	}
	wa, wb := strings.Fields(a), strings.Fields(b)
	short, long := wa, map[string]bool{}
	for _, w := range wb {
		long[w] = true
	}
	if len(wb) < len(wa) {
		short = wb
		long = map[string]bool{}
		for _, w := range wa {
			long[w] = true
		}
	}
	hit := 0
	for _, w := range short {
		if long[w] {
			hit++
		}
	}
	return len(short) >= 4 && float64(hit)/float64(len(short)) >= 0.8
}

func norm(s string) string {
	var b strings.Builder
	space := true
	for _, r := range strings.ToLower(s) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			space = false
		} else if !space {
			b.WriteByte(' ')
			space = true
		}
	}
	return strings.TrimSpace(b.String())
}
