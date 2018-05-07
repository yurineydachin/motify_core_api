package healthcheck

import (
	"context"
)

const (
	// Please use this constants for standard resources status checkers initialization
	ResourceTypeMySQL       = "mysql_status"
	ResourceTypeAerospike   = "aerospike_status"
	ResourceTypeElastic     = "elastic_search_status"
	ResourceTypeRabbitMQ    = "rabbit_mq_status"
	ResourceTypeMemcache    = "memcache_status"
	ResourceTypeCeph        = "ceph_status"
	ResourceTypeNfs         = "nfs_status"
	ResourceTypeEtcd        = "etcd_status"
	ResourceTypeStructCache = "struct_cache_status"
)

const (
	// There are possible resource status values in health check response
	ServiceResourceStatusOk       = "Ok"       // Resource is OK
	ServiceResourceStatusUnstable = "Unstable" // Resource is NOT OK, but service can work without it (resource is not critical)
	ServiceResourceStatusError    = "Error"    // Resource is NOT OK and it's critical
)

var deprecatedResourceTypeNames = map[string]struct{}{
	// main manadatory fields that must not be used as resource names
	"service":      {},
	"status":       {},
	"venture":      {},
	"version":      {},
	"build_date":   {},
	"git_describe": {},
	"go_version":   {},

	// common mistakes in resource names are not allowed anymore
	"is_db_ok":           {},
	"is_elastic_ok":      {},
	"is_cache_ok":        {},
	"is_memcache_ok":     {},
	"is_aerospike_ok":    {},
	"is_auto_cache_ok":   {},
	"is_struct_cache_ok": {},
	"db_is_ok":           {},
	"memcache_is_ok":     {},
	"cache_is_ok":        {},
	"elastic_is_ok":      {},
	"aerospike_is_ok":    {},
	"auto_cache_is_ok":   {},
	"struct_cache_is_ok": {},

	"error_details": {},
}

// initStatusCollector creates function that returns one out of three string values: "Ok", "Unstable", "Error",
// depending on error that provided checker throws.
func initStatusCollector(isCritical bool, checker func(context.Context) error) func(context.Context) interface{} {
	return func(ctx context.Context) interface{} {
		result := make(chan error, 1)
		go func() {
			defer func() {
				if r := recover(); r != nil {
					result <- ErrInvalidFunc
				}
			}()
			result <- checker(ctx)
		}()

		select {
		case <-ctx.Done():
			return getResourceStatus(isCritical, ctx.Err())
		case err := <-result:
			return getResourceStatus(isCritical, err)
		}
	}
}

// getResourceStatus selects one out of three strings for return, depending on isCritical and Error parameters.
// spec in confluence: https://confluence.lzd.co/display/DEV/Microservice+Architecture+%28SOA%29+Conventions#MicroserviceArchitecture(SOA)Conventions-GET<service>:<port>/health_check
func getResourceStatus(isCritical bool, err error) string {
	if err != nil {
		if isCritical {
			return ServiceResourceStatusError
		}
		return ServiceResourceStatusUnstable
	}
	return ServiceResourceStatusOk
}

func isKeyDeprecated(name string) bool {
	_, deprecated := deprecatedResourceTypeNames[name]
	return deprecated
}
