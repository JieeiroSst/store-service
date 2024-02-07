package snowflake

import (
	"github.com/bwmarrin/snowflake"
)

func GearedID() string {
	n, err := snowflake.NewNode(1)
	if err != nil {
		return ""
	}

	return n.Generate().Base36()
}
