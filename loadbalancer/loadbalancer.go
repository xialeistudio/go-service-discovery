package loadbalancer

import "github.com/xialeistudio/go-service-discovery/discovery"

type LoadBalancer interface {
	Select(nodes []*discovery.ServiceNode) *discovery.ServiceNode
}
