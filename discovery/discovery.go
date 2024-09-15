package discovery

import (
	"context"
	"net"
)

// ServiceNode 服务节点
type ServiceNode struct {
	ServiceName string
	IP          net.IP
	Port        int
	Tags        map[string]string
}

// NodeRegistry 服务节点注册中心
type NodeRegistry interface {
	// GetNodes 获取服务节点
	GetNodes(ctx context.Context, serviceName string, tags map[string]string) ([]*ServiceNode, error)
	// Register 注册服务节点
	Register(ctx context.Context, node *ServiceNode) error
	// Unregister 注销服务节点
	Unregister(ctx context.Context, node *ServiceNode) error
}

// MatchTags 匹配标签
func MatchTags(nodeTags, queryTags map[string]string) bool {
	for name, value := range queryTags {
		if nodeTags[name] != value {
			return false
		}
	}
	return true
}
