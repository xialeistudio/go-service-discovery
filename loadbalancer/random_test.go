package loadbalancer

import (
	"github.com/stretchr/testify/assert"
	"github.com/xialeistudio/go-service-discovery/discovery"
	"math/rand"
	"net"
	"testing"
)

func Test_random_Select(t *testing.T) {
	a := assert.New(t)
	r := rand.New(rand.NewSource(1))
	lb := random{r: r}
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

	node := lb.Select(nodes)
	a.Equal(nodes[2], node)

	node = lb.Select(nodes)
	a.Equal(nodes[0], node)

	node = lb.Select(nodes)
	a.Equal(nodes[2], node)
}
