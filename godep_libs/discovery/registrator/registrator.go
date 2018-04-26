package registrator

import (
	"context"
	"fmt"
	"sync"

	"motify_core_api/godep_libs/discovery"
	"motify_core_api/godep_libs/discovery/provider"
)

// IRegistrator is an interface handling registration in Service Discovery, implementing new standard
// https://confluence.lazada.com/pages/viewpage.action?pageId=25904666
//
// IRegistrator is meant to be statefull - it should be initialized with IProvider and persistent registration info.
type IRegistrator interface {
	// Register adds the application instance to the list of services in discovery provider.
	// The method is meant to be non-blocking.
	Register() error
	// Unregister removes the application instance from discovery list.
	// Client must wait until unregistration process is done, otherwise your instance could be still available
	// in the discovery list.
	Unregister()

	// EnableDiscovery explicitly marks the application instance as ready to serve incoming requests.
	// The method is meant to be non-blocking.
	EnableDiscovery() error
	// DisableDiscovery in opposite marks the instance as unable to serve incoming requests.
	// The instance is still registered in the service list, but not ready to serve the requests
	// by any reason.
	DisableDiscovery()
}

type registrator struct {
	provider         provider.IProvider
	registrationInfo IRegistrationInfo
	logger           discovery.ILogger

	registerWg     sync.WaitGroup
	registerCancel context.CancelFunc

	discoveryWg     sync.WaitGroup
	discoveryCancel context.CancelFunc
}

// New returns new service registrator.
// If provider is nil (for example when it could not be initialized) - returns dummy registrator.
func New(p provider.IProvider, info IRegistrationInfo, logger discovery.ILogger) IRegistrator {
	if p == nil {
		return &dummyRegistrator{}
	}

	if logger == nil {
		logger = discovery.NewNilLogger()
	}

	r := &registrator{
		provider:         p,
		registrationInfo: info,
		logger:           logger,
	}
	return r
}

// Register runs registration process in the discovery list.
// If previous call context is not finished returns an error.
func (r *registrator) Register() error {
	if r.registerCancel != nil {
		return fmt.Errorf("already registered")
	}

	ctx, cancel := context.WithCancel(context.Background())
	r.registerCancel = cancel

	r.registerWg.Add(1)
	go func() {
		defer r.registerWg.Done()

		err := r.provider.RegisterValues(ctx, r.registrationInfo.RegistrationData()...)
		if err != nil {
			r.logger.Errorf("registration error: %s", err)
		}
	}()

	r.logger.Infof("registration started")
	return nil
}

// Unregister cancels service registration and waits until finished.
func (r *registrator) Unregister() {
	if r.registerCancel == nil {
		// already canceled, nothing to do
		return
	}

	// disable discovery first
	r.DisableDiscovery()

	r.registerCancel()
	r.registerWg.Wait()
	r.logger.Infof("service registration canceled")

	r.registerCancel = nil
}

// EnableDiscovery registers service for being discovered by other services.
// If previous call context is not finished returns an error.
func (r *registrator) EnableDiscovery() error {
	if r.discoveryCancel != nil {
		return fmt.Errorf("discovery already enabled")
	}

	ctx, cancel := context.WithCancel(context.Background())
	r.discoveryCancel = cancel

	r.discoveryWg.Add(1)
	go func() {
		defer r.discoveryWg.Done()

		err := r.provider.RegisterValues(ctx, r.registrationInfo.DiscoveryData()...)
		if err != nil {
			r.logger.Errorf("error enabling discovery: %s", err)
		}
	}()

	r.logger.Infof("discovery enabled")
	return nil
}

// DisableDiscovery stops service discovery.
func (r *registrator) DisableDiscovery() {
	if r.discoveryCancel == nil {
		// already canceled, nothing to do
		return
	}

	r.discoveryCancel()
	r.discoveryWg.Wait()
	r.logger.Infof("discovery disabled")

	r.discoveryCancel = nil
}
