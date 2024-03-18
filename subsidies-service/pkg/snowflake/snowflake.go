package snowflake

import (
	"strconv"

	"github.com/bwmarrin/snowflake"
)

func GearedID() int {
	n, err := snowflake.NewNode(1)
	if err != nil {

		return 0
	}
	id, err := strconv.Atoi(n.Generate().String())
	if err != nil {

		return 0
	}

	return id
}
