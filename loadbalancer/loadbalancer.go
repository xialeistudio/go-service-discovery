package loadbalancer

import "go-service-discovery/discovery"

type LoadBalancer interface {
	Select(serviceName string, nodes []*discovery.ServiceNode) *discovery.ServiceNode
}
