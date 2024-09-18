package loadbalancer

import (
	"github.com/stretchr/testify/assert"
	"github.com/xialeistudio/go-service-discovery/discovery"
	"net"
	"testing"
)

func Test_weightedRoundRobin_Select(t *testing.T) {
	a := assert.New(t)
	nodes := []*discovery.ServiceNode{
		{
			ServiceName: "test",
			IP:          net.IPv4(127, 0, 0, 1),
			Port:        8484,
			Tags: map[string]string{
				"weight": "90",
			},
		},
		{
			ServiceName: "test",
			IP:          net.IPv4(127, 0, 0, 2),
			Port:        8484,
			Tags: map[string]string{
				"weight": "50",
			},
		},
		{
			ServiceName: "test",
			IP:          net.IPv4(127, 0, 0, 3),
			Port:        8484,
			Tags: map[string]string{
				"weight": "80",
			},
		},
	}

	lb := NewWeightedRoundRobin()
	node := lb.Select(nodes)
	a.Equal(nodes[0], node)

	node = lb.Select(nodes)
	a.Equal(nodes[0], node)

	node = lb.Select(nodes)
	a.Equal(nodes[2], node)
}
