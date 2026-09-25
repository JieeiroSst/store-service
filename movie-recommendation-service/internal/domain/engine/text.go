package engine

import (
	"math"
	"strings"
	"unicode"
)

type vector map[string]float64

var stopwords = map[string]struct{}{}

func init() {
	for _, w := range strings.Fields(`a an and are as at be but by for from has have in is it its of on or that the
		this to was were will with you your i we they he she not no so if then than about into over up out how what
		when where who why video videos official full new`) {
		stopwords[w] = struct{}{}
	}
}

func tokenize(s string) []string {
	fields := strings.FieldsFunc(strings.ToLower(s), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
	out := fields[:0]
	for _, f := range fields {
		if _, stop := stopwords[f]; stop || len([]rune(f)) < 2 {
			continue
		}
		out = append(out, f)
	}
	return out
}

func (v vector) normalize() {
	var sum float64
	for _, x := range v {
		sum += x * x
	}
	if sum == 0 {
		return
	}
	n := math.Sqrt(sum)
	for k := range v {
		v[k] /= n
	}
}

func dot(a, b vector) float64 {
	if len(a) > len(b) {
		a, b = b, a
	}
	var s float64
	for k, x := range a {
		s += x * b[k]
	}
	return s
}
