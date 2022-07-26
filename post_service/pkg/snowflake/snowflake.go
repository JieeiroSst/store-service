package snowflake

import (
	"github.com/bwmarrin/snowflake"
)

type snowflakeData struct {}

type SnowflakeData interface {
	GearedID() string
}

func NewSnowflake() SnowflakeData {
	return &snowflakeData{}
}

func (s *snowflakeData) GearedID() string {
	n, err := snowflake.NewNode(1)
	if err != nil {
		return ""
	}
	return n.Generate().String()
}