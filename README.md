# go-service-discovery

[中文文档](README-CN.md)

A Go-based microservices discovery toolkit designed for modern cloud-native applications. It provides a robust framework
for service registration, discovery, and monitoring, leveraging the power of etcd for distributed coordination.

## Features

+ Service Registration and Discovery: Register and discover microservices using a centralized registry.
+ Load Balancing: Offers built-in load balancing strategies including round-robin and random selection, etc.
+ Tag-based Discovery: Allows services to be discovered based on tags for more granular control.
+ Watch Mechanism: Watches for changes in service instances and updates the local cache accordingly.

## Supporting platforms

+ etcd v3.x

## Get Started

The following code uses etcd as an example to demonstrate how to use it.

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
