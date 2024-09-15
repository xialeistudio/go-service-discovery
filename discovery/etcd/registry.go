package etcd

import (
	"context"
	"fmt"
	json "github.com/json-iterator/go"
	"github.com/xialeistudio/go-service-discovery/discovery"
	"go.etcd.io/etcd/client/v3"
	"log"
	"strconv"
	"sync"
	"time"
)

var (
	// DialTimeout 默认的连接超时时间
	DialTimeout = 5 * time.Second
	// LeaseTTL 默认的租约时间
	LeaseTTL = 10 * time.Second
)

type registry struct {
	nodeListMap *sync.Map // <serviceName, <nodeKey, *ServiceNode>>

	client  *clientv3.Client
	watcher clientv3.Watcher
	kv      clientv3.KV
}

func NewRegistry(endpoints []string) (discovery.NodeRegistry, error) {
	config := clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: DialTimeout,
	}
	client, err := clientv3.New(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create etcd client: %v", err)
	}
	return &registry{
		nodeListMap: &sync.Map{},
		client:      client,
		kv:          clientv3.NewKV(client),
		watcher:     clientv3.NewWatcher(client),
	}, nil
}

func (r *registry) GetNodes(ctx context.Context, serviceName string, tags map[string]string) ([]*discovery.ServiceNode, error) {
	// 检查本地缓存是否有服务节点

	nodes, exists := r.nodeListMap.Load(serviceName)
	if !exists {
		var err error
		// 本地缓存为空，从 etcd 拉取
		nodes, err = r.pullNodes(ctx, serviceName)
		if err != nil {
			return nil, err
		}
		// 缓存到本地
		r.nodeListMap.Store(serviceName, nodes)
		// 启动协程监听节点变化
		go r.watchNodes(ctx, serviceName)
	}

	// 根据标签过滤服务节点
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

func (r *registry) Register(ctx context.Context, node *discovery.ServiceNode) error {
	value, err := json.MarshalToString(node)
	if err != nil {
		return fmt.Errorf("failed to marshal service node: %v", err)
	}
	// 创建租约
	lease, err := r.client.Grant(ctx, int64(LeaseTTL.Seconds()))
	if err != nil {
		return err
	}
	// 将服务节点信息写入etcd
	key := makeNodeKey(node)
	_, err = r.kv.Put(ctx, key, value, clientv3.WithLease(lease.ID))
	if err != nil {
		return fmt.Errorf("failed to put key to etcd: %v", err)
	}

	// 续租
	go r.keepLeaseAlive(lease.ID, node)
	return nil
}

func (r *registry) Unregister(ctx context.Context, node *discovery.ServiceNode) error {
	key := makeNodeKey(node)
	_, err := r.kv.Delete(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to delete key from etcd: %v", err)
	}
	r.removeNode(node)
	return nil
}

func makeNodeKey(node *discovery.ServiceNode) string {
	return node.ServiceName + "/" + node.IP.String() + ":" + strconv.Itoa(node.Port)
}

// 续租
func (r *registry) keepLeaseAlive(leaseId clientv3.LeaseID, node *discovery.ServiceNode) {
	ticker := time.NewTicker(LeaseTTL / 2)
	defer ticker.Stop()
	for range ticker.C {
		func() {
			_, err := r.client.KeepAliveOnce(context.Background(), leaseId)
			if err != nil {
				r.removeNode(node)
				return
			}
		}()
	}
}

func (r *registry) removeNode(node *discovery.ServiceNode) {
	r.nodeListMap.Range(func(key, value interface{}) bool {
		nodes := value.(*sync.Map)
		nodes.Delete(makeNodeKey(node))
		return true
	})
}

func (r *registry) pullNodes(ctx context.Context, serviceName string) (*sync.Map, error) {
	resp, err := r.kv.Get(ctx, serviceName+"/", clientv3.WithPrefix())
	if err != nil {
		return nil, fmt.Errorf("failed to get nodeListMap from etcd: %v", err)
	}
	var nodes sync.Map
	for _, kv := range resp.Kvs {
		node := &discovery.ServiceNode{}
		if err := json.Unmarshal(kv.Value, node); err != nil {
			return nil, fmt.Errorf("failed to unmarshal node data: %v", err)
		}
		nodes.Store(string(kv.Key), node)
	}

	return &nodes, nil
}

func (r *registry) watchNodes(ctx context.Context, serviceName string) {
	watchChan := r.watcher.Watch(ctx, serviceName+"/", clientv3.WithPrefix())
	for wResp := range watchChan {
		for _, ev := range wResp.Events {
			switch ev.Type {
			case clientv3.EventTypePut:
				// 新增或更新服务节点
				node := &discovery.ServiceNode{}
				if err := json.Unmarshal(ev.Kv.Value, node); err != nil {
					log.Printf("failed to unmarshal node data: %v", err)
					continue
				}
				r.nodeListMap.LoadOrStore(serviceName, &sync.Map{})
				nodes, _ := r.nodeListMap.Load(serviceName)
				nodes.(*sync.Map).Store(string(ev.Kv.Key), node)
			case clientv3.EventTypeDelete:
				// 删除服务节点
				nodes, _ := r.nodeListMap.Load(serviceName)
				nodes.(*sync.Map).Delete(string(ev.Kv.Key))
			}
		}
	}
}
