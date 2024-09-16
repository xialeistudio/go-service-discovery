package consul

import (
	"context"
	"fmt"
	capi "github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/api/watch"
	"github.com/xialeistudio/go-service-discovery/discovery"
	"log"
	"net"
	"sync"
)

type registry struct {
	nodeListMap *sync.Map // <serviceName, <nodeKey, *ServiceNode>>
	config      *capi.Config
	client      *capi.Client
}

func (r *registry) GetNodes(ctx context.Context, serviceName string, tags map[string]string) ([]*discovery.ServiceNode, error) {
	nodes, exists := r.nodeListMap.Load(serviceName)
	if !exists {
		var err error
		// 本地缓存为空，从 consul 拉取
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

func (r *registry) Register(ctx context.Context, node *discovery.ServiceNode) error {
	service := &capi.AgentServiceRegistration{
		ID:      makeNodeKey(node),
		Name:    node.ServiceName,
		Address: node.IP.String(),
		Port:    node.Port,
		Meta:    node.Tags,
		Check: &capi.AgentServiceCheck{
			TCP:                            fmt.Sprintf("%s:%d", node.IP.String(), node.Port),
			Timeout:                        "5s",
			Interval:                       "5s",
			DeregisterCriticalServiceAfter: "30s",
			Status:                         capi.HealthPassing,
		},
	}
	err := r.client.Agent().ServiceRegisterOpts(service, capi.ServiceRegisterOpts{}.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("register service error: %w", err)
	}
	return nil
}

func (r *registry) Unregister(ctx context.Context, node *discovery.ServiceNode) error {
	opts := &capi.QueryOptions{}
	opts = opts.WithContext(ctx)
	err := r.client.Agent().ServiceDeregisterOpts(makeNodeKey(node), opts)
	if err != nil {
		return fmt.Errorf("unregister service error: %w", err)
	}
	return nil
}

func (r *registry) pullNodes(ctx context.Context, name string) (*sync.Map, error) {
	opts := &capi.QueryOptions{}
	opts = opts.WithContext(ctx)
	services, err := r.client.Agent().ServicesWithFilterOpts(fmt.Sprintf(`Service=="%s"`, name), opts)
	if err != nil {
		return nil, fmt.Errorf("pull services error: %w", err)
	}
	var results sync.Map
	for _, service := range services {
		node := &discovery.ServiceNode{
			ServiceName: name,
			IP:          net.ParseIP(service.Address),
			Port:        service.Port,
			Tags:        service.Meta,
		}
		results.Store(makeNodeKey(node), node)
	}
	return &results, nil
}

func (r *registry) watchNodes(name string) {
	params := map[string]any{
		"type":    "service",
		"service": name,
	}
	plan, _ := watch.Parse(params)
	plan.Handler = func(u uint64, i interface{}) {
		if i == nil {
			return
		}
		val, ok := i.([]*capi.ServiceEntry)
		if !ok {
			return
		}
		if len(val) == 0 {
			r.nodeListMap.Store(name, &sync.Map{})
			return
		}
		var (
			healthInstances sync.Map
		)
		for _, instance := range val {
			if instance.Service.Service != name {
				continue
			}
			if instance.Checks.AggregatedStatus() == capi.HealthPassing {
				node := &discovery.ServiceNode{
					ServiceName: name,
					IP:          net.ParseIP(instance.Service.Address),
					Port:        instance.Service.Port,
					Tags:        instance.Service.Meta,
				}
				healthInstances.Store(makeNodeKey(node), node)
			}
		}
		r.nodeListMap.Store(name, &healthInstances)
	}
	defer plan.Stop()
	if err := plan.Run(r.config.Address); err != nil {
		log.Printf("watch service %s error: %v", name, err)
	}
}

func NewRegistry(config *capi.Config) (discovery.NodeRegistry, error) {
	client, err := capi.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("new consul client error: %w", err)
	}
	return &registry{
		nodeListMap: new(sync.Map),
		client:      client,
		config:      config,
	}, nil
}

func makeNodeKey(node *discovery.ServiceNode) string {
	return fmt.Sprintf("%s/%s:%d", node.ServiceName, node.IP.String(), node.Port)
}
