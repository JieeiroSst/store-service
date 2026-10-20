package model

type Actor struct {
	UserID  string
	Admin   bool
	Service bool
}

func (a Actor) IsPlayer() bool { return a.UserID != "" && !a.Service }
