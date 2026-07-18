package service

import (
	"testing"

	"github.com/google/uuid"
	"github.com/vortexui/vortexui/internal/domain"
)

func TestScoreNodes_HealthyNodes(t *testing.T) {
	nodes := []*domain.Node{
		{
			ID: uuid.New(), Name: "node-a", Status: domain.NodeConnected,
			Health: domain.NodeHealth{CPUPercent: 80, MemPercent: 60, CoreRunning: true, Connections: 50},
			PingMs: 100, UsageRatio: 0.5,
		},
		{
			ID: uuid.New(), Name: "node-b", Status: domain.NodeConnected,
			Health: domain.NodeHealth{CPUPercent: 20, MemPercent: 30, CoreRunning: true, Connections: 10},
			PingMs: 50, UsageRatio: 0.8,
		},
	}

	scores := ScoreNodes(nodes)
	if len(scores) != 2 {
		t.Fatalf("expected 2 scores, got %d", len(scores))
	}
	// node-b should score higher (lower CPU/mem/ping, higher usage ratio)
	if scores[0].Node.Name != "node-b" {
		t.Errorf("expected node-b first (best score), got %s", scores[0].Node.Name)
	}
	if scores[0].Score <= scores[1].Score {
		t.Errorf("first score (%f) should be > second (%f)", scores[0].Score, scores[1].Score)
	}
}

func TestScoreNodes_UnhealthyNodeGetsZero(t *testing.T) {
	nodes := []*domain.Node{
		{
			ID: uuid.New(), Name: "dead-node", Status: domain.NodeDisconnected,
			Health: domain.NodeHealth{CPUPercent: 10, MemPercent: 10, CoreRunning: false},
		},
	}
	scores := ScoreNodes(nodes)
	if scores[0].Score != 0 {
		t.Errorf("unhealthy node should score 0, got %f", scores[0].Score)
	}
}

func TestScoreNodes_EmptyList(t *testing.T) {
	scores := ScoreNodes(nil)
	if len(scores) != 0 {
		t.Errorf("expected empty result for nil input, got %d", len(scores))
	}
}

func TestScoreNodes_HighConnections(t *testing.T) {
	nodes := []*domain.Node{
		{
			ID: uuid.New(), Name: "loaded", Status: domain.NodeConnected,
			Health: domain.NodeHealth{CPUPercent: 10, MemPercent: 10, CoreRunning: true, Connections: 500},
		},
	}
	scores := ScoreNodes(nodes)
	// 500-100 = 400 excess connections * 0.1 = 40 penalty
	// base 100 - 5 (cpu) - 3 (mem) - 40 (conn) = 52
	expected := 52.0
	if scores[0].Score != expected {
		t.Errorf("expected score %f, got %f", expected, scores[0].Score)
	}
}

func TestScoreNodes_SortOrder(t *testing.T) {
	nodes := []*domain.Node{
		{
			ID: uuid.New(), Name: "worst", Status: domain.NodeConnected,
			Health: domain.NodeHealth{CPUPercent: 90, MemPercent: 90, CoreRunning: true},
		},
		{
			ID: uuid.New(), Name: "best", Status: domain.NodeConnected,
			Health: domain.NodeHealth{CPUPercent: 5, MemPercent: 5, CoreRunning: true},
		},
		{
			ID: uuid.New(), Name: "mid", Status: domain.NodeConnected,
			Health: domain.NodeHealth{CPUPercent: 50, MemPercent: 50, CoreRunning: true},
		},
	}
	scores := ScoreNodes(nodes)
	if scores[0].Node.Name != "best" {
		t.Errorf("expected best first, got %s", scores[0].Node.Name)
	}
	if scores[2].Node.Name != "worst" {
		t.Errorf("expected worst last, got %s", scores[2].Node.Name)
	}
}
