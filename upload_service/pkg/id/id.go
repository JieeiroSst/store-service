package id

import (
	"github.com/google/uuid"
)

type uuidData struct{}

type UUIDData interface {
	GearedID() string
}

func NewSnowflake() UUIDData {
	return &uuidData{}
}

func (u *uuidData) GearedID() string {
	id := uuid.New()

	return id.String()
}
