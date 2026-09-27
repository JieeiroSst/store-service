package apitest

import (
	"errors"
	"fmt"
	"math/big"
	"strings"
)

func EvalCalc(expr string) (*big.Rat, error) {
	p := &calcParser{s: strings.TrimSpace(expr)}
	if p.s == "" {
		return nil, errors.New("empty expression")
	}
	v, err := p.expr()
	if err != nil {
		return nil, err
	}
	p.skip()
	if p.i != len(p.s) {
		return nil, fmt.Errorf("unexpected %q at position %d", p.s[p.i], p.i)
	}
	return v, nil
}

type calcParser struct {
	s string
	i int
}

func (p *calcParser) skip() {
	for p.i < len(p.s) && p.s[p.i] == ' ' {
		p.i++
	}
}

func (p *calcParser) peek() byte {
	p.skip()
	if p.i < len(p.s) {
		return p.s[p.i]
	}
	return 0
}

func (p *calcParser) expr() (*big.Rat, error) {
	v, err := p.term()
	if err != nil {
		return nil, err
	}
	for p.peek() == '+' || p.peek() == '-' {
		op := p.peek()
		p.i++
		r, err := p.term()
		if err != nil {
			return nil, err
		}
		if op == '+' {
			v = new(big.Rat).Add(v, r)
		} else {
			v = new(big.Rat).Sub(v, r)
		}
	}
	return v, nil
}

func (p *calcParser) term() (*big.Rat, error) {
	v, err := p.factor()
	if err != nil {
		return nil, err
	}
	for p.peek() == '*' || p.peek() == '/' {
		op := p.peek()
		p.i++
		r, err := p.factor()
		if err != nil {
			return nil, err
		}
		if op == '*' {
			v = new(big.Rat).Mul(v, r)
		} else {
			if r.Sign() == 0 {
				return nil, errors.New("division by zero")
			}
			v = new(big.Rat).Quo(v, r)
		}
	}
	return v, nil
}

func (p *calcParser) factor() (*big.Rat, error) {
	switch c := p.peek(); {
	case c == '-':
		p.i++
		v, err := p.factor()
		if err != nil {
			return nil, err
		}
		return new(big.Rat).Neg(v), nil
	case c == '(':
		p.i++
		v, err := p.expr()
		if err != nil {
			return nil, err
		}
		if p.peek() != ')' {
			return nil, errors.New("missing )")
		}
		p.i++
		return v, nil
	case c >= '0' && c <= '9' || c == '.':
		start := p.i
		for p.i < len(p.s) && (p.s[p.i] >= '0' && p.s[p.i] <= '9' || p.s[p.i] == '.') {
			p.i++
		}
		v, ok := new(big.Rat).SetString(p.s[start:p.i])
		if !ok {
			return nil, fmt.Errorf("bad number %q", p.s[start:p.i])
		}
		return v, nil
	case c == 0:
		return nil, errors.New("unexpected end of expression")
	default:
		return nil, fmt.Errorf("unexpected %q at position %d", c, p.i)
	}
}

func numeric(v any) (*big.Rat, bool) {
	switch t := v.(type) {
	case interface{ String() string }:
		r, ok := new(big.Rat).SetString(t.String())
		return r, ok
	case string:
		r, ok := new(big.Rat).SetString(strings.TrimSpace(t))
		return r, ok
	case float64:
		r := new(big.Rat)
		return r.SetFloat64(t), r != nil
	case int:
		return new(big.Rat).SetInt64(int64(t)), true
	}
	return nil, false
}

func ResolveExpected(want any) (any, error) {
	s, ok := want.(string)
	if !ok || !strings.HasPrefix(strings.ToLower(strings.TrimSpace(s)), "calc:") {
		return want, nil
	}
	r, err := EvalCalc(strings.TrimSpace(s)[5:])
	if err != nil {
		return nil, err
	}
	return r, nil
}

func equalValues(got, want any) bool {
	if w, ok := want.(*big.Rat); ok {
		g, ok := numeric(got)
		return ok && g.Cmp(w) == 0
	}
	if g, gok := numeric(got); gok {
		if w, wok := numeric(want); wok {
			return g.Cmp(w) == 0
		}
	}
	return fmt.Sprint(got) == fmt.Sprint(want)
}

func showValue(v any) string {
	if r, ok := v.(*big.Rat); ok {
		if r.IsInt() {
			return r.Num().String()
		}
		return r.FloatString(6)
	}
	return fmt.Sprint(v)
}
