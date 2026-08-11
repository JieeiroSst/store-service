package application

import (
	"context"
	"fmt"
	"time"

	"github.com/JIeeiroSst/livestream-service/internal/domain/model"
	"github.com/JIeeiroSst/livestream-service/internal/domain/port"
)

const maxCandidates = 5
const overloadedThreshold = 0.9
const assignmentTTL = 24 * time.Hour

type nodeAssigner struct {
	nodes port.NodeRegistry
}

func newNodeAssigner(nodes port.NodeRegistry) *nodeAssigner {
	return &nodeAssigner{nodes: nodes}
}

func (a *nodeAssigner) assign(ctx context.Context, streamKey string) (*model.TranscodeNode, error) {
	if assigned, ok, err := a.nodes.GetAssignment(ctx, streamKey); err != nil {
		return nil, fmt.Errorf("get assignment: %w", err)
	} else if ok {
		if node, found, err := a.nodes.GetNode(ctx, assigned); err != nil {
			return nil, fmt.Errorf("get node: %w", err)
		} else if found {
			return node, nil
		}
		_ = a.nodes.ClearAssignment(ctx, streamKey)
	}

	candidates, err := a.nodes.TopCandidates(ctx, maxCandidates)
	if err != nil {
		return nil, fmt.Errorf("top candidates: %w", err)
	}

	for _, nodeID := range candidates {
		node, found, err := a.nodes.GetNode(ctx, nodeID)
		if err != nil || !found || !node.HasCapacity() || node.LoadPerCore > overloadedThreshold {
			continue
		}

		reserved, err := a.nodes.ReserveCapacity(ctx, nodeID)
		if err != nil || !reserved {
			continue
		}
		if err := a.nodes.SetAssignment(ctx, streamKey, nodeID, assignmentTTL); err != nil {
			_ = a.nodes.ReleaseCapacity(ctx, nodeID)
			continue
		}
		return node, nil
	}

	return nil, ErrNoNodeAvailable
}

func (a *nodeAssigner) claim(ctx context.Context, streamKey, nodeID string) error {
	assigned, ok, err := a.nodes.GetAssignment(ctx, streamKey)
	if err != nil {
		return fmt.Errorf("get assignment: %w", err)
	}
	if ok {
		if assigned != nodeID {
			return fmt.Errorf("stream key assigned to node %s, not %s", assigned, nodeID)
		}
		return nil
	}

	reserved, err := a.nodes.ReserveCapacity(ctx, nodeID)
	if err != nil {
		return fmt.Errorf("reserve capacity: %w", err)
	}
	if !reserved {
		return ErrNodeAtCapacity
	}
	return a.nodes.SetAssignment(ctx, streamKey, nodeID, assignmentTTL)
}

func (a *nodeAssigner) release(ctx context.Context, streamKey string) error {
	assigned, ok, err := a.nodes.GetAssignment(ctx, streamKey)
	if err != nil {
		return fmt.Errorf("get assignment: %w", err)
	}
	if ok {
		if err := a.nodes.ReleaseCapacity(ctx, assigned); err != nil {
			return fmt.Errorf("release capacity: %w", err)
		}
	}
	return a.nodes.ClearAssignment(ctx, streamKey)
}
