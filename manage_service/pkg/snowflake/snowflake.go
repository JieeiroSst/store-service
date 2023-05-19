package snowflake

import (
	"strconv"

	"github.com/JIeeiroSst/manage-service/pkg/log"

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
		log.Error(err)
		return 0
	}
	id, err := strconv.Atoi(n.Generate().String())
	if err != nil {
		log.Error(err)
		return 0
	}
	log.Info(id)
	return id
}
