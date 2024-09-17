package zookeeper

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-zookeeper/zk"
	json "github.com/json-iterator/go"
	"github.com/xialeistudio/go-service-discovery/discovery"
	"sync"
	"time"
)

var (
	// DialTimeout 默认的连接超时时间
	DialTimeout = 5 * time.Second
	// BasePath 服务注册的根路径
	BasePath = "/services"
)

type registry struct {
	nodeListMap *sync.Map // <serviceName, <nodeKey, *ServiceNode>>
	conn        *zk.Conn
}

func (r *registry) GetNodes(ctx context.Context, serviceName string, tags map[string]string) ([]*discovery.ServiceNode, error) {
	nodes, exists := r.nodeListMap.Load(serviceName)
	if !exists {
		var err error
		// 本地缓存为空，从 zookeeper 拉取
		nodes, err = r.pullNodes(ctx, serviceName)
		if err != nil {
			return nil, err
		}
		// 缓存到本地
		r.nodeListMap.Store(serviceName, nodes)
		// 启动watch
		go r.watchNodes(serviceName)
	}
	// 按标签过滤节点
	var filteredNodes []*discovery.ServiceNode
	nodes.(*sync.Map).Range(func(key, value interface{}) bool {
		node := value.(*discovery.ServiceNode)
		if discovery.MatchTags(node.Tags, tags) {
			filteredNodes = append(filteredNodes, node)
		}
		return true
	})

	return filteredNodes, nil
}

func (r *registry) Register(_ context.Context, node *discovery.ServiceNode) error {
	if err := r.ensureServiceNode(node.ServiceName); err != nil {
		return err
	}
	data, err := json.Marshal(node)
	if err != nil {
		return fmt.Errorf("failed to marshal node: %v", err)
	}
	nodePath := fmt.Sprintf("%s/%s", makeServicePath(node.ServiceName), makeNodeKey(node))
	_, err = r.conn.Create(nodePath, data, zk.FlagEphemeral, zk.WorldACL(zk.PermAll))
	if err != nil {
		return fmt.Errorf("failed to create node %v: %v", node, err)
	}
	return nil
}

func makeNodeKey(node *discovery.ServiceNode) string {
	return fmt.Sprintf("%s:%d", node.IP.String(), node.Port)
}

func (r *registry) Unregister(_ context.Context, node *discovery.ServiceNode) error {
	nodePath := fmt.Sprintf("%s/%s", makeServicePath(node.ServiceName), makeNodeKey(node))
	err := r.conn.Delete(nodePath, -1)
	if err != nil {
		return fmt.Errorf("failed to delete node %v: %v", node, err)
	}
	return nil
}

func (r *registry) ensureServiceNode(name string) error {
	nodePath := makeServicePath(name)
	exists, _, err := r.conn.Exists(nodePath)
	if err != nil {
		return fmt.Errorf("failed to check if node exists %v: %v", nodePath, err)
	}
	if !exists {
		_, err = r.conn.Create(nodePath, nil, zk.FlagPersistent, zk.WorldACL(zk.PermAll))
		if err != nil {
			return fmt.Errorf("failed to create node %v: %v", nodePath, err)
		}
	}
	return nil
}

func (r *registry) pullNodes(_ context.Context, serviceName string) (*sync.Map, error) {
	var nodes sync.Map
	nodePath := makeServicePath(serviceName)
	children, _, err := r.conn.Children(nodePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get children of node %v: %v", nodePath, err)
	}
	for _, child := range children {
		data, _, err := r.conn.Get(fmt.Sprintf("%s/%s", nodePath, child))
		if err != nil {
			return nil, fmt.Errorf("failed to get node %v: %v", child, err)
		}
		var node discovery.ServiceNode
		if err := json.Unmarshal(data, &node); err != nil {
			return nil, fmt.Errorf("failed to unmarshal node %v: %v", child, err)
		}
		nodes.Store(child, &node)
	}
	return &nodes, nil
}

func makeServicePath(serviceName string) string {
	return fmt.Sprintf("%s/%s", BasePath, serviceName)
}

func (r *registry) watchNodes(name string) {
	nodePath := makeServicePath(name)
	for {
		children, _, ch, err := r.conn.ChildrenW(nodePath)
		if err != nil {
			fmt.Printf("failed to watch children of node %s: %v\n", nodePath, err)
			return
		}
		var nodes sync.Map
		var count int
		for _, child := range children {
			data, _, err := r.conn.Get(fmt.Sprintf("%s/%s", nodePath, child))
			if err != nil { // 如果是节点解除注册触发的事件，此处无法获取到节点数据，直接跳过即可，不能报错，否则无法更新本地节点
				continue
			}
			var node discovery.ServiceNode
			if err := json.Unmarshal(data, &node); err != nil {
				fmt.Printf("failed to unmarshal node %s: %v\n", child, err)
				return
			}
			count++
			nodes.Store(child, &node)
		}
		r.nodeListMap.Store(name, &nodes)
		select {
		case <-ch:
		}
	}
}

func NewRegistry(servers []string) (discovery.NodeRegistry, error) {
	conn, _, err := zk.Connect(servers, DialTimeout)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to zookeeper: %v", err)
	}
	// create base path
	_, err = conn.Create(BasePath, nil, 0, zk.WorldACL(zk.PermAll))
	if err != nil && !errors.Is(err, zk.ErrNodeExists) {
		return nil, fmt.Errorf("failed to create base path %v: %v", BasePath, err)
	}
	return &registry{
		nodeListMap: &sync.Map{},
		conn:        conn,
	}, nil
}
