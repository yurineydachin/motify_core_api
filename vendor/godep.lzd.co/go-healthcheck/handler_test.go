package healthcheck

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type hcResultStructure struct {
	Status         string `json:"status"`
	Service        string `json:"service"`
	Venture        string `json:"venture"`
	Version        string `json:"version"`
	BuildDate      string `json:"build_date"`
	GitDescribe    string `json:"git_describe"`
	AppConfigPath  string `json:"app_config_path"`
	AppConfigReady bool   `json:"app_config_ready"`
	GoVersion      string `json:"go_version"`

	RabbitMQ    string `json:"rabbit_mq_status,omitempty"`
	Ceph        string `json:"ceph_status,omitempty"`
	MySQL       string `json:"mysql_status,omitempty"`
	Elastic     string `json:"elastic_search_status,omitempty"`
	Aerospike   string `json:"aerospike_status,omitempty"`
	Memcache    string `json:"memcache_status,omitempty"`
	StructCache string `json:"struct_cache_status,omitempty"`
	Nfs         string `json:"nfs_status,omitempty"`
	Custom1     string `json:"custom1,omitempty"`
	Custom2     int    `json:"custom2,omitempty"`

	ErrorDetail string `json:"error_details,omitempty"`
}

const waitForChecker = time.Millisecond * 1000

var (
	serviceID   = "test_service_name"
	venture     = "vn"
	version     = "service.version.123"
	buildDateTS = int64(1497432488)
	buildDate   = "2017-06-14 09:28:08 +00:00"
	gitDescribe = "v.123-1-gd18c24b"
	goVersion   = "1.8.3"
	configPath  = "/some/path/in/etcd"
)

func slowChecker(context.Context) error {
	time.Sleep(3 * time.Second)
	return nil
}

func fastChecker(context.Context) error {
	time.Sleep(30 * time.Millisecond)
	return nil
}
func panicChecker(context.Context) error {
	panic("test panic")
}

func TestNew_InvalidParameters_ErrorInResult(t *testing.T) {
	hc, err := New("", version, gitDescribe, goVersion, venture, buildDateTS, true)
	assert.Equal(t, ErrInvalidRequiredParam, err)
	assert.Nil(t, hc)

	hc, err = New(serviceID, "", gitDescribe, goVersion, venture, buildDateTS, true)
	assert.Equal(t, ErrInvalidRequiredParam, err)
	assert.Nil(t, hc)

	hc, err = New(serviceID, version, gitDescribe, "", venture, buildDateTS, true)
	assert.Equal(t, ErrInvalidRequiredParam, err)
	assert.Nil(t, hc)

	hc, err = New(serviceID, version, gitDescribe, goVersion, venture, 0, true)
	assert.Equal(t, ErrInvalidRequiredParam, err)
	assert.Nil(t, hc)

	hc, err = New(serviceID, version, gitDescribe, goVersion, venture, -100, true)
	assert.Equal(t, ErrInvalidRequiredParam, err)
	assert.Nil(t, hc)
}

func TestNew_NoAnyCheckerIsSet_DefaultHCDataMustBeInHandlerResponse(t *testing.T) {
	hc := NewHealthCheck(serviceID, version, gitDescribe, goVersion, venture, buildDateTS, false)
	assert.NotNil(t, hc, "Health Check is nil")

	// do not set any checker here

	res := hcResultStructure{}
	statusCode := triggerHealthCheckAndParseResponse(t, hc, &res)
	assert.Equal(t, http.StatusOK, statusCode, "Invalid status code")

	assert.Equal(t, serviceID, res.Service, "Invalid value")
	assert.Equal(t, venture, res.Venture, "Invalid value")
	assert.Equal(t, version, res.Version, "Invalid value")
	assert.Equal(t, buildDate, res.BuildDate, "Invalid value")
	assert.Equal(t, gitDescribe, res.GitDescribe, "Invalid value")
	assert.Equal(t, goVersion, res.GoVersion, "Invalid value")
	assert.False(t, res.AppConfigReady, "Invalid value")
	assert.Equal(t, "", res.AppConfigPath, "Invalid value")
}

func TestNew_SetTooLongCheckerForCriticalResource_CheckerResultShouldReturnError(t *testing.T) {
	hc, err := New(serviceID, version, gitDescribe, goVersion, venture, buildDateTS, false)
	assert.NoError(t, err)

	err = hc.SetResourceChecker(ResourceTypeMySQL, true, slowChecker)
	assert.NoError(t, err, "Error is not nil")

	time.Sleep(waitForChecker)

	res := map[string]interface{}{}
	statusCode := triggerHealthCheckAndParseResponse(t, hc, &res)
	assert.Equal(t, http.StatusInternalServerError, statusCode, "Invalid status code")

	resStatus, exist := res[ResourceTypeMySQL]
	assert.True(t, exist, "Resource status does not exist in response")
	resStatusConverted, ok := resStatus.(string)
	assert.True(t, ok, "Type assertion failed")
	assert.Equal(t, ServiceResourceStatusError, resStatusConverted, "Invalid resource status")
}

func TestNew_SetTooLongCheckerForNonCriticalResource_CheckerResultShouldReturnUnstable(t *testing.T) {
	hc, err := New(serviceID, version, gitDescribe, goVersion, venture, buildDateTS, false)
	assert.NoError(t, err)

	err = hc.SetResourceChecker(ResourceTypeElastic, false, slowChecker)
	assert.Nil(t, err, "Error")

	res := map[string]interface{}{}
	statusCode := triggerHealthCheckAndParseResponse(t, hc, &res)
	assert.Equal(t, http.StatusOK, statusCode, "Invalid status code")

	resStatus, exist := res[ResourceTypeElastic]
	assert.True(t, exist, "Resource status does not exist in response")
	assert.NotNil(t, resStatus, "Status is nil interface")
	resStatusConverted := resStatus.(string)
	assert.Equal(t, ServiceResourceStatusUnstable, resStatusConverted, "Invalid resource status")
}

func TestNew_SetFastCheckerForCriticalResource_CheckerResultShouldReturnOK(t *testing.T) {
	hc, _ := New(serviceID, version, gitDescribe, goVersion, venture, buildDateTS, false)

	err := hc.SetResourceChecker(ResourceTypeAerospike, true, fastChecker)
	assert.NoError(t, err, "Error is not nil")

	time.Sleep(waitForChecker)

	res := map[string]interface{}{}
	statusCode := triggerHealthCheckAndParseResponse(t, hc, &res)
	assert.Equal(t, http.StatusOK, statusCode, "Invalid status code")

	resStatus, exist := res[ResourceTypeAerospike]
	assert.True(t, exist, "Resource status does not exist in response")
	resStatusConverted, ok := resStatus.(string)
	assert.True(t, ok, "Type assertion failed")
	assert.Equal(t, ServiceResourceStatusOk, resStatusConverted, "Invalid resource status")
}

func TestHealthCheck_SetAppConfigInfo_CheckHCResult(t *testing.T) {
	hc, _ := New(serviceID, version, gitDescribe, goVersion, venture, buildDateTS, false)
	res := hcResultStructure{}
	statusCode := triggerHealthCheckAndParseResponse(t, hc, &res)
	assert.Equal(t, http.StatusOK, statusCode, "Invalid status code")

	assert.False(t, res.AppConfigReady, "Wrong value")
	assert.Equal(t, "", res.AppConfigPath, "Wrong value")

	// Set other values and re-check
	hc.SetAppConfigInfo(configPath, true)
	statusCode = triggerHealthCheckAndParseResponse(t, hc, &res)
	assert.Equal(t, http.StatusOK, statusCode, "Invalid status code")

	assert.True(t, res.AppConfigReady, "Wrong value")
	assert.Equal(t, configPath, res.AppConfigPath, "Wrong value")
}

func TestHealthCheck_SetResourceChecker_UseNilChecker_MustReturnError(t *testing.T) {
	hc := &HealthCheck{} // do not init internal checkers map, because it must be not used
	err := hc.SetResourceChecker(ResourceTypeStructCache, false, nil)
	assert.Equal(t, ErrInvalidFunc, err)
}

func TestHealthCheck_SetResourceChecker_UseDeprecatedResourceName_MustReturnError(t *testing.T) {
	hc := &HealthCheck{
		checkers: make(map[string]func(context.Context) interface{}),
	}
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
	for name := range deprecatedResourceTypeNames {
		err := hc.SetResourceChecker(name, false, fastChecker)
		assert.Equal(t, ErrKeyDeprecated, err)
	}

}

func TestHealthCheck_SetResourceChecker_UseSameResourseNameTwice_MustReturnError(t *testing.T) {
	hc := &HealthCheck{
		checkers: make(map[string]func(context.Context) interface{}),
	}
	err := hc.SetResourceChecker(ResourceTypeStructCache, false, fastChecker)
	assert.NoError(t, err, "Error ocured")

	err = hc.SetResourceChecker(ResourceTypeStructCache, false, fastChecker)
	assert.Equal(t, ErrKeyAlreadyExist, err)
}

func TestHealthCheck_Handler(t *testing.T) {
	hc, _ := New(serviceID, version, gitDescribe, goVersion, venture, buildDateTS, false)
	handler := hc.Handler()
	assert.NotNil(t, handler, "Nil handler")

	req := httptest.NewRequest("GET", DefaultHealthCheckPath, nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	resp := w.Result()
	body, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err, "Error is not nil")
	assert.True(t, len(body) > 0, "Empty body")
}

func TestHealthCheck_SetCustomValue(t *testing.T) {
	hc, _ := New(serviceID, version, gitDescribe, goVersion, venture, buildDateTS, false)
	key1 := "custom_key_1"
	value1 := "string"
	key2 := "custom_key_2"
	value2 := 100

	err := hc.SetCustomValue(key1, func(context.Context) interface{} {
		return value1
	})
	assert.NoError(t, err, "Error must be nil")

	err = hc.SetCustomValue(key2, func(context.Context) interface{} {
		return value2
	})
	assert.NoError(t, err, "Error must be nil")

	err = hc.SetCustomValue(key1, func(context.Context) interface{} {
		return value1
	})
	assert.Equal(t, ErrKeyAlreadyExist, err)
}

func TestHealthCheck_SetVenture(t *testing.T) {
	hc, _ := New(serviceID, version, gitDescribe, goVersion, "", buildDateTS, false)
	res := hcResultStructure{}
	statusCode := triggerHealthCheckAndParseResponse(t, hc, &res)
	assert.Equal(t, http.StatusInternalServerError, statusCode, "Invalid status code")
	assert.Equal(t, "", res.Venture, "Invalid venture value")

	hc.SetVenture("sg")
	statusCode = triggerHealthCheckAndParseResponse(t, hc, &res)
	assert.Equal(t, http.StatusOK, statusCode, "Invalid status code")
	assert.Equal(t, "sg", res.Venture, "Invalid venture value")
}

func TestHealthCheck_FailoverResponse(t *testing.T) {
	hc, _ := New(serviceID, version, gitDescribe, goVersion, venture, buildDateTS, false)
	hc.SetCustomValue("key", func(context.Context) interface{} {
		// return value that could not be marshaled
		return func() {}
	})

	res := hcResultStructure{}
	statusCode := triggerHealthCheckAndParseResponse(t, hc, &res)
	assert.Equal(t, http.StatusInternalServerError, statusCode, "Invalid status code")
	assert.Equal(t, ServiceStatusError, res.Status)
	assert.NotEmpty(t, res.ErrorDetail)
	assert.Equal(t, serviceID, res.Service)
	assert.Equal(t, version, res.Version)
	assert.Equal(t, gitDescribe, res.GitDescribe)
	assert.Equal(t, venture, res.Venture)
	assert.Equal(t, buildDate, res.BuildDate)
}

func TestHealthCheck_SetResourceChecker_PanicInChecker(t *testing.T) {
	hc, _ := New(serviceID, version, gitDescribe, goVersion, venture, buildDateTS, false)
	err := hc.SetResourceChecker(ResourceTypeRabbitMQ, true, panicChecker)
	assert.NoError(t, err)

	res := hcResultStructure{}
	statusCode := triggerHealthCheckAndParseResponse(t, hc, &res)
	assert.Equal(t, http.StatusInternalServerError, statusCode)
	assert.Equal(t, ServiceStatusError, res.Status)
	assert.Equal(t, ServiceResourceStatusError, res.RabbitMQ)
	assert.Equal(t, serviceID, res.Service)
	assert.Equal(t, version, res.Version)
	assert.Equal(t, gitDescribe, res.GitDescribe)
	assert.Equal(t, venture, res.Venture)
	assert.Equal(t, buildDate, res.BuildDate)
}

func TestHealthCheck_CustomValuesInContext(t *testing.T) {
	ctx := context.WithValue(context.Background(), "test_key", "test_value")
	hc, _ := New(serviceID, version, gitDescribe, goVersion, venture, buildDateTS, false)
	checker := func(ctx context.Context) error {
		v := ctx.Value("test_key")
		val, ok := v.(string)
		assert.True(t, ok)
		assert.Equal(t, "test_value", val)

		return nil
	}
	err := hc.SetResourceChecker(ResourceTypeCeph, true, checker)
	assert.NoError(t, err)

	assert.True(t, hc.SelfCheck(ctx))

	res := hcResultStructure{}
	req := httptest.NewRequest("GET", DefaultHealthCheckPath, nil)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	hc.HandlerFunc()(w, req)

	httpResponse := w.Result()
	body, err := ioutil.ReadAll(httpResponse.Body)
	assert.NoError(t, err, "Error is not nil")
	err = json.Unmarshal(body, &res)
	assert.NoError(t, err, "Error is not nil")

	assert.Equal(t, http.StatusOK, httpResponse.StatusCode)
	assert.Equal(t, ServiceStatusOK, res.Status)
	assert.Equal(t, ServiceResourceStatusOk, res.Ceph)
	assert.Equal(t, serviceID, res.Service)
	assert.Equal(t, version, res.Version)
	assert.Equal(t, gitDescribe, res.GitDescribe)
	assert.Equal(t, venture, res.Venture)
	assert.Equal(t, buildDate, res.BuildDate)
}

func triggerHealthCheckAndParseResponse(t *testing.T, hc *HealthCheck, buf interface{}) int {
	handlerFunc := hc.HandlerFunc()
	req := httptest.NewRequest("GET", DefaultHealthCheckPath, nil)
	w := httptest.NewRecorder()
	handlerFunc(w, req)
	httpResponse := w.Result()
	body, err := ioutil.ReadAll(httpResponse.Body)
	assert.NoError(t, err, "Error is not nil")
	err = json.Unmarshal(body, buf)
	assert.NoError(t, err, "Error is not nil")

	return httpResponse.StatusCode
}

func TestHealthCheck_ComplexTestLikeRealUsage(t *testing.T) {
	hc, err := New(serviceID, version, gitDescribe, goVersion, "", buildDateTS, true)
	assert.NoError(t, err)

	res := hcResultStructure{}
	statusCode := triggerHealthCheckAndParseResponse(t, hc, &res)
	assert.False(t, hc.SelfCheck(nil))
	assert.Equal(t, http.StatusInternalServerError, statusCode)
	assert.Equal(t, ServiceStatusError, res.Status)
	assert.Empty(t, res.ErrorDetail)
	assert.Equal(t, serviceID, res.Service)
	assert.Equal(t, version, res.Version)
	assert.Equal(t, gitDescribe, res.GitDescribe)
	assert.Equal(t, "", res.Venture)
	assert.Equal(t, buildDate, res.BuildDate)

	hc.SetVenture("vn")
	res = hcResultStructure{}
	statusCode = triggerHealthCheckAndParseResponse(t, hc, &res)
	assert.False(t, hc.SelfCheck(nil))
	assert.Equal(t, http.StatusInternalServerError, statusCode)
	assert.Equal(t, ServiceStatusError, res.Status)
	assert.Empty(t, res.ErrorDetail)
	assert.Equal(t, serviceID, res.Service)
	assert.Equal(t, version, res.Version)
	assert.Equal(t, gitDescribe, res.GitDescribe)
	assert.Equal(t, "vn", res.Venture)
	assert.Equal(t, buildDate, res.BuildDate)
	assert.Equal(t, false, res.AppConfigReady)

	hc.SetAppConfigInfo("/some/path", false)
	res = hcResultStructure{}
	statusCode = triggerHealthCheckAndParseResponse(t, hc, &res)
	assert.False(t, hc.SelfCheck(nil))
	assert.Equal(t, http.StatusInternalServerError, statusCode)
	assert.Equal(t, ServiceStatusError, res.Status)
	assert.Empty(t, res.ErrorDetail)
	assert.Equal(t, serviceID, res.Service)
	assert.Equal(t, version, res.Version)
	assert.Equal(t, gitDescribe, res.GitDescribe)
	assert.Equal(t, "vn", res.Venture)
	assert.Equal(t, buildDate, res.BuildDate)
	assert.Equal(t, false, res.AppConfigReady)
	assert.Equal(t, "/some/path", res.AppConfigPath)

	hc.SetAppConfigInfo("/some/path", true)
	res = hcResultStructure{}
	statusCode = triggerHealthCheckAndParseResponse(t, hc, &res)
	assert.True(t, hc.SelfCheck(nil))
	assert.Equal(t, http.StatusOK, statusCode)
	assert.Equal(t, ServiceStatusOK, res.Status)
	assert.Empty(t, res.ErrorDetail)
	assert.Equal(t, serviceID, res.Service)
	assert.Equal(t, version, res.Version)
	assert.Equal(t, gitDescribe, res.GitDescribe)
	assert.Equal(t, "vn", res.Venture)
	assert.Equal(t, buildDate, res.BuildDate)
	assert.Equal(t, true, res.AppConfigReady)
	assert.Equal(t, "/some/path", res.AppConfigPath)

	err = hc.SetResourceChecker(ResourceTypeMySQL, true, fastChecker)
	assert.NoError(t, err)
	err = hc.SetResourceChecker(ResourceTypeElastic, true, fastChecker)
	assert.NoError(t, err)
	err = hc.SetResourceChecker(ResourceTypeAerospike, false, fastChecker)
	assert.NoError(t, err)
	err = hc.SetResourceChecker(ResourceTypeStructCache, false, fastChecker)
	assert.NoError(t, err)

	res = hcResultStructure{}
	statusCode = triggerHealthCheckAndParseResponse(t, hc, &res)
	assert.True(t, hc.SelfCheck(nil))
	assert.Equal(t, http.StatusOK, statusCode)
	assert.Equal(t, ServiceStatusOK, res.Status)
	assert.Empty(t, res.ErrorDetail)
	assert.Equal(t, serviceID, res.Service)
	assert.Equal(t, version, res.Version)
	assert.Equal(t, gitDescribe, res.GitDescribe)
	assert.Equal(t, "vn", res.Venture)
	assert.Equal(t, buildDate, res.BuildDate)
	assert.Equal(t, ServiceResourceStatusOk, res.Aerospike)
	assert.Equal(t, ServiceResourceStatusOk, res.MySQL)
	assert.Equal(t, ServiceResourceStatusOk, res.Elastic)
	assert.Equal(t, ServiceResourceStatusOk, res.StructCache)

	err = hc.SetResourceChecker(ResourceTypeNfs, false, slowChecker)
	assert.NoError(t, err)

	res = hcResultStructure{}
	statusCode = triggerHealthCheckAndParseResponse(t, hc, &res)
	assert.True(t, hc.SelfCheck(nil))
	assert.Equal(t, http.StatusOK, statusCode)
	assert.Equal(t, ServiceStatusOK, res.Status)
	assert.Empty(t, res.ErrorDetail)
	assert.Equal(t, serviceID, res.Service)
	assert.Equal(t, version, res.Version)
	assert.Equal(t, gitDescribe, res.GitDescribe)
	assert.Equal(t, "vn", res.Venture)
	assert.Equal(t, buildDate, res.BuildDate)
	assert.Equal(t, ServiceResourceStatusOk, res.Aerospike)
	assert.Equal(t, ServiceResourceStatusOk, res.MySQL)
	assert.Equal(t, ServiceResourceStatusOk, res.Elastic)
	assert.Equal(t, ServiceResourceStatusOk, res.StructCache)
	assert.Equal(t, ServiceResourceStatusUnstable, res.Nfs)

	err = hc.SetResourceChecker(ResourceTypeMemcache, true, slowChecker)
	assert.NoError(t, err)

	res = hcResultStructure{}
	statusCode = triggerHealthCheckAndParseResponse(t, hc, &res)
	assert.False(t, hc.SelfCheck(nil))
	assert.Equal(t, http.StatusInternalServerError, statusCode)
	assert.Equal(t, ServiceStatusError, res.Status)
	assert.Empty(t, res.ErrorDetail)
	assert.Equal(t, serviceID, res.Service)
	assert.Equal(t, version, res.Version)
	assert.Equal(t, gitDescribe, res.GitDescribe)
	assert.Equal(t, "vn", res.Venture)
	assert.Equal(t, buildDate, res.BuildDate)
	assert.Equal(t, ServiceResourceStatusOk, res.Aerospike)
	assert.Equal(t, ServiceResourceStatusOk, res.MySQL)
	assert.Equal(t, ServiceResourceStatusOk, res.Elastic)
	assert.Equal(t, ServiceResourceStatusOk, res.StructCache)
	assert.Equal(t, ServiceResourceStatusUnstable, res.Nfs)
	assert.Equal(t, ServiceResourceStatusError, res.Memcache)

	err = hc.SetCustomValue("custom1", func(ctx context.Context) interface{} {
		return ctx.Value("custom1")
	})
	assert.NoError(t, err)
	err = hc.SetCustomValue("custom2", func(ctx context.Context) interface{} {
		return 2
	})
	assert.NoError(t, err)

	res = hcResultStructure{}
	ctx := context.WithValue(context.Background(), "custom1", "custom_value1")
	req := httptest.NewRequest("GET", DefaultHealthCheckPath, nil)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	hc.HandlerFunc()(w, req)

	httpResponse := w.Result()
	var body []byte
	body, err = ioutil.ReadAll(httpResponse.Body)
	assert.NoError(t, err, "Error is not nil")
	err = json.Unmarshal(body, &res)
	assert.NoError(t, err, "Error is not nil")

	assert.False(t, hc.SelfCheck(nil))
	assert.Equal(t, http.StatusInternalServerError, statusCode)
	assert.Equal(t, ServiceStatusError, res.Status)
	assert.Empty(t, res.ErrorDetail)
	assert.Equal(t, serviceID, res.Service)
	assert.Equal(t, version, res.Version)
	assert.Equal(t, gitDescribe, res.GitDescribe)
	assert.Equal(t, "vn", res.Venture)
	assert.Equal(t, buildDate, res.BuildDate)
	assert.Equal(t, ServiceResourceStatusOk, res.Aerospike)
	assert.Equal(t, ServiceResourceStatusOk, res.MySQL)
	assert.Equal(t, ServiceResourceStatusOk, res.Elastic)
	assert.Equal(t, ServiceResourceStatusOk, res.StructCache)
	assert.Equal(t, ServiceResourceStatusUnstable, res.Nfs)
	assert.Equal(t, ServiceResourceStatusError, res.Memcache)
	assert.Equal(t, "custom_value1", res.Custom1)
	assert.Equal(t, 2, res.Custom2)

	lib1 := testLib{}
	err = hc.SetResourceChecker(ResourceTypeCeph, true, &lib1)
	assert.NoError(t, err)

	res = hcResultStructure{}
	ctx = context.WithValue(context.Background(), "custom1", "custom_value1")
	req = httptest.NewRequest("GET", DefaultHealthCheckPath, nil)
	req = req.WithContext(ctx)
	w = httptest.NewRecorder()

	hc.HandlerFunc()(w, req)

	httpResponse = w.Result()
	body, err = ioutil.ReadAll(httpResponse.Body)
	assert.NoError(t, err, "Error is not nil")
	err = json.Unmarshal(body, &res)
	assert.NoError(t, err, "Error is not nil")

	assert.False(t, hc.SelfCheck(nil))
	assert.Equal(t, http.StatusInternalServerError, statusCode)
	assert.Equal(t, ServiceStatusError, res.Status)
	assert.Empty(t, res.ErrorDetail)
	assert.Equal(t, serviceID, res.Service)
	assert.Equal(t, version, res.Version)
	assert.Equal(t, gitDescribe, res.GitDescribe)
	assert.Equal(t, "vn", res.Venture)
	assert.Equal(t, buildDate, res.BuildDate)
	assert.Equal(t, ServiceResourceStatusOk, res.Aerospike)
	assert.Equal(t, ServiceResourceStatusOk, res.MySQL)
	assert.Equal(t, ServiceResourceStatusOk, res.Elastic)
	assert.Equal(t, ServiceResourceStatusOk, res.StructCache)
	assert.Equal(t, ServiceResourceStatusUnstable, res.Nfs)
	assert.Equal(t, ServiceResourceStatusError, res.Memcache)
	assert.Equal(t, ServiceResourceStatusOk, res.Ceph)
	assert.Equal(t, "custom_value1", res.Custom1)
	assert.Equal(t, 2, res.Custom2)
}

type testLib struct {
	mtx sync.Mutex
	err error
}

func (l *testLib) HealthCheck(ctx context.Context) (err error) {
	l.mtx.Lock()
	defer l.mtx.Unlock()
	return l.err
}

func TestHealthCheck_SetResourceChecker_InvalidChecker(t *testing.T) {
	var wrongCheckers []interface{}

	var ch1 interface{}
	var ch2 *HealthCheck
	wrongCheckers = append(wrongCheckers, ch1)
	wrongCheckers = append(wrongCheckers, ch2)
	wrongCheckers = append(wrongCheckers, struct{}{})
	wrongCheckers = append(wrongCheckers, HealthCheck{})
	wrongCheckers = append(wrongCheckers, 0)
	wrongCheckers = append(wrongCheckers, "abc")
	wrongCheckers = append(wrongCheckers, func() {})

	hc := HealthCheck{}
	for _, checker := range wrongCheckers {
		err := hc.SetResourceChecker(ResourceTypeRabbitMQ, false, checker)
		assert.Equal(t, ErrInvalidFunc, err)
	}

}

// Used for auto-generated documentation
func ExampleNewHealthCheck() {
	var buildDateTimestamp = int64(1497432488)
	var serviceID = "service_name_as_in_etcd"
	var serviceVersion = "version"
	var gitDescribe = "git-describe value"
	var goVersion = "1.8.3"
	var isAppConfigRequired = true
	var venture = "venture" // can be empty on start

	// Init base health check
	hc, err := New(serviceID, serviceVersion, gitDescribe, goVersion, venture, buildDateTimestamp, isAppConfigRequired)
	if err != nil {
		panic(err)
	}

	// Init checkers for external resources that your service uses.
	mySQLChecker := func(context.Context) error {
		// do some real work here
		return nil
	}
	aerospikeChecker := func(context.Context) error {
		// do some real work here
		return nil
	}

	// Set MySQL checker func. And this resource is critical (service can't work if resource is down)
	if err := hc.SetResourceChecker(ResourceTypeMySQL, true, mySQLChecker); err != nil {
		panic(err)
	}
	// Set Aerospike checker func. And this resource is not critical (service can work without it)
	if err := hc.SetResourceChecker(ResourceTypeAerospike, false, aerospikeChecker); err != nil {
		panic(err)
	}

	// If your service already can work with configuration stored in ETCD
	// you need to set current configuration path in ETCD and flag if ETCD config is enabled
	hc.SetAppConfigInfo("some/path/in/etcd", true)

	// If your service gets venture from ETCD and you don't know what venture used on service start
	// you can set venture value later
	hc.SetVenture("vn")

	// If you wish to set custom key/value in health check response
	hc.SetCustomValue("custom_key_in_health_check", func(ctx context.Context) interface{} {
		// do any logic here or just return data
		// for example, you can get data from context (if you set some custom data in context in middleware)
		// check TestHealthCheck_CustomValuesInContext
		return ctx.Value("some_key")
	})

	// start to serve it
	http.Handle(DefaultHealthCheckPath, hc.Handler())
	// OR
	//http.HandleFunc(DefaultHealthCheckPath, hc.HandlerFunc())

	http.ListenAndServe(":8080", nil)

	// If you run this example and open http://127.0.0.1:8080/health_check
	// it returns JSON
	/*
		{
		  "aerospike_status": "Ok",
		  "app_config_path": "some/path/in/etcd",
		  "app_config_ready": true,
		  "build_date": "build-date-is-here",
		  "git_describe": "gi-describe value",
		  "go_version": "1.8.3",
		  "mysql_status": "Ok",
		  "service": "service_name_as_in_etcd",
		  "status": "Ok",
		  "venture": "venture",
		  "version": "vn",
		  "custom_key_in_health_check": "some_value",
		}
	*/
}
