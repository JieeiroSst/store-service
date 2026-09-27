package dbcheck

import (
	"math/big"
	"strings"
)

type rat struct{ r *big.Rat }

func (x *rat) parse(s string) (*rat, bool) {
	r, ok := new(big.Rat).SetString(strings.TrimSpace(s))
	if !ok {
		return nil, false
	}
	return &rat{r}, true
}

func (x *rat) cmp(y *rat) bool { return x.r.Cmp(y.r) == 0 }
