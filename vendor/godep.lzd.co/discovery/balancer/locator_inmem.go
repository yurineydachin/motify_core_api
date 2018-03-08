package balancer

import (
	"context"
	"strings"

	etcd "github.com/coreos/etcd/client"
	"godep.lzd.co/discovery/locator"
)

// inMemLocator is used for testing
type inMemLocator struct {
	nodes []string
}

// NewInMemLocator returns new inMemLocator instance for testing
func NewInMemLocator(nodes []string) *inMemLocator {
	return &inMemLocator{nodes}
}

// Get returns predefined nodes list
func (l *inMemLocator) Get(ctx context.Context, info locator.LocationInfo) ([]locator.ServiceInfo, error) {
	discoveryInfo := make([]locator.ServiceInfo, 0, len(l.nodes))
	for _, v := range l.nodes {
		values := strings.Split(v, "=")
		discoveryInfo = append(discoveryInfo, locator.ServiceInfo{Name: values[0], Value: values[1]})
	}
	return discoveryInfo, nil
}

// Locate returns etcd locate response for predefined nodes list
func (l *inMemLocator) Locate(ctx context.Context, info locator.LocationInfo, out chan *etcd.Response) error {
	etcdNodes := make(etcd.Nodes, 0, len(l.nodes))
	for _, v := range l.nodes {
		values := strings.Split(v, "=")
		etcdNodes = append(etcdNodes, &etcd.Node{Key: values[0], Value: values[1]})
	}
	if len(etcdNodes) != 0 {
		out <- &etcd.Response{Action: "get", Node: &etcd.Node{Nodes: etcdNodes}}
	}
	return nil
}
