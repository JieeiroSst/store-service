package common

type StatusInvoices int

const (
	PENDING StatusInvoices = iota + 1
	APPROVE
	REJECT
)

func (s StatusInvoices) Value() int {
	return int(s)
}
