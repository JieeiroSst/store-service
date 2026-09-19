package idgen

import (
	"fmt"

	"github.com/JIeeiroSst/post-service/internal/domain/port"
	"github.com/bwmarrin/snowflake"
)

type snowflakeGenerator struct {
	node *snowflake.Node
}

func NewIDGenerator() (port.IDGenerator, error) {
	node, err := snowflake.NewNode(1)
	if err != nil {
		return nil, fmt.Errorf("create snowflake node: %w", err)
	}
	return &snowflakeGenerator{node: node}, nil
}

func (g *snowflakeGenerator) NewID() string {
	return g.node.Generate().String()
}
