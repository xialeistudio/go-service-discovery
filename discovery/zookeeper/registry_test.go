package zookeeper

import (
	"context"
	json "github.com/json-iterator/go"
	"github.com/stretchr/testify/assert"
	"github.com/xialeistudio/go-service-discovery/discovery"
	"net"
	"testing"
	"time"
)

func TestEtcdRegistry(t *testing.T) {
	a := assert.New(t)

	r, err := NewRegistry([]string{"localhost:2181"})
	a.Nil(err)

	t.Run("Register single node", func(t *testing.T) {
		node := &discovery.ServiceNode{
			ServiceName: "test",
			IP:          net.IPv4(127, 0, 0, 1),
			Port:        8484,
			Tags: map[string]string{
				"version": "1.0",
			},
		}
		// 注册节点
		err = r.Register(context.Background(), node)
		a.Nil(err)
		time.Sleep(time.Millisecond * 100)
		// 获取节点
		nodes, err := r.GetNodes(context.Background(), "test", map[string]string{
			"version": "1.0",
		})
		a.Nil(err)
		a.Len(nodes, 1)
		a.Equal(node, nodes[0])
		// 注销节点
		err = r.Unregister(context.Background(), node)
		a.Nil(err)
		time.Sleep(time.Millisecond * 100)
		// 获取节点
		nodes, err = r.GetNodes(context.Background(), "test", map[string]string{
			"version": "1.0",
		})
		a.Nil(err)
		a.Len(nodes, 0)
	})
	t.Run("Register multi nodeListMap with multi tags", func(t *testing.T) {
		node1 := &discovery.ServiceNode{
			ServiceName: "test",
			IP:          net.IPv4(127, 0, 0, 1),
			Port:        8484,
			Tags: map[string]string{
				"version": "1.0",
				"region":  "cn",
			},
		}
		node2 := &discovery.ServiceNode{
			ServiceName: "test",
			IP:          net.IPv4(127, 0, 0, 2),
			Port:        8484,
			Tags: map[string]string{
				"version": "1.0",
				"region":  "us",
			},
		}
		// 注册节点
		err = r.Register(context.Background(), node1)
		a.Nil(err)
		err = r.Register(context.Background(), node2)
		a.Nil(err)
		time.Sleep(time.Millisecond * 100)
		// 获取节点
		nodes, err := r.GetNodes(context.Background(), "test", map[string]string{
			"version": "1.0",
			"region":  "cn",
		})
		a.Nil(err)
		a.Len(nodes, 1)
		str, _ := json.MarshalToString(nodes)
		t.Log(str)
		a.Equal(node1, nodes[0])
		// 获取节点
		nodes, err = r.GetNodes(context.Background(), "test", map[string]string{
			"version": "1.0",
			"region":  "us",
		})
		a.Nil(err)
		a.Len(nodes, 1)
		a.Equal(node2, nodes[0])
		// 获取节点
		nodes, err = r.GetNodes(context.Background(), "test", nil)
		a.Nil(err)
		a.Len(nodes, 2)
		// 注销节点
		err = r.Unregister(context.Background(), node1)
		a.Nil(err)
		err = r.Unregister(context.Background(), node2)
		a.Nil(err)
		time.Sleep(time.Millisecond * 100)
		// 获取节点
		nodes, err = r.GetNodes(context.Background(), "test", map[string]string{
			"version": "1.0",
		})
		a.Nil(err)
		a.Len(nodes, 0)
	})
}
