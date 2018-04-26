package service

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"go/format"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	etcd_client "github.com/coreos/etcd/client"
	etcdcl3 "github.com/coreos/etcd/clientv3"
	"github.com/facebookgo/grace/gracehttp"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sergei-svistunov/gorpc"
	"github.com/sergei-svistunov/gorpc/transport/cache"
	"github.com/sergei-svistunov/gorpc/transport/http_json"
	"github.com/sergei-svistunov/gorpc/transport/http_json/adapter"
	"motify_core_api/godep_libs/discovery/balancer"
	lzd_grpc "motify_core_api/godep_libs/discovery/balancer/grpc"
	"motify_core_api/godep_libs/discovery/locator"
	"motify_core_api/godep_libs/discovery/provider"
	"motify_core_api/godep_libs/discovery/provider/etcdV3"
	"motify_core_api/godep_libs/discovery/registrator"
	"motify_core_api/godep_libs/etcd"
	"motify_core_api/godep_libs/go-config"
	"motify_core_api/godep_libs/go-dconfig"
	healthcheck "motify_core_api/godep_libs/go-healthcheck"
	golog "motify_core_api/godep_libs/go-log"
	gotrace "motify_core_api/godep_libs/go-trace"
	"motify_core_api/godep_libs/mobapi_lib/admin"
	"motify_core_api/godep_libs/mobapi_lib/cache/inmem"
	"motify_core_api/godep_libs/mobapi_lib/closer"
	"motify_core_api/godep_libs/mobapi_lib/handler"
	"motify_core_api/godep_libs/mobapi_lib/handlersmanager"
	"motify_core_api/godep_libs/mobapi_lib/logger"
	"motify_core_api/godep_libs/mobapi_lib/resources"
	"motify_core_api/godep_libs/mobapi_lib/sessionlogger"
	"motify_core_api/godep_libs/mobapi_lib/sessionmocker"
	"motify_core_api/godep_libs/mobapi_lib/token"
	"motify_core_api/godep_libs/mobapi_lib/utils"
	"motify_core_api/godep_libs/mobapi_lib/utils/middleware"
	"google.golang.org/grpc"
)

var (
	AppVersion  string // this value set by compiler
	GoVersion   string // this value set by compiler
	BuildDate   string // this value set by compiler
	GitDescribe string // this value set by compiler
)

// Service implements all service methods
type Service struct {
	id                   string
	handlersPath         string
	httpMux              *http.ServeMux
	dconfManager         *dconfig.Manager
	transportCache       cache.ICache
	hm                   handlersmanager.IHandlersManager
	resources            []resources.IResource
	hc                   *healthcheck.HealthCheck
	hcMtx                sync.Mutex
	etcdClient           etcd_client.Client
	httpJSONHandler      http.Handler
	clientGenHandler     http.HandlerFunc
	apiHandlerCallbacks  http_json.APIHandlerCallbacks
	swaggerJSONHandler   *http_json.SwaggerJSONHandler
	swaggerJSONCallbacks http_json.SwaggerJSONCallbacks
	etcdClientV3         *etcdcl3.Client
	etcdProvider         provider.IProvider
	etcdRegistrator      registrator.IRegistrator
	isInitialized        bool
}

func init() {
	config.RegisterBool("version", "display binary version and info and exit", false)
	config.RegisterBool("print-interface", "Print interface and exit", false)
	config.RegisterBool("gen-client-lib", "Print client sdk to stdout", false)
	config.RegisterBool("print-debug", "Print debug output on internal server error", false)

	config.RegisterString("venture", "Current venture", "UNKNOWN")
	config.RegisterString("env", "Current envirement", "dev")

	config.RegisterString("etcd-namespace", "Current namespace", "lazada_api")
	config.RegisterString("etcd-endpoints", "List of etcd hosts joined by ,", "http://localhost:4001")
	config.RegisterBool("etcd-registration-enabled", "Enable registration service in etcd", true)

	config.RegisterString("addr", "<hostname/ip>:<port> for handlers. This param is overridden by the 'port' param", ":8080")
	config.RegisterString("admin-addr", "<hostname/ip>:<port> for administrative page. This param is overridden by the 'admin-port'", ":8081")
	config.RegisterString("profile-addr", "<hostname/ip>:<port> for metrics. This param is overridden by the 'profile-port' param", ":8082")

	// https://confluence.lazada.com/display/RE/Requirements+to+Docker-based+deploy+process+for+GO+components
	config.RegisterUint("port", "The parameter of port for main HTTP listener. Numeric type between the range of 1024 and 32767. Overrides the 'addr' param", 0)
	config.RegisterUint("admin-port", "The parameter of port for listener administrative tools. Numeric type between the range of 1024 and 32767. Overrides the 'admin-addr' param", 0)
	config.RegisterUint("profile-port", "The parameter of port for metrics. Numeric type between the range of 1024 and 32767. Overrides the 'profile-addr' param", 0)
	config.RegisterUint("grpc_port", "port number for gRPC services", 0)

	config.RegisterUint("ssl-port", "SSL port for handlers", 0)
	config.RegisterString("ssl-certificate", "Specifies a file with the certificate in the PEM format for the given virtual server", "")
	config.RegisterString("ssl-certificate-key", "Specifies a file with the secret key in the PEM format for the given virtual server", "")
	config.RegisterString("ssl-client-certificate", "Specifies a file with trusted CA certificates in the PEM format used to verify client certificates", "")

	config.RegisterString("token-triple-des-key", "24-bit key for token DES encryption", "")
	config.RegisterString("token-salt", "8-bit salt for token DES encryption", "")

	config.RegisterBool("inmem-cache-enabled", "Use inmem cache", true)

	config.RegisterString("syslog_addr_type", "Syslog address type (unixgram, tcd, udp). StdOut lgging by default if empty", "")
	config.RegisterString("syslog_addr", "Syslog address. StdOut logging by default if empty", "")

	config.RegisterString("advertised-hostname", "os.Hostname replacement", "")

	dconfig.RegisterBool("print-debug", "Enable output debug info in a response", false,
		func(val bool) {
			http_json.PrintDebug = val
		})
}

// New creates new Service
func New(id string, handlersPath string) *Service {
	service := &Service{
		id:           id,
		handlersPath: handlersPath,
	}

	err := service.parseFlags()
	prepareLogger(id)

	if err != nil {
		logger.Error(nil, err.Error())
	}
	logger.Notice(nil, "app configuration: %s", config.String())

	return service
}

// Init performs service initialization without running.
// It prepares etcd/grpc/discovery clients
func (service *Service) Init() error {
	if service.isInitialized {
		return nil
	}

	if err := initToken(); err != nil {
		logger.Critical(nil, "failed to init token encryption: %v", err)
		return err
	}

	sessionLogger, err := sessionlogger.NewSessionLoggerFromFlags(service.dconfManager)
	if err != nil {
		logger.Critical(nil, err.Error())
		return err
	}

	err = sessionmocker.InitSessionMockerFromFlags()
	if err != nil {
		logger.Critical(nil, "failed to init session mocker: %v", err)
		return err
	}

	if service.transportCache == nil {
		service.transportCache = inmem.NewFromFlags("")
	}

	if service.dconfManager == nil {
		service.dconfManager = dconfig.NewManager(service.id, logger.GetLoggerInstance())
	}

	venture, _ := config.GetString("venture")

	addr, port := getAddr()
	_, admPort := getAdmAddr()
	_, profilePort := getProfileAddr()
	_, grpcPort := getGrpcAddr(port)
	env, _ := config.GetString("env")
	rolloutID := "stable"
	hostname := addr[:strings.Index(addr, ":")]

	if hostname == "" || hostname == "0.0.0.0" {
		var err error
		hostname, err = middleware.GetHostname()
		if err != nil {
			logger.Critical(nil, "could not get hostname: %v", err)
			return err
		}
	}
	opentracing.SetGlobalTracer(gotrace.NewTracer(golog.NewCollector(logger.GetLoggerInstance().Logger(), hostname, AppVersion)))

	err = service.initDiscoveryRegistrator(venture, env, rolloutID, int(port), int(admPort), int(profilePort), int(grpcPort))
	if err != nil {
		logger.Critical(nil, "Failed to initialize discovery registrator: %v", err)
		return err
	}

	dconfig.RegisterDuration("handler_timeout", "Timeout for processing requests (0 - unlimited)", 0,
		func(timeout time.Duration) {
			if h, ok := service.httpJSONHandler.(*http_json.APIHandler); ok {
				h.SetTimeout(timeout)
				logger.Debug(nil, "handler_timeout changed, new value: %s", timeout)
			}
		})
	timeout, _ := dconfig.GetDuration("handler_timeout")

	service.apiHandlerCallbacks = handler.NewHTTPHandlerCallbacks(service.id, AppVersion, hostname, sessionLogger)
	service.hm = handlersmanager.New(service.handlersPath)

	service.httpJSONHandler = http_json.NewAPIHandler(
		service.hm.GetGoRPCHandlersManager(), service.transportCache, service.apiHandlerCallbacks).SetTimeout(timeout)
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
	service.httpMux = mux

	service.swaggerJSONCallbacks = admin.NewSwaggerJSONCallbacks(venture)

	mux.HandleFunc("/rev.txt", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		fmt.Fprintln(w, AppVersion)
	})

	buildDate, er := time.Parse("2006-01-02 15:04:05", BuildDate)
	if er != nil {
		logger.Critical(nil, "Build date string '%s' parsing error. Please check your Makefile: %v", BuildDate, er)
		return err
	}
	service.hc = healthcheck.NewHealthCheck(service.id, AppVersion, GitDescribe, GoVersion, venture, buildDate.Unix())
	structCacheChecker := func(ctx context.Context) error {
		body := 200
		key := []byte("health_check_test")
		value := &cache.CacheEntry{
			Content:           []byte("content"),
			CompressedContent: []byte("compressed_content"),
			Hash:              "hash",
			Body:              body,
		}
		service.transportCache.Put(key, value)
		result := service.transportCache.Get(key)

		var err error
		if result == nil {
			err = fmt.Errorf("nil result")
		}
		if len(result.Content) != len(value.Content) ||
			len(result.CompressedContent) != len(value.CompressedContent) ||
			result.Hash != value.Hash ||
			result.Body != value.Body {
			err = fmt.Errorf("Invalid data in cache")
		}

		return err
	}
	service.hc.SetResourceChecker(healthcheck.ResourceTypeStructCache, false, structCacheChecker)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		admin.RPSCounter.Inc(1)
		switch r.URL.Path {
		case "", "/":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte{})
		case "/health_check", "/health_check/":
			service.hc.HandlerFunc()(w, r)
		default:
			service.httpJSONHandler.ServeHTTP(w, r)
		}
	})

	var swaggerPort uint64
	if port != admPort {
		swaggerPort = port
	}
	service.swaggerJSONHandler = http_json.NewSwaggerJSONHandler(service.hm.GetGoRPCHandlersManager(), uint16(swaggerPort), service.swaggerJSONCallbacks)

	service.isInitialized = true

	return nil
}

func (service *Service) Run() {
	if err := service.run(); err != nil {
		logger.Critical(nil, "Service is sropped with error: %s", err)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()
		if er := logger.Flush(ctx); er != nil {
			panic(err)
		}
	}
}

// Run starts service and register it in discovery
func (service *Service) run() error {
	if !service.isInitialized {
		return fmt.Errorf("You've tried to run non-initialized service. Please run Service.Init() first")
	}

	if printInterface, _ := config.GetBool("print-interface"); printInterface {
		swagger, err := http_json.GenerateSwaggerJSON(service.hm.GetGoRPCHandlersManager(), "", service.swaggerJSONCallbacks)
		if err == nil {
			err = json.NewEncoder(os.Stdout).Encode(swagger)
		}
		if err != nil {
			logger.Critical(nil, err.Error())
			return err
		}
		os.Exit(0)
	}
	if genClientLib, _ := config.GetBool("gen-client-lib"); genClientLib {
		code, err := service.GenerateClientLib()
		if err != nil {
			return err
		}
		fmt.Println(string(code))
		os.Exit(0)
	}

	addr, _ := getAddr()
	adminAddr, _ := getAdmAddr()
	profileAddr, _ := getProfileAddr()
	venture, _ := config.GetString("venture")
	env, _ := config.GetString("env")
	ns, _ := config.GetString("etcd-namespace")

	sslPort, _ := config.GetUint("ssl-port")
	sslCertificate, _ := config.GetString("ssl-certificate")
	sslCertificateKey, _ := config.GetString("ssl-certificate-key")
	sslClientCertificate, _ := config.GetString("ssl-client-certificate")

	service.dconfManager.Run(service.etcdClient, ns, venture, env)

	profileMux := http.NewServeMux()

	if registrationEnabled, _ := config.GetBool("etcd-registration-enabled"); registrationEnabled {
		var err error
		if err = service.etcdRegistrator.Register(); err != nil {
			return err
		}

		if err = service.etcdRegistrator.EnableDiscovery(); err != nil {
			return err
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
	ah = admin.NewHTTPHandler(service.id, AppVersion, venture, env, service.resources, service.swaggerJSONHandler, service.clientGenHandler)
	ph = profileMux

	servers := []*http.Server{
		{
			Addr:    addr,
			Handler: h,
		},
		{
			Addr:    adminAddr,
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

	logger.Alert(nil, "Service %s starts serving; venture: %s, env: %s, version: %s, git-describe: %s", service.id, venture, env, AppVersion, GitDescribe)

	return gracehttp.Serve(servers...)
}

func (service *Service) initDiscoveryRegistrator(venture, env, rolloutID string, port, adminPort, profilePort, grpcPort int) (err error) {
	var host string
	host, err = middleware.GetHostname()
	if err != nil {
		return fmt.Errorf("cannot get hostname: %v", err)
	}
	log := utils.DiscoveryLogger{}
	etcdEndpoints, _ := config.GetStringSlice("etcd-endpoints")

	// init ETCD v2 client
	service.etcdClient, err = etcd.NewClient(etcdEndpoints)
	if err != nil {
		return fmt.Errorf("Failed to create etcd client: %v", err)
	}

	// init ETCD v3 client
	service.etcdClientV3, err = etcd.NewClientV3(utils.GetGrpcEndpoints(etcdEndpoints))
	if err != nil {
		return fmt.Errorf("Failed to create etcd client v3: %v", err)
	}
	service.etcdProvider = etcdV3.NewProvider(service.etcdClientV3, log)

	info := registrator.AppRegistrationParams{
		ServiceName: service.id,
		RolloutType: rolloutID,
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

	service.etcdRegistrator, err = registrator.NewV2V3(info, service.etcdProvider, service.etcdClient, log)
	if err != nil {
		return fmt.Errorf("Failed to create etcd registrator: %v", err)
	}
	return nil
}

// MustRegisterHandlers registers handlers that implement gorpc.IHandler interface and throws panic if error occured
func (service *Service) MustRegisterHandlers(handlers ...gorpc.IHandler) {
	if err := service.RegisterHandlers(handlers...); err != nil {
		panic(err)
	}
}

// RegisterHandlers registers handlers that implement gorpc.IHandler interface and return error if it occured
func (service *Service) RegisterHandlers(handlers ...gorpc.IHandler) error {
	for _, h := range handlers {
		if err := service.registerHandler(h); err != nil {
			return err
		}
	}
	return nil
}

func (service *Service) registerHandler(handler gorpc.IHandler) error {
	if !service.isInitialized {
		return fmt.Errorf("You've tried to register handler in non-initialized service. Please run Service.Init() first")
	}
	return service.hm.RegisterHandler(handler)
}

// RegisterResources registers resources that implement IResource interface
func (service *Service) RegisterResources(resources ...resources.IResource) {
	service.resources = append(service.resources, resources...)
}

// AddExternalService registers service as external resource and returns prepared weighted round robin balancer
// that implements balancer.ILoadBalancer interface
func (service *Service) AddExternalService(serviceName string) balancer.ILoadBalancer {
	venture, _ := config.GetString("venture")
	env, _ := config.GetString("env")

	opts := balancer.FallbackBalancerEtcd2Options{
		ServiceName:  serviceName,
		Venture:      venture,
		Environment:  env,
		BalancerType: balancer.TypeRoundRobin,
	}

	wrr := balancer.NewFallbackBalancerEtcd2(service.etcdProvider, service.etcdClient, utils.DiscoveryLogger{}, opts)

	service.RegisterResources(balancerResource{serviceName, wrr})

	return wrr
}

// AddExternalGRPCService registers service as external resource and returns prepared weighted round robin balancer
// that implements grpc.Balancer interface
func (service *Service) AddExternalGRPCService(serviceName string) grpc.Balancer {
	wrr := balancer.NewRoundRobin(
		locator.New(service.etcdProvider, utils.DiscoveryLogger{}),
		utils.DiscoveryLogger{},
		balancer.LoadBalancerOptions{
			ServiceName:  serviceName,
			EndpointType: locator.TypeAppAdditional,
		})
	b := lzd_grpc.NewBalancer(wrr)

	service.RegisterResources(grpcResource{serviceName, b})
	return b
}

// Healthcheck returns healthcheck instance. See more info in repo https://bitbucket.lzd.co/projects/GOLIBS/repos/go-healthcheck/browse
func (service *Service) Healthcheck() *healthcheck.HealthCheck {
	service.hcMtx.Lock()
	defer service.hcMtx.Unlock()

	return service.hc
}

// GenerateClientLib generates service client library GO code
func (service *Service) GenerateClientLib() ([]byte, error) {
	generator := adapter.NewHttpJsonLibGenerator(service.hm.GetGoRPCHandlersManager(), "", service.id+"GoRPC")
	code, err := generator.Generate()
	if err != nil {
		return nil, err
	}
	return format.Source(code)
}

func (service *Service) parseFlags() error {
	if err := config.Parse(); err != nil {
		panic(err)
	}

	if displayVersion, _ := config.GetBool("version"); displayVersion {
		fmt.Printf("%s version %s\n", service.id, AppVersion)
		fmt.Printf("built at %s with compiler %s\n", BuildDate, GoVersion)
		fmt.Printf("git describe %s\n", GitDescribe)
		os.Exit(0)
	}

	if err := config.ParseAll(); err != nil {
		return err
	}

	return nil
}

func prepareLogger(serviceName string) {
	syslogAddrType, _ := config.GetString("syslog_addr_type")
	syslogAddr, _ := config.GetString("syslog_addr")

	dconfig.RegisterString("log-level", "Logging level (DEBUG, INFO, NOTICE, WARNING, ERROR, CRITICAL, ALERT, EMERGENCY), case-insensitive", "ERROR",
		func(val string) {
			if ok := logger.ParseAndSetLevel(val); !ok {
				logger.Error(nil, "Could not parse and set log level %s", val)
			}
		})
	level, _ := dconfig.GetString("log-level")

	if err := logger.Init(serviceName, syslogAddrType, syslogAddr, level); err != nil {
		panic(err)
	}

	if genClientLib, _ := config.GetBool("gen-client-lib"); genClientLib {
		ok := logger.ParseAndSetLevel("WARNING")
		if !ok {
			logger.Error(nil, "Could not set logger level to WARNING. Parsing error")
		} else {
			logger.Info(nil, "Force set logging level to DEBUG because of gen-client-lib setting is %t", genClientLib)
		}
	}

	if printInterface, _ := config.GetBool("print-interface"); printInterface {
		ok := logger.ParseAndSetLevel("DEBUG")
		if !ok {
			logger.Error(nil, "Could not set logger level to DEBUG. Parsing error")
		} else {
			logger.Info(nil, "Force set logging level to DEBUG because of print-interface setting is %t", printInterface)
		}
	}
}

func getAddr() (string, uint64) {
	if port, _ := config.GetUint("port"); port != 0 {
		return ":" + strconv.FormatUint(uint64(port), 10), uint64(port)
	}

	addr, _ := config.GetString("addr")
	_, strPort, _ := net.SplitHostPort(addr)
	port, err := strconv.ParseUint(strPort, 10, 16)
	if err != nil {
		panic(err)
	}
	return addr, port

}

func getAdmAddr() (string, uint64) {
	if port, _ := config.GetUint("admin-port"); port != 0 {
		return ":" + strconv.FormatUint(uint64(port), 10), uint64(port)
	}

	addr, _ := config.GetString("admin-addr")
	_, strPort, _ := net.SplitHostPort(addr)
	port, err := strconv.ParseUint(strPort, 10, 16)
	if err != nil {
		panic(err)
	}
	return addr, port
}

func getProfileAddr() (string, uint64) {
	if port, _ := config.GetUint("profile-port"); port != 0 {
		return ":" + strconv.FormatUint(uint64(port), 10), uint64(port)
	}

	addr, _ := config.GetString("profile-addr")
	_, strPort, _ := net.SplitHostPort(addr)
	port, err := strconv.ParseUint(strPort, 10, 16)
	if err != nil {
		panic(err)
	}
	return addr, port
}

//getGrpcAddr return address for gRPC server to listening
//parameter is restful port - main port of service
//we follow the rule of port convention that defined by:
//https://confluence.lazada.com/display/DEV/Services+Reference+Documentation#ServicesReferenceDocumentation-Portassignmentconventionsforapplications
func getGrpcAddr(port uint64) (string, uint64) {
	grpcPort, _ := config.GetUint("grpc_port")
	if grpcPort > 0 {
		return ":" + strconv.FormatUint(uint64(grpcPort), 10), uint64(grpcPort)
	}
	return ":" + strconv.FormatUint(uint64(port+6), 10), uint64(port) + 6
}

// APIHandler is exposed for tests
func (service *Service) APIHandler() http.Handler {
	return service.httpJSONHandler
}

func initToken() error {
	key, _ := config.GetString("token-triple-des-key")
	salt, _ := config.GetString("token-salt")
	return token.InitTokenV1([]byte(key), []byte(salt))
}
