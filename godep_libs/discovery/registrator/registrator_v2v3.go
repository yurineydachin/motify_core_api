package registrator

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"sync"

	etcdclient "github.com/coreos/etcd/client"
	"godep.lzd.co/discovery"
	"godep.lzd.co/discovery/provider"
)

const (
	// namespace is a common namespace for all Lazada services in discovery V2
	namespace = "motify_api"
)

// registratorV2V3 is temp IRegistrator implementation, aggregating both V2 and V3 registrators under single interface.
// TODO: remove totally after all services move to etcdV3
type registratorV2V3 struct {
	logger discovery.ILogger

	registrationParams *AppRegistrationParams
	registratorV2      *etcdRegistrator
	registratorV3      IRegistrator

	discoveryWg     sync.WaitGroup
	discoveryCancel context.CancelFunc
}

// NewV2V3 returns registrator supporting both V2 and V3 registrators
func NewV2V3(params AppRegistrationParams, p provider.IProvider,
	client etcdclient.Client, logger discovery.ILogger) (IRegistrator, error) {

	if logger == nil {
		logger = discovery.NewNilLogger()
	}

	if err := validateV2Info(params); err != nil {
		return nil, err
	}

	info, err := NewAppRegistrationInfo(params)
	if err != nil {
		return nil, err
	}
	rV3 := New(p, info, logger)

	r := &registratorV2V3{
		logger:             logger,
		registrationParams: &params,
		registratorV2:      newEtcdRegistrator(client, logger),
		registratorV3:      rV3,
	}
	return r, nil
}

// Register runs asynchronous registration process.
// It's just a registratorV3 registration, because old discovery convention does not
// support such functionality.
func (r *registratorV2V3) Register() error {
	return r.registratorV3.Register()
}

// Unregister cancels service registration and waits until finished
func (r *registratorV2V3) Unregister() {
	// disable discovery first - it unregisters the service from V2 completely
	r.DisableDiscovery()

	r.registratorV3.Unregister()
}

// EnableDiscovery registers service for being discovered by other services.
// If previous call context is not finished returns an error
func (r *registratorV2V3) EnableDiscovery() error {
	if r.discoveryCancel != nil {
		return fmt.Errorf("discovery already enabled")
	}

	// Register in etcdV3
	err := r.registratorV3.EnableDiscovery()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	r.discoveryCancel = cancel
	// Register in etcdV2 if rollout type is stable.
	// Unstable services should not register in V2 discovery because otherwise
	// old services may discover them along with stable ones.
	if r.registrationParams.RolloutType == provider.RolloutTypeStable {
		r.registerV2(ctx)
	}

	return nil
}

// DisableDiscovery stops service discovery
func (r *registratorV2V3) DisableDiscovery() {
	if r.discoveryCancel == nil {
		// already canceled, nothing to do
		return
	}

	// disable V3 discovery
	r.registratorV3.DisableDiscovery()

	r.discoveryCancel()
	r.discoveryWg.Wait()
	r.logger.Infof("discovery in registrator V2 is disabled")

	r.discoveryCancel = nil
}

func (r *registratorV2V3) registerV2(ctx context.Context) {
	r.discoveryWg.Add(1)
	go func() {
		defer r.discoveryWg.Done()
		r.registratorV2.Register(
			ctx,
			nodesRegistrationInfo(r.registrationParams),
		)
	}()

	r.discoveryWg.Add(1)
	go func() {
		defer r.discoveryWg.Done()
		r.registratorV2.Register(
			ctx,
			versionsRegistrationInfo(r.registrationParams),
		)
	}()

	r.discoveryWg.Add(1)
	go func() {
		defer r.discoveryWg.Done()
		r.registratorV2.Register(
			ctx,
			monitoringRegistrationInfo(r.registrationParams),
		)
	}()
}

// validateV2Info makes some explicit validation of values, used for registering
// in old etcd2 discovery.
func validateV2Info(params AppRegistrationParams) error {
	errorTmpl := "invalid registration params: %s"

	switch {
	case params.Environment == "":
		return fmt.Errorf(errorTmpl, "Environment is empty")
	case params.Venture == "":
		return fmt.Errorf(errorTmpl, "Venture is empty")
	case params.ServiceName == "":
		return fmt.Errorf(errorTmpl, "ServiceName is empty")
	case params.HTTPPort <= 0 || params.HTTPPort > 65535:
		return fmt.Errorf(errorTmpl, "invalid HTTPPort")
	case params.Host == "":
		return fmt.Errorf(errorTmpl, "Host is empty")
	}
	return nil
}

// getKey returns propper registrationInfoV2.Key for application
func getKey(params *AppRegistrationParams) string {
	return fmt.Sprintf("%s:%d", params.Host, params.HTTPPort)
}

func nodesRegistrationInfo(params *AppRegistrationParams) registrationInfoV2 {
	return registrationInfoV2{
		Namespace:   namespace,
		Venture:     params.Venture,
		Environment: params.Environment,
		ServiceName: params.ServiceName,
		Value:       fmt.Sprintf("http://%s:%d", params.Host, params.HTTPPort),
		Key:         getKey(params),
		Property:    discovery.NodesProperty,
	}
}

func versionsRegistrationInfo(params *AppRegistrationParams) registrationInfoV2 {
	return registrationInfoV2{
		Namespace:   namespace,
		Venture:     params.Venture,
		Environment: params.Environment,
		ServiceName: params.ServiceName,
		Value:       params.Version.AppVersion,
		Key:         getKey(params),
		Property:    discovery.VersionsProperty,
	}
}

func monitoringRegistrationInfo(params *AppRegistrationParams) registrationInfoV2 {
	return registrationInfoV2{
		Namespace:   namespace,
		Venture:     params.Venture,
		Environment: params.Environment,
		ServiceName: params.ServiceName,
		Value:       monitoringV2Value(params),
		Key:         getKey(params),
		Property:    discovery.MetricsProperty,
	}
}

// monitoringV2Value returns monitoring value for registrator V2.
// The format is slightly changed in V3, so we need to keep the compatibility,
// but I don't want to import "godep.lzd.co/monitoring_prometheus" as is done in every component
func monitoringV2Value(params *AppRegistrationParams) string {
	// ignore the error, because it's checked in validation
	b, _ := json.Marshal(monitoringValue{
		URL: net.JoinHostPort(params.Host, strconv.Itoa(params.MonitoringPort)),
		Tags: []metricsTag{
			{"venture", params.Venture},
			{"env", params.Environment},
			{"service", params.ServiceName},
		},
	})
	return string(b)
}
