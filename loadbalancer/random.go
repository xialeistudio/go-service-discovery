package loadbalancer

import (
	"github.com/xialeistudio/go-service-discovery/discovery"
	"math/rand"
	"time"
)

type random struct {
	r *rand.Rand
}

func (r random) Select(nodes []*discovery.ServiceNode) *discovery.ServiceNode {
	if len(nodes) == 0 {
		return nil
	}
	index := r.r.Intn(len(nodes))
	return nodes[index]
}

func NewRandom() LoadBalancer {
	return &random{
		r: rand.New(rand.NewSource(time.Now().Unix())),
	}
}
