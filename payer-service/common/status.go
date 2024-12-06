package common

type StatusType int

const (
	Peding StatusType = iota + 1
	Susscess
	Reject
)

func (s StatusType) Value() int {
	return int(s)
}
