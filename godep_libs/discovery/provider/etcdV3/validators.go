package etcdV3

import (
	"fmt"

	"motify_core_api/godep_libs/discovery/provider"
)

func validateRegistrationData(kvs ...provider.KV) error {
	if len(kvs) == 0 {
		return fmt.Errorf("empty registration data")
	}

	for i, kv := range kvs {
		if err := validateKV(kv); err != nil {
			return fmt.Errorf("invalid registration data, kvs[%d]: %s", i, err)
		}
	}

	return nil
}

func validateKV(kv provider.KV) error {
	// Don't validate raw values - guys want to set even empty data
	if kv.RawKey != "" {
		return nil
	}

	if err := kv.Service.Validate(); err != nil {
		return err
	}

	if kv.Namespace == "" {
		return fmt.Errorf("empty Namespace")
	}
	// discovery lib always fills some value for service-based key, so better to check it
	if kv.Value == "" {
		return fmt.Errorf("empty Value")
	}
	return nil
}
