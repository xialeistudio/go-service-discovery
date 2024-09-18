package loadbalancer

import (
	"github.com/stretchr/testify/assert"
	"github.com/xialeistudio/go-service-discovery/discovery"
	"net"
	"testing"
)

func Test_roundRobin_Select(t *testing.T) {
	a := assert.New(t)
	nodes := []*discovery.ServiceNode{
		{
			ServiceName: "test",
			IP:          net.IPv4(127, 0, 0, 1),
			Port:        8484,
			Tags:        map[string]string{},
		},
		{
			ServiceName: "test",
			IP:          net.IPv4(127, 0, 0, 2),
			Port:        8484,
			Tags:        map[string]string{},
		},
		{
			ServiceName: "test",
			IP:          net.IPv4(127, 0, 0, 3),
			Port:        8484,
			Tags:        map[string]string{},
		},
	}

	lb := NewRoundRobin()
	node := lb.Select(nodes)
	a.Equal(nodes[0], node)

	node = lb.Select(nodes)
	a.Equal(nodes[1], node)
}
