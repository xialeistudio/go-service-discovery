package loadbalancer

import (
	"github.com/spf13/cast"
	"github.com/xialeistudio/go-service-discovery/discovery"
	"sync"
)

type weightedRoundRobin struct {
	mu            sync.Mutex
	index         int
	currentWeight int
	totalWeight   int
	nodes         []*discovery.ServiceNode
}

func (w *weightedRoundRobin) Select(nodes []*discovery.ServiceNode) *discovery.ServiceNode {
	w.mu.Lock()
	defer w.mu.Unlock()

	if len(nodes) == 0 {
		return nil
	}

	// 节点变动时重新计算权重
	if !equalNodes(w.nodes, nodes) {
		w.nodes = nodes
		w.totalWeight = 0
		for _, node := range nodes {
			w.totalWeight += cast.ToInt(node.Tags["weight"])
		}
	}

	for {
		w.index = (w.index + 1) % len(nodes)
		if w.index == 0 {
			w.currentWeight -= gcd(nodes)
			if w.currentWeight <= 0 {
				w.currentWeight = w.totalWeight
				if w.currentWeight == 0 {
					return nil
				}
			}
		}

		if cast.ToInt(nodes[w.index].Tags["weight"]) >= w.currentWeight {
			return nodes[w.index]
		}
	}
}

func equalNodes(a, b []*discovery.ServiceNode) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func gcd(nodes []*discovery.ServiceNode) int {
	gcdValue := cast.ToInt(nodes[0].Tags["weight"])
	for _, node := range nodes {
		gcdValue = gcdTwo(gcdValue, cast.ToInt(node.Tags["weight"]))
	}
	return gcdValue
}

func gcdTwo(a, b int) int {
	if b == 0 {
		return a
	}
	return gcdTwo(b, a%b)
}

func NewWeightedRoundRobin() LoadBalancer {
	return &weightedRoundRobin{
		index:         -1,
		currentWeight: 0,
		totalWeight:   0,
	}
}
