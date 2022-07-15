package snowflake

import (
	"strconv"

	"github.com/bwmarrin/snowflake"
)

type snowflakeData struct{}

type SnowflakeData interface {
	GearedID() int
}

func NewSnowflake() SnowflakeData {
	return &snowflakeData{}
}

func (s *snowflakeData) GearedID() int {
	n, err := snowflake.NewNode(1)
	if err != nil {
		return 0
	}
	id ,err := strconv.Atoi(n.Generate().String())
	if err != nil  {
		return 0
	}
	return id
}
