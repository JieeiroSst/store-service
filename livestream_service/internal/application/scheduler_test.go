package application

import (
	"context"
	"testing"

	"github.com/JIeeiroSst/livestream-service/config"
	"github.com/JIeeiroSst/livestream-service/internal/domain/model"
)

func TestAssignNodeSkipsOverloadedAndPicksAvailable(t *testing.T) {
	nodes := newFakeNodeRegistry()
	nodes.nodes["busy"] = &model.TranscodeNode{ID: "busy", Capacity: 3, LoadPerCore: 0.95}
	nodes.nodes["free"] = &model.TranscodeNode{ID: "free", Capacity: 5, LoadPerCore: 0.1}
	runner := newFakeTranscodeRunner()
	s := &nodeSchedulerUsecase{assigner: newNodeAssigner(nodes), runner: runner, cfg: &config.Config{}}

	got, err := s.AssignNode(context.Background(), "sk_1")
	if err != nil {
		t.Fatalf("AssignNode() error = %v", err)
	}
	if got.ID != "free" {
		t.Errorf("AssignNode() picked %q, want the non-overloaded node %q", got.ID, "free")
	}
	if nodes.nodes["free"].Capacity != 4 {
		t.Errorf("expected reserved capacity to drop by 1, got %d", nodes.nodes["free"].Capacity)
	}
	assigned, ok, _ := nodes.GetAssignment(context.Background(), "sk_1")
	if !ok || assigned != "free" {
		t.Errorf("expected assignment recorded for sk_1 -> free, got %q (ok=%v)", assigned, ok)
	}
}

func TestAssignNodeReturnsExistingAssignmentIfNodeStillAlive(t *testing.T) {
	nodes := newFakeNodeRegistry()
	nodes.nodes["node-1"] = &model.TranscodeNode{ID: "node-1", Capacity: 5}
	_ = nodes.SetAssignment(context.Background(), "sk_1", "node-1", 0)
	s := &nodeSchedulerUsecase{assigner: newNodeAssigner(nodes), runner: newFakeTranscodeRunner(), cfg: &config.Config{}}

	got, err := s.AssignNode(context.Background(), "sk_1")
	if err != nil {
		t.Fatalf("AssignNode() error = %v", err)
	}
	if got.ID != "node-1" {
		t.Errorf("AssignNode() = %q, want the already-assigned node-1", got.ID)
	}
	// Sticky reassignment must not re-reserve capacity.
	if nodes.nodes["node-1"].Capacity != 5 {
		t.Errorf("expected capacity untouched on sticky reassignment, got %d", nodes.nodes["node-1"].Capacity)
	}
}

func TestAssignNodeFallsBackWhenAssignedNodeIsGone(t *testing.T) {
	nodes := newFakeNodeRegistry()
	_ = nodes.SetAssignment(context.Background(), "sk_1", "dead-node", 0)
	nodes.nodes["alive"] = &model.TranscodeNode{ID: "alive", Capacity: 5}
	s := &nodeSchedulerUsecase{assigner: newNodeAssigner(nodes), runner: newFakeTranscodeRunner(), cfg: &config.Config{}}

	got, err := s.AssignNode(context.Background(), "sk_1")
	if err != nil {
		t.Fatalf("AssignNode() error = %v", err)
	}
	if got.ID != "alive" {
		t.Errorf("AssignNode() = %q, want it to fall back to the alive node", got.ID)
	}
}

func TestAssignNodeErrorsWhenNoCapacityAnywhere(t *testing.T) {
	nodes := newFakeNodeRegistry()
	nodes.nodes["full"] = &model.TranscodeNode{ID: "full", Capacity: 0}
	s := &nodeSchedulerUsecase{assigner: newNodeAssigner(nodes), runner: newFakeTranscodeRunner(), cfg: &config.Config{}}

	if _, err := s.AssignNode(context.Background(), "sk_1"); err != ErrNoNodeAvailable {
		t.Fatalf("AssignNode() error = %v, want ErrNoNodeAvailable", err)
	}
}

func TestReleaseNodeClearsAssignmentAndReturnsCapacity(t *testing.T) {
	nodes := newFakeNodeRegistry()
	nodes.nodes["node-1"] = &model.TranscodeNode{ID: "node-1", Capacity: 4}
	_ = nodes.SetAssignment(context.Background(), "sk_1", "node-1", 0)
	s := &nodeSchedulerUsecase{assigner: newNodeAssigner(nodes), runner: newFakeTranscodeRunner(), cfg: &config.Config{}}

	if err := s.ReleaseNode(context.Background(), "sk_1"); err != nil {
		t.Fatalf("ReleaseNode() error = %v", err)
	}
	if nodes.nodes["node-1"].Capacity != 5 {
		t.Errorf("expected capacity returned to 5, got %d", nodes.nodes["node-1"].Capacity)
	}
	if _, ok, _ := nodes.GetAssignment(context.Background(), "sk_1"); ok {
		t.Error("expected assignment to be cleared")
	}
}

func TestCheckStaleRestartsEveryStaleStreamKey(t *testing.T) {
	runner := newFakeTranscodeRunner()
	runner.stale = []string{"a", "b"}
	s := &nodeSchedulerUsecase{assigner: newNodeAssigner(newFakeNodeRegistry()), runner: runner, cfg: &config.Config{}}

	if err := s.CheckStale(context.Background()); err != nil {
		t.Fatalf("CheckStale() error = %v", err)
	}
	if len(runner.restarted) != 2 {
		t.Errorf("expected both stale stream keys restarted, got %v", runner.restarted)
	}
}
