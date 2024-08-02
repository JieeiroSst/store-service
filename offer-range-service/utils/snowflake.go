package utils

import (
	"github.com/bwmarrin/snowflake"
)

func GearedID() int {
	n, err := snowflake.NewNode(1)
	if err != nil {
		return 0
	}
	return int(n.Generate().Int64())
}
