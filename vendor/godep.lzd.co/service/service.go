package service

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"go/format"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"godep.lzd.co/discovery/balancer"
	"godep.lzd.co/discovery/provider"
	"godep.lzd.co/discovery/provider/etcdV3"
	"godep.lzd.co/discovery/registrator"
	"godep.lzd.co/etcd"
	"godep.lzd.co/go-healthcheck"
	golog "godep.lzd.co/go-log"
	"godep.lzd.co/go-trace"
	"godep.lzd.co/metrics"
	"godep.lzd.co/metrics/httpmon"
	"godep.lzd.co/service/admin"
	"godep.lzd.co/service/cache/inmem"
	"godep.lzd.co/service/closer"
	"godep.lzd.co/service/config"
	"godep.lzd.co/service/dconfig"
	"godep.lzd.co/service/handlersmanager"
	"godep.lzd.co/service/interfaces"
	"godep.lzd.co/service/k8s"
	"godep.lzd.co/service/logger"
	"godep.lzd.co/service/response"
	"godep.lzd.co/service/utils"
	"godep.lzd.co/swgui"

	etcdcl "github.com/coreos/etcd/client"
	etcdcl3 "github.com/coreos/etcd/clientv3"
	"github.com/davecgh/go-spew/spew"
	"github.com/facebookgo/grace/gracehttp"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sergei-svistunov/gorpc"
	"github.com/sergei-svistunov/gorpc/transport/cache"
	"github.com/sergei-svistunov/gorpc/transport/http_json"
	"github.com/sergei-svistunov/gorpc/transport/http_json/adapter"
	"google.golang.org/grpc"
)

// TODO: add real version set by compiler?
var (
	AppVersion  string // this value set by compiler
	GoVersion   string // this value set by compiler
	BuildDate   string // this value set by compiler
	GitRev      string // this value set by compiler
	GitHash     string // this value set by compiler
	GitDescribe string // this value set by compiler
)

const (
	loggerFlushTimout       = time.Second * 3
	handlerNameCtxKey       = "handlerName"
	appNameUnknown          = "unknown"
	defaultLoggerLevel      = "ERROR"
	waitForETCDRetryTimeout = time.Second * 3
)

type Service struct {
	id                                  string
	venture                             string
	handlersPath                        string
	options                             Options
	httpMux                             *http.ServeMux
	DconfManager                        *dconfig.Manager
	transportCache                      cache.ICache
	hm                                  interfaces.IHandlersManager
	resources                           []interfaces.IResource
	proxyClient                         *http.Client
	headers                             map[string]string
	httpJSONHandler                     http.Handler
	httpMiddleware                      func(http.Handler) http.Handler
	apiHandlerCallbacks                 http_json.APIHandlerCallbacks
	swaggerJSONCallbacks                http_json.SwaggerJSONCallbacks
	swaggerJSONCustomHandlerConstructor func(resources []interfaces.IResource, venture string) http.Handler
	swaggerDocsPostProcess              func([]byte) []byte
	swaggerUIHandler                    http.Handler
	swaggerJSONHandler                  http.Handler
	clientGenHandler                    http.HandlerFunc
	gorpcHM                             *gorpc.HandlersManager
	grpcServer                          *grpc.Server
	etcdClientV3                        *etcdcl3.Client
	etcdClientV2                        etcdcl.Client
	etcdRegistrator                     registrator.IRegistrator
	etcdProvider                        provider.IProvider
	hc                                  *healthcheck.HealthCheck
	hcMtx                               sync.Mutex
	hcCheckers                          []HealthChecker
	baseHttpServer                      *http.Server // while service is not initialized it implements health check handler
}

type Options struct {
	HM                                  interfaces.IHandlersManager
	ProxyClient                         *http.Client
	TransportCache                      cache.ICache
	Headers                             map[string]string
	APIHandlerCallbacks                 http_json.APIHandlerCallbacks
	SwaggerJSONCallbacks                http_json.SwaggerJSONCallbacks
	SwaggerJSONCustomHandlerConstructor func(resources []interfaces.IResource, venture string) http.Handler
	SwaggerDocsPostProcess              func([]byte) []byte

	// @deprecated and not used anymore
	HealthCheckFunc func() map[string]interface{}
}

func (service *Service) SetMiddleware(h func(http.Handler) http.Handler) {
	service.httpMiddleware = h
}

// SetGrpcServer set a gRPC server for service object
func (service *Service) SetGrpcServer(server *grpc.Server) {
	service.grpcServer = server
}

// SetDocumentJSONHandler set handler for json document request
func (service *Service) SetDocumentJSONHandler(h http.Handler) {
	service.swaggerJSONHandler = h
}

// SetDocumentUIHandler set handler for document UI
func (service *Service) SetDocumentUIHandler(h http.Handler) {
	service.swaggerUIHandler = h
}

func init() {
	config.RegisterBool("version", "display binary version and info and exit", false)
	config.RegisterBool("print-debug", "Print debug output on internal server error", false)

	config.RegisterString("venture", "Current venture", "UNKNOWN")
	config.RegisterString("env", "Current envirement", "dev")
	config.RegisterString("rollout_type", "Type of the instance in terms of progressive rollout, like 'stable', 'unstable1', ..., 'unstableN'", "stable")

	config.RegisterString("etcd-namespace", "Current namespace", "motify_api")
	config.RegisterString("etcd-endpoints", "List of etcd hosts joined by ,", "http://localhost:4001")
	config.RegisterBool("etcd-registration-enabled", "Enable registration service in etcd", true)
	config.RegisterBool("write-access-log", "Write access log to stdout", false)

	config.RegisterString("addr", "<hostname/ip>:<port> for handlers", ":8080")
	config.RegisterString("adm-addr", "<hostname/ip>:<port> for administrative page", ":9080")

	// ## Parameters needed for running in docker container ##
	// https://confluence.lazada.com/display/RE/Requirements+to+Docker-based+deploy+process+for+GO+components
	config.RegisterUint("port", "The parameter of port for main HTTP listener. Numeric type between the range of 1024 and 32767. Overrides the 'addr' param", 0)
	config.RegisterUint("admin-port", "The parameter of port for listener administrative tools. Numeric type between the range of 1024 and 32767. Overrides the 'addr' param", 0)
	config.RegisterUint("profile-port", "Not used", 0)
	config.RegisterUint("grpc-port", "port number for gRPC services", 0)
	// #######################################################

	config.RegisterString("syslog_addr_type", "Syslog address type (unixgram, tcd, udp). StdOut lgging by default if empty", "")
	config.RegisterString("syslog_addr", "Syslog address. StdOut logging by default if empty", "")

	config.RegisterUint("ssl-port", "SSL port for handlers", 0)
	config.RegisterString("ssl-certificate", "Specifies a file with the certificate in the PEM format for the given virtual server", "")
	config.RegisterString("ssl-certificate-key", "Specifies a file with the secret key in the PEM format for the given virtual server", "")
	config.RegisterString("ssl-client-certificate", "Specifies a file with trusted CA certificates in the PEM format used to verify client certificates", "")

	config.RegisterString("multicast-addr", "<ip>:<port> for multicast", "224.0.0.1")
	config.RegisterString("discovery-nodes", "List of service host names joined by ','", ``)

	config.RegisterString("docs-certificates", "List of certificates allowed to use /docs. Separator is ','", "")
	config.RegisterBool("docs-on-public-port", "Documentation also available on public port if true", false)

	config.RegisterString("proxy-url", "Proxy for external http requests", "")
	config.RegisterBool("inmem-cache-enabled", "Use inmem cache", true)
}

func New(id string, handlersPath string) *Service {
	return NewWithOpts(id, handlersPath, Options{})
}

func NewWithOpts(id string, handlersPath string, opts Options) *Service {
	service := &Service{
		id:           id,
		handlersPath: handlersPath,
		options:      opts,
	}
	service.parseFlags()
	prepareLogger(id)
	service.SetOptions(opts)

	logger.Notice(nil, "app configuration: %s", config.String())
	return service
}

func (service *Service) SetOptions(opts Options) {
	service.options = opts

	service.swaggerJSONCallbacks = opts.SwaggerJSONCallbacks
	service.swaggerJSONCustomHandlerConstructor = opts.SwaggerJSONCustomHandlerConstructor
	service.apiHandlerCallbacks = opts.APIHandlerCallbacks
	service.swaggerDocsPostProcess = opts.SwaggerDocsPostProcess

	if opts.Headers != nil {
		service.headers = opts.Headers
	}
	if service.headers == nil {
		service.headers = map[string]string{}
	}

	if opts.ProxyClient != nil {
		service.proxyClient = opts.ProxyClient
	}
	var err error
	if service.proxyClient == nil {
		service.proxyClient, err = NewProxyClientFromFlags()
		if err != nil {
			logger.Critical(nil, "failed to create proxy client: %v", err)

			var ctx, cancel = context.WithTimeout(context.Background(), loggerFlushTimout)
			if loggerErr := logger.Flush(ctx); loggerErr != nil {
				fmt.Printf("Service '%s' has stopped with error: %s\nLogger was unable to flush last messages with error: %s", service.id, err, loggerErr)
			}
			cancel()

			os.Exit(1)
		}
	}

	if opts.TransportCache != nil {
		service.transportCache = opts.TransportCache
	}
	if service.transportCache == nil {
		service.transportCache = inmem.NewFromFlags("")
	}

	if opts.HM != nil {
		service.hm = opts.HM
	}
}

func (service *Service) parseFlags() {
	if err := config.Parse(); err != nil {
		logger.Critical(nil, err.Error())

		var ctx, cancel = context.WithTimeout(context.Background(), loggerFlushTimout)
		if loggerErr := logger.Flush(ctx); loggerErr != nil {
			fmt.Printf("Service '%s' has stopped with error: %s\nLogger was unable to flush last messages with error: %s", service.id, err, loggerErr)
		}
		cancel()

		os.Exit(1)
	}

	if displayVersion, _ := config.GetBool("version"); displayVersion {
		fmt.Printf("%s version %s\n", service.id, AppVersion)
		fmt.Printf("built at %s with compiler %s\n", BuildDate, GoVersion)
		fmt.Printf("from git commit %s\n", GitRev)
		os.Exit(0)
	}

	if err := config.ParseAll(); err != nil {
		logger.Warning(nil, err.Error())
	}
}

func prepareLogger(serviceName string) {
	syslogAddrType, _ := config.GetString("syslog_addr_type")
	syslogAddr, _ := config.GetString("syslog_addr")

	if err := logger.Init(serviceName, syslogAddrType, syslogAddr, defaultLoggerLevel); err != nil {
		panic(err)
	}
}

type RegistratorCancelFunc func()

func (service *Service) MustRegisterHandlers(handlers ...gorpc.IHandler) {
	if err := service.RegisterHandlers(handlers...); err != nil {
		panic(err)
	}
}

func (service *Service) RegisterHandlers(handlers ...gorpc.IHandler) error {
	for _, handler := range handlers {
		if err := service.RegisterHandler(handler); err != nil {
			return err
		}
	}
	return nil
}

func (service *Service) RegisterHandler(handler gorpc.IHandler) error {
	return service.hm.RegisterHandler(handler)
}

func (service *Service) RegisterResources(resources ...interfaces.IResource) {
	service.resources = append(service.resources, resources...)
}

func (service *Service) RegisterResource(resource interfaces.IResource) {
	service.resources = append(service.resources, resource)
}

func getAddr() (string, string, uint64) {
	// "port" param should override other settings
	dockerPort, _ := config.GetUint("port")
	if dockerPort != 0 {
		strPort := strconv.FormatUint(uint64(dockerPort), 10)
		return ":" + strPort, "", uint64(dockerPort)
	}
	addr, _ := config.GetString("addr")
	hostname, strPort, err := net.SplitHostPort(addr)
	if err != nil {
		panic(err)
	}
	port, err := strconv.ParseUint(strPort, 10, 16)
	if err != nil {
		panic(err)
	}

	return addr, hostname, port
}

func getAdmAddr(port uint64) (string, uint64) {
	// "admin-port" param should override other settings
	dockerPort, _ := config.GetUint("admin-port")
	if dockerPort != 0 {
		strPort := strconv.FormatUint(uint64(dockerPort), 10)
		return ":" + strPort, uint64(dockerPort)
	}

	admAddr, _ := config.GetString("adm-addr")

	strAdmHost, strAdmPort, err := net.SplitHostPort(admAddr)
	if err != nil {
		panic(err)
	}

	var admPort uint64
	if strAdmPort == "" {
		admPort = port + 2
		strAdmPort = strconv.FormatUint(admPort, 10)
		admAddr = strAdmHost + ":" + strAdmPort
	} else {
		admPort, err = strconv.ParseUint(strAdmPort, 10, 16)
		if err != nil {
			panic(err)
		}
	}

	return admAddr, admPort
}

func getProfileAddr() (string, uint64) {
	// "profile-port" param should override other settings
	dockerPort, _ := config.GetUint("profile-port")
	strPort := strconv.FormatUint(uint64(dockerPort), 10)
	return ":" + strPort, uint64(dockerPort)
}

//getGrpcAddr return address for gRPC server to listening
//parameter is restful port - main port of service
//we follow the rule of port convention that defined by:
//https://confluence.lazada.com/display/DEV/Services+Reference+Documentation#ServicesReferenceDocumentation-Portassignmentconventionsforapplications
func getGrpcAddr(port uint64) (string, uint64) {
	grpcPort, _ := config.GetUint("grpc-port")
	if grpcPort > 0 {
		return ":" + strconv.FormatUint(uint64(grpcPort), 10), uint64(grpcPort)
	}
	return ":" + strconv.FormatUint(uint64(port+6), 10), uint64(port) + 6
}

func (service *Service) Init() error {
	addr, hostname, port := getAddr()
	_, admPort := getAdmAddr(port)
	_, profilePort := getProfileAddr()
	if profilePort == 0 {
		profilePort = admPort + 2
	}

	service.venture, _ = config.GetString("venture")
	_, grpcPort := getGrpcAddr(port)
	env, _ := config.GetString("env")
	rolloutType, _ := config.GetString("rollout_type")

	if hostname == "" || hostname == "0.0.0.0" {
		var err error
		hostname, err = k8s.GetHostname()
		if err != nil {
			logger.Critical(nil, "could not get hostname: %v", err)
			return err
		}
	}
	opentracing.SetGlobalTracer(
		gotrace.NewTracer(
			gotrace.WithAppEnv(service.id, AppVersion, hostname),
			gotrace.WithSpanRecorder(
				gotrace.NewRecorder(
					gotrace.WithLogCollector(
						golog.NewSpanCollector(
							logger.GetLoggerInstance().Writer()))))))

	service.initHealthCheck(addr, rolloutType)

	etcdReady, err := service.initDiscoveryRegistrator(service.venture, env, hostname, rolloutType, int(port), int(admPort), int(profilePort), int(grpcPort))
	if err != nil {
		logger.Critical(nil, "Failed to initialize discovery registrator: %v", err)
		return err
	}

	// wait for ETCD
	<-etcdReady

	if service.hm == nil {
		service.hm = handlersmanager.New(service.handlersPath)
	}

	if service.DconfManager == nil {
		service.DconfManager = dconfig.NewManager(service.id)
	}

	dconfig.RegisterString("log-level", "Logging level (DEBUG, INFO, NOTICE, WARNING, ERROR, CRITICAL, ALERT, EMERGENCY), case-insensitive", defaultLoggerLevel,
		func(val string) {
			if ok := logger.ParseAndSetLevel(val); !ok {
				logger.Error(nil, "Could not parse and set log level %s", val)
			}
		})
	level, _ := dconfig.GetString("log-level")
	logger.ParseAndSetLevel(level)

	dconfig.RegisterBool("print-debug", "Enable output debug info in a response", false,
		func(val bool) {
			http_json.PrintDebug = val
		})

	dconfig.RegisterDuration("handler_timeout", "Timeout for processing requests (0 - unlimited)", 0, func(timeout time.Duration) {
		if h, ok := service.httpJSONHandler.(*http_json.APIHandler); ok {
			h.SetTimeout(timeout)
			logger.Debug(nil, "handler_timeout changed, new value: %s", timeout)
		}
	})

	timeout, _ := dconfig.GetDuration("handler_timeout")
	gorpcHM := service.hm.GetGoRPCHandlersManager()
	service.gorpcHM = gorpcHM
	service.httpJSONHandler = http_json.NewAPIHandler(
		gorpcHM, service.transportCache, service.getApiHandlerCallbacks()).SetTimeout(timeout)

	var (
		docsUIHandler http.Handler
		swaggerPort   uint64
	)
	if port != admPort {
		swaggerPort = port
	}
	if service.swaggerUIHandler == nil {
		service.swaggerUIHandler = admin.NewDocsHandler(service.swaggerDocsPostProcess)
		docsUIHandler = http.StripPrefix("/docs", service.swaggerUIHandler)
	} else if _, ok := service.swaggerUIHandler.(*swgui.Handler); ok {
		docsUIHandler = service.swaggerUIHandler
	}

	if service.swaggerJSONHandler == nil {
		service.swaggerJSONHandler = http_json.NewSwaggerJSONHandler(gorpcHM, uint16(swaggerPort), service.swaggerJSONCallbacks)
	}

	service.clientGenHandler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		code, err := service.GenerateClientLib()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Write(code)
	})

	mux := http.NewServeMux()

	// TODO: shouldn't it all be hidden behind certs?
	mux.Handle("/jsonrpc", &jsonRPCHandler{service.id, hostname, service.httpJSONHandler})

	if enabled, _ := config.GetBool("docs-on-public-port"); enabled {
		mux.Handle("/swagger.json", service.swaggerJSONHandler)
		mux.Handle("/docs/", docsUIHandler)
		mux.Handle("/client.go", service.clientGenHandler)
	} else if certs, ok := config.GetStringSlice("docs-certificates"); ok {
		mux.Handle("/swagger.json", admin.CheckCertificate(certs, service.swaggerJSONHandler))
		mux.Handle("/docs/", admin.CheckCertificate(certs, docsUIHandler))
		mux.Handle("/client.go", admin.CheckCertificate(certs, service.clientGenHandler))
	}

	mux.HandleFunc("/rev.txt", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		fmt.Fprintln(w, AppVersion)
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		admin.RPSCounter.Inc(1)
		switch r.URL.Path {
		case "", "/", "/health_check", "/health_check/":
			service.hc.HandlerFunc()(w, r)
		default:
			service.httpJSONHandler.ServeHTTP(w, r)
		}
	})
	service.httpMux = mux

	// TODO(axel) replace it with metrics call
	// TODO: send it in goroutine in case if gracehttp can't be started
	// countServerStarted := metrics.GetOrRegisterCounter("total.service_started", service.monitoring.GetMetricsRegistry())
	// countServerStarted.Inc(1)
	return nil
}

func (service *Service) initHealthCheck(addr, rolloutType string) error {
	buildDate, err := time.Parse("2006-01-02 15:04:05", BuildDate)
	if err != nil {
		logger.Critical(nil, "Build date string '%s' parsing error. Please check your Makefile: %v", BuildDate, err)
		return err
	}
	service.hc = healthcheck.NewHealthCheck(service.id, AppVersion, GitDescribe, GoVersion, service.venture, buildDate.Unix(), false)
	structCacheChecker := func(ctx context.Context) error {
		key := []byte("health_check_test")
		value := &cache.CacheEntry{
			Content:           []byte("content"),
			CompressedContent: []byte("compressed_content"),
			Hash:              "hash",
			Body:              200,
		}
		service.transportCache.Put(key, value)
		result := service.transportCache.Get(key)

		var err error
		if result == nil {
			err = errors.New("nil result")
		}
		if len(result.Content) != len(value.Content) ||
			len(result.CompressedContent) != len(value.CompressedContent) ||
			result.Hash != value.Hash ||
			result.Body != value.Body {
			err = errors.New("Invalid data in cache")
		}

		return err
	}
	service.hc.SetResourceChecker(healthcheck.ResourceTypeStructCache, false, structCacheChecker)

	etcdChecker := func(ctx context.Context) error {
		var err = errors.New("ETCD error")
		if service.etcdClientV3 == nil {
			return err
		}
		if service.etcdProvider == nil {
			return err
		}
		if service.etcdRegistrator == nil {
			return err
		}

		f := provider.KeyFilter{
			Namespace:   provider.NamespaceDiscovery,
			Type:        provider.TypeApp,
			Name:        service.id,
			RolloutType: rolloutType,
			Owner:       provider.DefaultOwner,
			ClusterType: provider.DefaultClusterType,
		}

		if _, err := service.etcdProvider.Get(ctx, f); err != nil {
			return err
		}

		return nil
	}
	service.hc.SetResourceChecker(healthcheck.ResourceTypeEtcd, true, etcdChecker)

	for _, checker := range service.hcCheckers {
		if err := service.hc.SetResourceChecker(checker.ResourceType, checker.IsCritical, checker.Checker); err != nil {
			logger.Error(nil, "Can't set resource checker: %s", err)
		}
	}

	mux := http.NewServeMux()
	mux.Handle(healthcheck.DefaultHealthCheckPath, service.Healthcheck())
	service.baseHttpServer = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	go func() {
		if err := service.baseHttpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Critical(nil, err.Error())
			logger.Flush(nil)
			os.Exit(1)
		}
	}()

	return nil
}

func (service *Service) initDiscoveryRegistrator(venture, env, host, rolloutType string, port, adminPort, profilePort, grpcPort int) (<-chan struct{}, error) {
	log := utils.DiscoveryLogger{}
	etcdEndpoints, _ := config.GetStringSlice("etcd-endpoints")
	if len(etcdEndpoints) == 0 {
		return nil, errors.New("It seems 'etcd-endpoints' is empty")
	}

	etcdReady := make(chan struct{}, 1)

	initETCD := func() error {
		logger.Info(nil, "Try to init ETCD")
		var err error

		// init ETCD v2 client
		service.etcdClientV2, err = etcd.NewClient(etcdEndpoints)
		if err != nil {
			return err
		}

		// init ETCD v3 client
		service.etcdClientV3, err = etcd.NewClientV3(utils.GetGrpcEndpoints(etcdEndpoints))
		if err != nil {
			return err
		}

		service.etcdProvider = etcdV3.NewProvider(service.etcdClientV3, log)
		if service.etcdProvider == nil {
			return errors.New("NewProvider() failed")
		}
		info := registrator.AppRegistrationParams{
			ServiceName: service.id,
			RolloutType: rolloutType,
			Host:        host,
			HTTPPort:    port,
			// Admin info
			AdminPort: adminPort,
			Version: registrator.VersionInfo{
				AppVersion: AppVersion,
			},
			// Monitoring info
			MonitoringPort: profilePort,
			Venture:        venture,
			Environment:    env,
		}

		//only register gRPC service if port is valid and grpc server exist
		if grpcPort > 0 && service.grpcServer != nil {
			info.GRPCPort = grpcPort
		}

		service.etcdRegistrator, err = registrator.NewV2V3(info, service.etcdProvider, service.etcdClientV2, log)

		if err != nil {
			return fmt.Errorf("Failed to create etcd registrator: %v", err)
		}

		etcdReady <- struct{}{}
		return nil
	}

	go func() {
		for {
			if err := initETCD(); err != nil {
				logger.Critical(nil, "ETCD initialization failed: %s", err.Error())
				time.Sleep(waitForETCDRetryTimeout)
				continue
			}
			return
		}
	}()

	return etcdReady, nil
}

func (service *Service) Run() error {
	var startedAt = time.Now()
	logger.Alert(nil, "Start service version `%s`", AppVersion)

	err := service.run()

	var ctx, cancel = context.WithTimeout(context.Background(), loggerFlushTimout)
	defer cancel()

	if err != nil {
		logger.Emergency(nil, "Service version `%s` run failed: %s. Service uptime: %s", AppVersion, err, time.Since(startedAt))
	} else {
		logger.Alert(nil, "Service version `%s` terminated gracefully. Service uptime: %s", AppVersion, time.Since(startedAt))
	}

	if loggerErr := logger.Flush(ctx); loggerErr != nil {
		fmt.Printf("Service '%s' has stopped with error: %s\nLogger was unable to flush last messages with error: %s", service.id, err, loggerErr)
	}
	return err
}

func (service *Service) run() error {
	err := service.Init()
	if err != nil {
		return err
	}

	addr, _, port := getAddr()
	admAddr, admPort := getAdmAddr(port)
	profileAddr, _ := getProfileAddr()
	env, _ := config.GetString("env")
	ns, _ := config.GetString("etcd-namespace")
	multicastAddr, _ := config.GetString("multicast-addr")

	sslPort, _ := config.GetUint("ssl-port")
	sslCertificate, _ := config.GetString("ssl-certificate")
	sslCertificateKey, _ := config.GetString("ssl-certificate-key")
	sslClientCertificate, _ := config.GetString("ssl-client-certificate")

	logger.Info(nil, "Service %s is starting; venture: %s, env: %s, version: %s, commit: %s", service.id, service.venture, env, AppVersion, GitRev)

	if !strings.Contains(multicastAddr, ":") {
		multicastAddr += ":" + fmt.Sprintf("%d", admPort)
	}

	service.DconfManager.Run(service.etcdClientV2, ns, service.venture, env)

	profileMux := http.NewServeMux()

	if registrationEnabled, _ := config.GetBool("etcd-registration-enabled"); registrationEnabled {
		if err = service.etcdRegistrator.Register(); err != nil {
			return fmt.Errorf("Cannot register service to ETCD discovery: %s", err)
		}

		if err = service.etcdRegistrator.EnableDiscovery(); err != nil {
			return fmt.Errorf("Cannot enable dicovery ability: %s", err)
		}
		//cancelRegister := service.register(ns, venture, env, fmt.Sprintf("%d", port), fmt.Sprintf("%d", profilePort))
		closer.Add(func() {
			service.etcdRegistrator.Unregister()
		})
		profileMux.HandleFunc("/registerstop", func(w http.ResponseWriter, req *http.Request) {
			service.etcdRegistrator.DisableDiscovery()
		})
	}

	profileMux.Handle("/metrics", promhttp.Handler())

	defer func() {
		closer.CloseAll()
		closer.Wait()
	}()

	var h, ah, ph http.Handler

	h = service.httpMux
	if service.httpMiddleware != nil {
		h = service.httpMiddleware(h)
	}
	if service.swaggerJSONCustomHandlerConstructor != nil {
		service.swaggerJSONHandler = service.swaggerJSONCustomHandlerConstructor(service.resources, service.venture)
	}
	ah = admin.NewHTTPHandler(service.id, AppVersion, service.venture, env, service.resources, service.DconfManager, service.swaggerJSONCallbacks, service.swaggerUIHandler, service.swaggerJSONHandler, service.clientGenHandler)
	ph = profileMux

	if al, _ := config.GetBool("write-access-log"); al {
		logger.Warning(nil, "Access logs are always enabled, 'write-access-log' parameter is deprecated")
	}

	servers := []*http.Server{
		{
			Addr:    addr,
			Handler: h,
		},
		{
			Addr:    admAddr,
			Handler: ah,
		},
		{
			Addr:    profileAddr,
			Handler: ph,
		},
	}

	if sslPort != 0 && sslCertificate != "" && sslCertificateKey != "" {
		cert, err := tls.LoadX509KeyPair(sslCertificate, sslCertificateKey)
		if err != nil {
			log.Fatal(err)
		}
		tlsConfig := &tls.Config{
			MinVersion:               tls.VersionTLS10,
			ClientAuth:               tls.NoClientCert,
			PreferServerCipherSuites: true,
			Certificates:             []tls.Certificate{cert},
		}
		if sslClientCertificate != "" {
			caCert, err := ioutil.ReadFile(sslClientCertificate)
			if err != nil {
				log.Fatal(err)
			}
			caCertPool := x509.NewCertPool()
			caCertPool.AppendCertsFromPEM(caCert)

			tlsConfig.ClientCAs = caCertPool
			tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
		}
		servers = append(servers, &http.Server{
			Addr:      fmt.Sprintf(":%d", sslPort),
			Handler:   h,
			TLSConfig: tlsConfig,
		})
	}
	//start gRPC if it was set
	if service.grpcServer != nil {
		grpcAddr, grpcPort := getGrpcAddr(port)
		lis, err := net.Listen("tcp", grpcAddr)
		if err != nil {
			return fmt.Errorf("failed to listen on gRPC addresss (%d): %s", grpcPort, err)
		}
		go func() {
			logger.Critical(nil, "gRPC Server is serving on port: %d", grpcPort)
			service.grpcServer.Serve(lis)
		}()
		closer.Add(func() {
			logger.Critical(nil, "gRPC Server (on port %d) is stopped", grpcPort)
			service.grpcServer.Stop()
		})
	}

	// stop base server that started on init (for health check handling)
	if err := service.baseHttpServer.Shutdown(context.Background()); err != nil {
		logger.Critical(nil, err.Error())
	}
	// fix 'port already in use' error
	time.Sleep(time.Millisecond * 50)

	logger.Alert(nil, "Service version `%s` started", AppVersion)
	return gracehttp.Serve(servers...)
}

func (service *Service) getApiHandlerCallbacks() http_json.APIHandlerCallbacks {
	return http_json.APIHandlerCallbacks{
		OnStartServing: func(ctx context.Context, req *http.Request) {
			if service.apiHandlerCallbacks.OnStartServing != nil {
				service.apiHandlerCallbacks.OnStartServing(ctx, req)
			}
		},
		OnEndServing: func(ctx context.Context, req *http.Request, startTime time.Time) {
			span := opentracing.SpanFromContext(ctx)
			if span == nil {
				return
			}
			code := "200"
			if r, ok := response.FromContext(ctx); ok {
				code = strconv.Itoa(r.StatusCode)
			}
			appName := appNameUnknown
			if sc, ok := span.Context().(gotrace.Span); ok {
				appName = sc.ParentAppInfo.AppName
			}
			var (
				labels = map[string]string{
					"code":        code,
					"handler":     metrics.MakePathTag(req.URL.Path),
					"client_name": appName,
				}
				duration = metrics.SinceMs(startTime)
			)
			httpmon.ResponseTime.With(labels).Observe(duration)
			httpmon.ResponseTimeSummary.With(labels).Observe(duration)

			span.SetTag(string(ext.HTTPMethod), req.Method)
			span.SetTag(string(ext.HTTPStatusCode), code)

			span.Finish()

			if service.apiHandlerCallbacks.OnEndServing != nil {
				service.apiHandlerCallbacks.OnEndServing(ctx, req, startTime)
			}
		},
		OnInitCtx: func(ctx context.Context, req *http.Request) context.Context {
			var handler string
			if exist, ok := ctx.Value("handlerNotFound").(bool); ok && exist {
				handler = "non_existent_handler"
			} else {
				handler = metrics.MakePathTag(req.URL.Path)
			}

			ctx = context.WithValue(ctx, handlerNameCtxKey, handler)

			var span opentracing.Span
			span, ctx = gotrace.StartSpanFromRequest(ctx, req.URL.Path, opentracing.HTTPHeadersCarrier(req.Header))

			appName := appNameUnknown
			if sc, ok := span.Context().(gotrace.Span); ok {
				appName = sc.ParentAppInfo.AppName
			} else {
				logger.Error(ctx, opentracing.ErrInvalidSpanContext.Error())
			}

			httpmon.RequestCount.WithLabelValues(handler, appName).Inc()
			if service.apiHandlerCallbacks.OnInitCtx != nil {
				ctx = service.apiHandlerCallbacks.OnInitCtx(ctx, req)
			}
			ctx = response.NewContext(ctx, &response.Response{http.StatusOK})
			return ctx
		},
		On404: func(ctx context.Context, req *http.Request) {
			if r, ok := response.FromContext(ctx); ok {
				r.StatusCode = http.StatusNotFound
			}
			ctx = context.WithValue(ctx, "handlerNotFound", true)
			path := utils.GetPath(req)
			logger.Info(ctx, "path %q not found", path)
			if service.apiHandlerCallbacks.On404 != nil {
				service.apiHandlerCallbacks.On404(ctx, req)
			}
		},
		OnError: func(ctx context.Context, w http.ResponseWriter, req *http.Request, resp interface{}, err *gorpc.CallHandlerError) {
			var statusCode = http.StatusInternalServerError
			var path = utils.GetPath(req)

			logger.Error(ctx, "Error on processing request '%q': %v", path, err)

			switch err.Type {
			case gorpc.ErrorReturnedFromCall:
				// business errors must be served as successful results
				statusCode = http.StatusOK
			case gorpc.ErrorInParameters:
			case gorpc.ErrorInvalidMethod:
				logger.Error(ctx, "Invalid method has used")
			case gorpc.ErrorWriteResponse:
				if isBrokenPipe(err.Err) {
					logger.Error(ctx, "could not write response of %q: %v; client disconnected", path, err)
				} else {
					logger.Critical(ctx, "could not write response of %q: %v; response: %s", path, err, spew.Sprintf("%#v", resp))
				}
			}

			if ctx.Err() == context.DeadlineExceeded {
				statusCode = http.StatusServiceUnavailable
			}

			if r, ok := response.FromContext(ctx); ok {
				r.StatusCode = statusCode
			}

			if service.apiHandlerCallbacks.OnError != nil {
				service.apiHandlerCallbacks.OnError(ctx, w, req, resp, err)
			}
		},
		OnPanic: func(ctx context.Context, w http.ResponseWriter, r interface{}, trace []byte, req *http.Request) {
			if r, ok := response.FromContext(ctx); ok {
				r.StatusCode = http.StatusInternalServerError
			}
			logger.Critical(ctx, "%v\n%s", r, trace)
			if service.apiHandlerCallbacks.OnPanic != nil {
				service.apiHandlerCallbacks.OnPanic(ctx, w, r, trace, req)
			}
		},
		OnSuccess: func(ctx context.Context, req *http.Request, handlerResponse interface{}, startTime time.Time) {
			if service.apiHandlerCallbacks.OnSuccess != nil {
				service.apiHandlerCallbacks.OnSuccess(ctx, req, handlerResponse, startTime)
			}
		},
		OnBeforeWriteResponse: func(ctx context.Context, w http.ResponseWriter) {
			if err := gotrace.InjectSpanToResponseFromContext(ctx, opentracing.HTTPHeadersCarrier(w.Header())); err != nil {
				logger.Error(ctx, "Could not inject headers to response: %s", err)
			}
			for header, value := range service.headers {
				w.Header().Set(header, value)
			}
			if service.apiHandlerCallbacks.OnBeforeWriteResponse != nil {
				service.apiHandlerCallbacks.OnBeforeWriteResponse(ctx, w)
			}
		},
		GetCacheKey: func(ctx context.Context, req *http.Request, params interface{}) []byte {
			if service.apiHandlerCallbacks.GetCacheKey != nil {
				return service.apiHandlerCallbacks.GetCacheKey(ctx, req, params)
			}
			return nil
		},
		OnCacheHit: func(ctx context.Context, entry *cache.CacheEntry) {
			if service.apiHandlerCallbacks.OnCacheHit != nil {
				service.apiHandlerCallbacks.OnCacheHit(ctx, entry)
			}
		},
		OnCacheMiss: func(ctx context.Context) {
			if service.apiHandlerCallbacks.OnCacheMiss != nil {
				service.apiHandlerCallbacks.OnCacheMiss(ctx)
			}
		},
	}
}

// APIHandler is exposed for tests
func (service *Service) APIHandler() http.Handler {
	return service.httpJSONHandler
}

func (service *Service) AddExternalService(serviceName string) balancer.ILoadBalancer {
	venture, _ := config.GetString("venture")
	env, _ := config.GetString("env")
	log := utils.DiscoveryLogger{}

	opts := balancer.FallbackBalancerEtcd2Options{
		ServiceName:  serviceName,
		Venture:      venture,
		Environment:  env,
		BalancerType: balancer.TypeRoundRobin,
	}
	wrr := balancer.NewFallbackBalancerEtcd2(service.etcdProvider, service.etcdClientV2, log, opts)
	service.RegisterResource(BalancerResource{serviceName, wrr})

	return wrr
}

// Healthcheck returns healthcheck instance. See more info in repo https://bitbucket.lzd.co/projects/GOLIBS/repos/go-healthcheck/browse
func (service *Service) Healthcheck() *healthcheck.HealthCheck {
	service.hcMtx.Lock()
	defer service.hcMtx.Unlock()

	return service.hc
}

type HealthChecker struct {
	ResourceType string
	IsCritical   bool
	Checker      interface{}
}

func (service *Service) AddHealthcheckFunc(resourceType string, isCritical bool, checker interface{}) {
	service.hcCheckers = append(service.hcCheckers, HealthChecker{
		ResourceType: resourceType,
		IsCritical:   isCritical,
		Checker:      checker,
	})
}

type BalancerResource struct {
	name string
	b    balancer.ILoadBalancer
}

func (r BalancerResource) Caption() string {
	return r.name
}

func (r BalancerResource) Status() interfaces.Status {
	stats := interfaces.Status{
		Header: []string{"Key", "Probability", "Hit count", "RTT last", "RTT average"},
	}
	balancerStats := r.b.Stats()
	sort.Sort(balancer.StatsByProbability(balancerStats))
	for _, bs := range r.b.Stats() {
		var l string
		if bs.Healthy {
			l = interfaces.ResourceStatusOK
		} else {
			l = interfaces.ResourceStatusFail
		}

		stats.Rows = append(stats.Rows, interfaces.StatusRow{
			Level: l,
			Data: []string{
				bs.Value,
				strconv.FormatFloat(bs.HitProbability*100., 'f', 2, 64) + "%",
				strconv.FormatUint(bs.HitCount, 10),
				bs.RTT.String(),
				bs.RTTAverage.String(),
			},
		})
	}
	return stats
}

func (r BalancerResource) GetAdminURL() (string, error) {
	resourceURL, err := r.b.Next()
	if err != nil {
		return resourceURL, err
	}

	parts := strings.Split(resourceURL, ":")
	if len(parts) < 3 {
		return "", fmt.Errorf("Could not convert URL '%s' to admin URL, because port is not specified", resourceURL)
	}

	var port int
	port, err = strconv.Atoi(parts[2])
	if err != nil {
		return "", fmt.Errorf("Could not parse port (%s): %s", parts[2], err.Error())
	}

	return parts[0] + ":" + parts[1] + ":" + strconv.Itoa(port+2), nil
}

// @deprecated
func DefaultHealthCheck() map[string]interface{} {
	return map[string]interface{}{
		"status": "OK",
	}
}

func NewProxyClientFromFlags() (*http.Client, error) {
	proxyAddr, _ := config.GetString("proxy-url")
	if proxyAddr == "" {
		return http.DefaultClient, nil
	}

	proxyURL, err := url.Parse(proxyAddr)
	if err != nil {
		return nil, err
	}
	return &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		},
	}, nil
}

func isBrokenPipe(err error) bool {
	if oe, oeok := err.(*net.OpError); oeok {
		if se, seok := oe.Err.(*os.SyscallError); seok {
			if se.Err != syscall.EPIPE {
				log.Printf("DUMP: %#v", se.Err)
			}
			return se.Err == syscall.EPIPE
		}
	}
	return false
}

func (service *Service) GenerateClientLib() ([]byte, error) {
	generator := adapter.NewHttpJsonLibGenerator(service.hm.GetGoRPCHandlersManager(), "", service.id+"GoRPC")
	code, err := generator.Generate()
	if err != nil {
		return nil, err
	}
	return format.Source(code)
}
