package registrator

// Deprecated: this file contains deprecated structs and interfaces for
// etcd2 registration. All of them were moved here from the discovery/discovery package,
// for private usage in v2v3 registrator.
//
// TODO: purge after dropping etcd2 support

import (
	"context"
	"fmt"
	"strings"
	"time"

	etcdcl "github.com/coreos/etcd/client"
	"motify_core_api/godep_libs/discovery"
)

const defaultKeyTTL = 30 * time.Second

// registrationInfoV2 contains etcdV2 service discovery information.
type registrationInfoV2 struct {
	Namespace   string
	Venture     string
	Environment string
	ServiceName string
	Property    string
	// Service info TTL
	TTL time.Duration
	// Service info update interval
	Interval time.Duration
	Key      string
	Value    string
}

// Key returns full key in key-value store
func (i registrationInfoV2) StorageKey() string {
	return fmt.Sprintf("%s/%s/%s/%s/%s/", i.Namespace, i.Venture, i.Environment, i.ServiceName, i.Property)
}

type etcdRegistrator struct {
	logger  discovery.ILogger
	keysAPI etcdcl.KeysAPI
}

// newEtcdRegistrator creates ServiceRegistrator which will update service information in etcd in certain intervals.
func newEtcdRegistrator(client etcdcl.Client, logger discovery.ILogger) *etcdRegistrator {
	if logger == nil {
		logger = discovery.NewNilLogger()
	}
	return &etcdRegistrator{
		logger:  logger,
		keysAPI: etcdcl.NewKeysAPI(client),
	}
}

// Run starts service registration in certain intervals. This method is blocking.
func (r *etcdRegistrator) Register(ctx context.Context, info registrationInfoV2) {
	if info.TTL == 0 {
		info.TTL = defaultKeyTTL
	}
	if info.Interval == 0 {
		// choose registration interval smaller than expiration time of registration info to keep it alive
		info.Interval = info.TTL / 2
	}
	key := info.StorageKey() + info.Key

	value := info.Value
	// temporary hack, GO-4096
	// TODO: remove when Infra is ready
	// https://confluence.lazada.com/display/INFRA/Standards+and+agreements
	//
	// We should be ready for Infra change - don't duplicate "-pub" if found.
	if info.Property == discovery.NodesProperty && !strings.Contains(value, "-pub") {
		if i := strings.LastIndex(value, ".iddc"); i != -1 {
			value = value[:i] + "-pub" + value[i:]
		}
		if i := strings.LastIndex(value, ".sgdc"); i != -1 {
			value = value[:i] + "-pub" + value[i:]
		}
		if i := strings.LastIndex(value, ".hkdc"); i != -1 {
			value = value[:i] + "-pub" + value[i:]
		}
	}

	// We must pass background context to unregister() because defer unregisration will work when original context
	// is already canceled.
	defer r.unregister(context.Background(), key)

	ticker := time.NewTicker(info.Interval)
	defer ticker.Stop()

	for {
		registerCtx, cancelRegister := context.WithTimeout(ctx, info.Interval)
		go r.register(registerCtx, key, value, info.TTL)

		select {
		case <-ctx.Done():
			cancelRegister()
			return
		case <-ticker.C:
			cancelRegister()
		}
	}
}

func (r *etcdRegistrator) register(ctx context.Context, key, value string, ttl time.Duration) {
	r.logger.Debugf("register %q=%q, ttl: %s", key, value, ttl)

	_, err := r.keysAPI.Set(ctx, key, value, &etcdcl.SetOptions{TTL: ttl, PrevExist: etcdcl.PrevIgnore})
	if err != nil && err != context.DeadlineExceeded && err != context.Canceled {
		r.logger.Warningf("could not register %q: %v", key, err)
	}
}

func (r *etcdRegistrator) unregister(ctx context.Context, key string) {
	r.logger.Debugf("unregister %q", key)

	_, err := r.keysAPI.Delete(ctx, key, &etcdcl.DeleteOptions{})
	if err != nil && err != context.DeadlineExceeded && err != context.Canceled {
		r.logger.Warningf("could not unregister %q: %v", key, err)
	}
}
