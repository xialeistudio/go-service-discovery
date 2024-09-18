package loadbalancer

import (
	"github.com/xialeistudio/go-service-discovery/discovery"
)

type roundRobin struct {
	index int
}

func (r *roundRobin) Select(nodes []*discovery.ServiceNode) *discovery.ServiceNode {
	if len(nodes) == 0 {
		return nil
	}
	node := nodes[r.index]
	r.index = (r.index + 1) % len(nodes)
	return node
}

func NewRoundRobin() LoadBalancer {
	return &roundRobin{}
}
