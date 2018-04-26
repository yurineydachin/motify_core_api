package healthcheck

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"sync"
	"time"
)

const (
	// DefaultHealthCheckPath must be used as handler path for HC by SOA convention
	DefaultHealthCheckPath = "/health_check"

	// ServiceStatusOK - "status" value in response for Successful HC
	ServiceStatusOK = "Ok"
	// ServiceStatusOK - "status" value in response for Failed HC
	ServiceStatusError = "Error"

	// TimeFormat date/time format in health check handler response
	TimeFormat = "2006-01-02 15:04:05 -07:00"

	// wait for health_check results timeout
	handlerTimeout = time.Millisecond * 1000
)

var (
	// ErrInvalidFunc must be returned in case of nil checker func usage in SetResourceChecker() method call
	ErrInvalidFunc = errors.New("Checker func is nil")
	// ErrKeyAlreadyExist must be returned in case of double call of SetResourceChecker() method for same resource type
	ErrKeyAlreadyExist = errors.New("Checker for resource type already exist")
	// ErrKeyDeprecated must be returned in case of deprecated resource name usage in SetResourceChecker() method call
	ErrKeyDeprecated = errors.New("Resource type name is deprecated. Please use proper one")
	// ErrRequiredParamMissed must be returned if at least one of required parameters is empty
	ErrInvalidRequiredParam = errors.New("Empty or invalid required parameter")
)

// HealthCheck represents health check aggregator.
// It allows you to define checkers for all resources that your service uses
// and easily get http.Hanlder or http.HandlerFunc for health check feature.
// It fits convention https://confluence.lzd.co/display/DEV/Microservice+Architecture+%28SOA%29+Conventions#MicroserviceArchitecture(SOA)Conventions-GET%3Cservice%3E%3A%3Cport%3E/health_check
type HealthCheck struct {
	fieldsMtx         sync.RWMutex
	appConfigRequired bool
	appConfigReady    bool
	appConfigPath     string
	service           string
	version           string
	buildDate         string
	gitDescribe       string
	goVersion         string
	venture           string

	checkersMtx sync.RWMutex
	checkers    map[string]func(context.Context) interface{}
}

// lib is interface of resource client library that has implemented "HealthCheck(context.Context)error" method
// that applicable to use as resource checker
type lib interface {
	HealthCheck(context.Context) error
}

// internalErrorResponse response structure that used for failover in case of unrecoverable error
type internalErrorResponse struct {
	Service        string `json:"service"`
	Venture        string `json:"venture"`
	Version        string `json:"version"`
	BuildDate      string `json:"build_date"`
	GitDescribe    string `json:"git_describe"`
	AppConfigPath  string `json:"app_config_path"`
	AppConfigReady bool   `json:"app_config_ready"`
	GoVersion      string `json:"go_version"`
	Status         string `json:"status"`
	ErrorDetail    string `json:"error_details"`
}

func New(serviceID, version, gitDescribe, goVersion, venture string, buildDate int64, appEtcdConfigRequired bool) (hc *HealthCheck, err error) {
	hc = NewHealthCheck(serviceID, version, gitDescribe, goVersion, venture, buildDate, appEtcdConfigRequired)
	if hc == nil {
		err = ErrInvalidRequiredParam
	}

	return
}

// NewHealthCheck creates base health check without any external resources checkers
// @deprecated
func NewHealthCheck(serviceID, version, gitDescribe, goVersion, venture string, buildDate int64, appConfigRequired bool) *HealthCheck {
	if serviceID == "" || version == "" || goVersion == "" || buildDate <= 0 {
		return nil
	}

	hc := &HealthCheck{
		service:           serviceID,
		venture:           venture,
		buildDate:         time.Unix(buildDate, 0).UTC().Format(TimeFormat),
		gitDescribe:       gitDescribe,
		version:           version,
		goVersion:         goVersion,
		appConfigRequired: appConfigRequired,
		checkers:          make(map[string]func(context.Context) interface{}),
	}

	return hc
}

// SetResourceChecker performs setting a checker for external resource.
// Please use pre-defined constants for common resource types like MySQL, Aerospike, Elasticsearch etc. (ResourceTypeMySQL, ResourceTypeAerospike, ...)
func (hc *HealthCheck) SetResourceChecker(resourceType string, isCritical bool, checker interface{}) error {
	hc.checkersMtx.Lock()
	defer hc.checkersMtx.Unlock()

	if err := hc.validate(resourceType, checker); err != nil {
		return err
	}

	// if provided checker is function with signature "func(context.Context)error"
	if f, ok := checker.(func(context.Context) error); ok {
		hc.checkers[resourceType] = initStatusCollector(isCritical, f)
		return nil
	}

	// if provided checker is object that have HealthCheck method
	if l, ok := checker.(lib); ok {
		hc.checkers[resourceType] = initStatusCollector(isCritical, l.HealthCheck)
		return nil
	}

	return ErrInvalidFunc
}

// SetCustomValue performs setting a callback for custom key in health_check handler JSON response.
// Returning interface{} must work fine with json.Marshal() otherwise healthcheck returns failover error response.
func (hc *HealthCheck) SetCustomValue(key string, f func(ctx context.Context) interface{}) error {
	hc.checkersMtx.Lock()
	defer hc.checkersMtx.Unlock()

	if err := hc.validate(key, f); err != nil {
		return err
	}

	hc.checkers[key] = f

	return nil
}

// HandlerFunc returns http.HandlerFunc interface for health check requests processing
func (hc *HealthCheck) HandlerFunc() http.HandlerFunc {
	return hc.handleHealthCheckRequest
}

// Handler returns http.Handler interface for health check requests processing
func (hc *HealthCheck) Handler() http.Handler {
	return hc
}

// ServeHTTP implements http.Handler interface
func (hc *HealthCheck) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	hc.handleHealthCheckRequest(w, r)
}

// SetAppConfigInfo allows you to set application etcd config path and app_config_ready flag 'on fly'
func (hc *HealthCheck) SetAppConfigInfo(appConfigPath string, appConfigReady bool) {
	hc.fieldsMtx.Lock()
	hc.appConfigPath = appConfigPath
	hc.appConfigReady = appConfigReady
	hc.fieldsMtx.Unlock()
}

// SetVenture sets venture string (useful if your service reads venture value from ETCD after service start)
func (hc *HealthCheck) SetVenture(venture string) {
	hc.fieldsMtx.Lock()
	hc.venture = venture
	hc.fieldsMtx.Unlock()
}

// SelfCheck returns TRUE if everything OK, otherwise it returns false
// context can be nil (it must be not nil only in case of custom values in context usage)
func (hc *HealthCheck) SelfCheck(ctx context.Context) bool {
	return !hc.isSomethingWrong(ctx, nil)
}

func (hc *HealthCheck) isSomethingWrong(ctx context.Context, hcResult map[string]interface{}) bool {
	var somethingWrong = false
	if ctx == nil {
		ctx = context.Background()
	}

	hc.checkersMtx.RLock()
	chkrsLen := len(hc.checkers)
	if chkrsLen > 0 {
		var wg sync.WaitGroup
		var results = make(chan result, chkrsLen)
		var ctx, cancel = context.WithTimeout(ctx, handlerTimeout)
		defer cancel()

		wg.Add(chkrsLen)
		for name, checker := range hc.checkers {
			go func(key string, chkr func(context.Context) interface{}) {
				results <- result{
					Key:   key,
					Value: chkr(ctx),
				}
				wg.Done()
			}(name, checker)
		}
		hc.checkersMtx.RUnlock()

		wg.Wait()

		counter := 0
		for data := range results {
			if hcResult != nil {
				hcResult[data.Key] = data.Value
			}
			if data.Value == ServiceResourceStatusError {
				somethingWrong = true
			}

			counter++
			if counter == chkrsLen {
				close(results)
			}
		}
	}
	if chkrsLen == 0 {
		hc.checkersMtx.RUnlock()
	}

	if somethingWrong {
		return somethingWrong
	}

	if hc.venture == "" || (hc.appConfigRequired && !hc.appConfigReady) {
		somethingWrong = true
	}

	return somethingWrong
}

func (hc *HealthCheck) validate(key string, f interface{}) error {
	if f == nil {
		return ErrInvalidFunc
	}
	if isKeyDeprecated(key) {
		return ErrKeyDeprecated
	}
	if _, exist := hc.checkers[key]; exist {
		return ErrKeyAlreadyExist
	}

	return nil
}

type result struct {
	Key   string
	Value interface{}
}

func (hc *HealthCheck) handleHealthCheckRequest(w http.ResponseWriter, req *http.Request) {
	hc.fieldsMtx.RLock()
	defer hc.fieldsMtx.RUnlock()

	var hcResult = map[string]interface{}{
		"service":          hc.service,
		"venture":          hc.venture,
		"version":          hc.version,
		"build_date":       hc.buildDate,
		"git_describe":     hc.gitDescribe,
		"app_config_ready": hc.appConfigReady,
		"app_config_path":  hc.appConfigPath,
		"go_version":       hc.goVersion,
	}

	var somethingWrong = hc.isSomethingWrong(req.Context(), hcResult)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	if somethingWrong {
		hcResult["status"] = ServiceStatusError
	} else {
		hcResult["status"] = ServiceStatusOK
	}

	b, err := json.MarshalIndent(hcResult, "", "    ")
	if err != nil {
		failoverResult := internalErrorResponse{
			Service:     hc.service,
			Venture:     hc.venture,
			Version:     hc.version,
			GoVersion:   hc.goVersion,
			GitDescribe: hc.gitDescribe,
			BuildDate:   hc.buildDate,

			Status:      ServiceStatusError,
			ErrorDetail: err.Error(),
		}
		b, _ = json.MarshalIndent(failoverResult, "", "    ")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(b)
		return
	}

	if hcResult["status"] == ServiceStatusOK {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.Write(b)
}
