# go-service-discovery

go-service-discovery 是一个为现代云原生应用设计的基于 Go 语言的微服务发现工具包。它提供了一个强大的框架，用于服务注册、发现和监控，利用 etcd 的力量进行分布式协调。

## 特性
+ 服务注册与发现：使用集中式注册中心注册和发现微服务。
+ 负载均衡：提供内置的负载均衡策略，包括轮询和随机选择等。
+ 基于标签的发现：允许基于标签发现服务，以实现更细粒度的控制。
+ 观察机制：监听服务实例的变化，并相应地更新本地缓存。

## 支持的平台
+ etcd v3.x
+ consul v1

## 快速开始

以下以 etcd 为例展示如何使用。

```go
r, err := NewRegistry([]string{"localhost:2379"})
if err != nil {
log.Fatalf("failed to create registry: %v", err)
}
// Register a service node
err = r.Register(context.Background(), node)
// Unregister a service node
err = r.Unregister(context.Background(), node)
// Discover a service node
nodes, err := r.Discover(context.Background(), "service-name", map[string]string{
// tags
})
```