package client

import (
	"context"
	"go-service-discovery/discovery"
	"go-service-discovery/loadbalancer"
)

// Client 服务发现客户端
type Client struct {
	LoadBalancer loadbalancer.LoadBalancer
	Registry     discovery.NodeRegistry
}

// New 创建服务发现客户端
func New(loadBalancer loadbalancer.LoadBalancer, registry discovery.NodeRegistry) *Client {
	return &Client{LoadBalancer: loadBalancer, Registry: registry}
}

// Resolve 获取服务节点
func (c *Client) Resolve(ctx context.Context, serviceName string) (*discovery.ServiceNode, error) {
	nodes, err := c.Registry.GetNodes(ctx, serviceName, nil)
	if err != nil {
		return nil, err
	}

	node := c.LoadBalancer.Select(serviceName, nodes)
	return node, nil
}
