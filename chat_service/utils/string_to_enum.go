package utils

import "fmt"


type Capability string

const (
	Read   Capability = "READ"
	Create Capability = "CREATE"
	Update Capability = "UPDATE"
	Delete Capability = "DELETE"
	List   Capability = "LIST"
)

func (c Capability) String() string {
	return string(c)
}

func ParseCapability(s string) (c Capability, err error) {
	capabilities := map[Capability]struct{}{
		Read:   {},
		Create: {},
		Update: {},
		Delete: {},
		List:   {},
	}

	cap := Capability(s)
	_, ok := capabilities[cap]
	if !ok {
		return c, fmt.Errorf(`cannot parse:[%s] as capability`, s)
	}
	return cap, nil
}
