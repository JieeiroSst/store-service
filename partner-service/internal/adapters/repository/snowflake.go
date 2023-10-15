package repository

import (
	"fmt"

	"github.com/bwmarrin/snowflake"
)

func snowflakeID() string {
	node, err := snowflake.NewNode(1)
	if err != nil {
		fmt.Println(err)
	}
	return node.Generate().String()
}
