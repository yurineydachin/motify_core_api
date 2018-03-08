package etcd

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"os"
	"time"

	"github.com/coreos/etcd/client"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/pkg/transport"
)

const (
	dialTimeout = 5 * time.Second

	// dialKeepAliveTime is the time in seconds after which client pings the server to see if
	// transport is alive.
	dialKeepAliveTime = 3 * time.Second
	// dialKeepAliveTimeout is the time in seconds that the client waits for a response for the
	// keep-alive probe.  If the response is not received in this time, the connection is closed.
	dialKeepAliveTimeout = 2 * time.Second
)

var (
	certFile = os.Getenv("SERVICE_CERT")
	keyFile  = os.Getenv("SERVICE_KEY")
	caFile   = os.Getenv("CA_CERT")
)

// NewClient returns etcd2 client configured with given endpoints.
func NewClient(endpoints []string) (client.Client, error) {
	return NewClientWithTimeout(endpoints, dialTimeout)
}

// NewClientWithTimeout returns etcd2 client configured with given endpoints and configurable DialTimeout.
func NewClientWithTimeout(endpoints []string, timeout time.Duration) (client.Client, error) {
	var (
		t   client.CancelableTransport
		err error
	)

	if certFile != "" || keyFile != "" || caFile != "" {
		t, err = transport.NewTransport(transport.TLSInfo{
			CertFile: certFile,
			KeyFile:  keyFile,
			CAFile:   caFile,
		}, timeout)
		if err != nil {
			return nil, err
		}
	}

	return client.New(client.Config{
		Endpoints: endpoints,
		Transport: t,
	})
}

// NewClientV3 returns etcd3 client configured with given endpoints.
// The call blocks until establishing connection with dialTimeout const timeout.
func NewClientV3(endpoints []string) (*clientv3.Client, error) {
	return NewClientV3WithTimeout(endpoints, dialTimeout)
}

// NewClientV3WithTimeout returns etcd3 client configured with given endpoints and configurable DialTImeout.
// If timeout is 0 the call is non-blocking.
func NewClientV3WithTimeout(endpoints []string, timeout time.Duration) (*clientv3.Client, error) {
	cfg := clientv3.Config{
		Endpoints:            endpoints,
		DialTimeout:          timeout,
		DialKeepAliveTime:    dialKeepAliveTime,
		DialKeepAliveTimeout: dialKeepAliveTimeout,
	}
	if certFile == "" && keyFile == "" && caFile == "" {
		return clientv3.New(cfg)
	}

	// Load client cert
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}

	// Load CA cert
	caCert, err := ioutil.ReadFile(caFile)
	if err != nil {
		return nil, err
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
	}
	tlsConfig.BuildNameToCertificate()
	cfg.TLS = tlsConfig

	return clientv3.New(cfg)
}
