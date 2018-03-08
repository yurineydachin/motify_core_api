package locator

// Deprecated: this file contains deprecated structs and interfaces for
// etcd2 service locator. All of them were moved here from the discovery/discovery package.
//
// TODO: purge after dropping etcd2 support

import (
	"context"
	"fmt"
	"time"

	etcdcl "github.com/coreos/etcd/client"
	"godep.lzd.co/discovery"
)

// LocationInfo contains params to get service discovery information.
type LocationInfo struct {
	Namespace   string
	Venture     string
	Environment string
	ServiceName string
	Property    string
}

// StorageKey returns full key in key-value store
func (i LocationInfo) StorageKey() string {
	return fmt.Sprintf("%s/%s/%s/%s/%s/", i.Namespace, i.Venture, i.Environment, i.ServiceName, i.Property)
}

// ServiceInfo contains service discovery information.
type ServiceInfo struct {
	Name  string
	Value string
}

// IServiceLocator2 helps to locate services.
type IServiceLocator2 interface {
	Get(ctx context.Context, info LocationInfo) ([]ServiceInfo, error)
	Locate(ctx context.Context, info LocationInfo, out chan *etcdcl.Response) error
}

type etcdLocator struct {
	logger  discovery.ILogger
	keysAPI etcdcl.KeysAPI
}

// NewLocatorEtcd2 returns new IServiceLocator2 instance
func NewLocatorEtcd2(client etcdcl.Client, logger discovery.ILogger) IServiceLocator2 {
	return &etcdLocator{
		logger:  logger,
		keysAPI: etcdcl.NewKeysAPI(client),
	}
}

func (l *etcdLocator) Get(ctx context.Context, info LocationInfo) ([]ServiceInfo, error) {
	resp, err := l.keysAPI.Get(ctx, info.StorageKey(), &etcdcl.GetOptions{
		Recursive: false,
		Sort:      false,
	})
	if err != nil {
		return nil, err
	}

	services := make([]ServiceInfo, len(resp.Node.Nodes))
	for i, node := range resp.Node.Nodes {
		services[i] = ServiceInfo{Name: node.Key, Value: node.Value}
	}
	return services, nil
}

func (l *etcdLocator) Locate(parentCtx context.Context, info LocationInfo, out chan *etcdcl.Response) error {
	ctx, cancel := context.WithCancel(parentCtx)
	defer cancel()

	key := info.StorageKey()
	var backoff Backoff

	handlerError := func(err error) {
		l.logger.Warningf("unexpected etcd error: %v", err)

		if etcdErr, ok := err.(etcdcl.Error); ok {
			// Key not found
			if etcdErr.Code == etcdcl.ErrorCodeKeyNotFound {
				_, err := l.keysAPI.Set(ctx, key, "", &etcdcl.SetOptions{
					PrevExist: etcdcl.PrevNoExist,
					Dir:       true,
				})
				if err == nil {
					return
				}
				l.logger.Warningf("failed to create key %q: %v", key, err)
			}
		}

		// Prevent errors from consuming all resources.
		select {
		case <-time.After(backoff.Duration()):
		case <-ctx.Done():
		}
	}

	for {
		resp, err := l.keysAPI.Get(ctx, key, &etcdcl.GetOptions{Recursive: true})
		if err != nil {
			switch err {
			case context.DeadlineExceeded, context.Canceled:
				return err
			default:
				handlerError(err)
				continue
			}
		}
		waitIndex := resp.Index
		out <- resp
		backoff.Reset()

		watcher := l.keysAPI.Watcher(key, &etcdcl.WatcherOptions{
			AfterIndex: waitIndex,
			Recursive:  true,
		})
	WATCH:
		for waitIndex != 0 {
			resp, err := watcher.Next(ctx)
			if err != nil {
				switch err {
				case context.DeadlineExceeded, context.Canceled:
					return err
				default:
					handlerError(err)
					break WATCH
				}
			}
			waitIndex = resp.Node.ModifiedIndex
			out <- resp
			backoff.Reset()
		}
	}
}
