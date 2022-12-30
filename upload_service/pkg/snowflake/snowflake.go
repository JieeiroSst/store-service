package snowflake

import (
	"fmt"
	"strconv"

	"github.com/JIeeiroSst/upload-service/pkg/log"
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
		log.Error("Genarate id failed")
		return 0
	}
	id, err := strconv.Atoi(n.Generate().String())
	if err != nil {
		log.Error("Genarate id failed")
		return 0
	}
	log.Error(fmt.Sprintf("Genarate id success %v", id))
	return id
}