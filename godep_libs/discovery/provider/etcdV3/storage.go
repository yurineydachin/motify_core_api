package etcdV3

import (
	"fmt"
	"strings"

	"godep.lzd.co/discovery/provider"
)

// serviceKey represents storage key structure
type serviceKey struct {
	Namespace string
	Service   provider.Service
}

func newServiceKey(namespace string, service provider.Service) serviceKey {
	return serviceKey{
		Namespace: namespace,
		Service:   service,
	}
}

// parseServiceKey parses key string into discovery service key struct
func parseServiceKey(k string) (serviceKey, error) {
	parts := strings.Split(k, "/")
	if len(parts) != 8 {
		return serviceKey{}, fmt.Errorf("can't parse key '%s': parts number missmatch", k)
	}
	for i, s := range parts {
		if i == 0 {
			// first part is empty, should not check it
			continue
		}
		if s == "" {
			return serviceKey{}, fmt.Errorf("can't parse key '%s': empty part found", k)
		}
	}

	t, err := provider.ParseServiceType(parts[2])
	if err != nil {
		return serviceKey{}, fmt.Errorf("can't parse key '%s': %s", k, err)
	}

	return serviceKey{
		Namespace: parts[1],
		Service: provider.Service{
			Type:         t,
			Name:         parts[3],
			RolloutType:  parts[4],
			Owner:        parts[5],
			ClusterType:  parts[6],
			InstanceName: provider.InstanceName(parts[7]),
		},
	}, nil
}

// storageKey returns storage key string
// Example: /discovery/app/bob_api/stable/shared/common/thlzdlivego2.sgdc:5031
func (k serviceKey) storageKey() string {
	return fmt.Sprintf("/%s/%s/%s/%s/%s/%s/%s",
		k.Namespace, k.Service.Type, k.Service.Name, k.Service.RolloutType, k.Service.Owner,
		k.Service.ClusterType, k.Service.InstanceKey())
}

// keyPrefix returns key prefix from discovery key filter
func keyPrefix(filter provider.KeyFilter) string {
	if filter.Prefix != "" {
		return filter.Prefix
	}

	prefix := fmt.Sprintf("/%s/%s/%s/%s/%s/%s/",
		filter.Namespace, filter.Type, filter.Name, filter.RolloutType, filter.Owner,
		filter.ClusterType)
	// Trim left any "//" in prefix - to clean-up empty parts
	parts := strings.Split(prefix, "//")
	if len(parts) > 1 {
		prefix = parts[0] + "/"
	}

	return prefix
}
