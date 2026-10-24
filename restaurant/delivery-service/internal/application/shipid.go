package application

import (
	"sync"

	"github.com/bwmarrin/snowflake"
)

var (
	shipIDOnce sync.Once
	shipIDNode *snowflake.Node
)

func newShipID() int {
	shipIDOnce.Do(func() {
		shipIDNode, _ = snowflake.NewNode(1)
	})
	if shipIDNode == nil {
		return 0
	}
	return int(shipIDNode.Generate().Int64())
}
