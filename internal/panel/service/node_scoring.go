package service

import (
	"sort"

	"github.com/vortexui/vortexui/internal/domain"
)

// NodeScore holds a node's quality score for subscription ordering.
type NodeScore struct {
	Node  *domain.Node
	Score float64
}

// ScoreNodes calculates quality scores for nodes and returns them sorted best-first.
// Score formula: base 100, penalized by CPU/mem/connections/latency, bonused by usage ratio.
func ScoreNodes(nodes []*domain.Node) []NodeScore {
	scores := make([]NodeScore, 0, len(nodes))
	for _, n := range nodes {
		score := computeNodeScore(n)
		scores = append(scores, NodeScore{Node: n, Score: score})
	}
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].Score > scores[j].Score
	})
	return scores
}

func computeNodeScore(n *domain.Node) float64 {
	if !n.IsHealthy() {
		return 0
	}
	// Base: 100 points
	score := 100.0

	// CPU penalty: -0.5 point per %
	score -= n.Health.CPUPercent * 0.5

	// Memory penalty: -0.3 per %
	score -= n.Health.MemPercent * 0.3

	// Connection load penalty
	if n.Health.Connections > 100 {
		score -= float64(n.Health.Connections-100) * 0.1
	}

	// Latency penalty (if ping is available)
	if n.PingMs > 0 {
		score -= float64(n.PingMs) * 0.05 // 200ms ping = -10 points
	}

	// Usage ratio bonus (higher ratio = preferred node)
	score += n.UsageRatio * 10

	if score < 0 {
		score = 0
	}
	return score
}
