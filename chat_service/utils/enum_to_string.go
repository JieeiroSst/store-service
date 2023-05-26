package utils

import "fmt"

type MyEnum int

const (
	Foo MyEnum = 1
	Bar MyEnum = 2
)

func (e MyEnum) String() string {
	switch e {
	case Foo:
		return "Foo"
	case Bar:
		return "Bar"
	default:
		return fmt.Sprintf("%d", int(e))
	}
}
