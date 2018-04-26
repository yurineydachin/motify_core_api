/*
Install test CA certificate, for Ubuntu:
	sudo mkdir /usr/share/ca-certificates/lazada
	sudo cp certs/ca.cert.pem /usr/share/ca-certificates/lzd_etcd_ca.crt
	sudo dpkg-reconfigure ca-certificates
	sudo update-ca-certificates

Run etcd before start tests:

	etcd --listen-client-urls "https://localhost:2379" --advertise-client-urls "https://localhost:2379" \
  		--trusted-ca-file=./certs/ca.cert.pem --cert-file=./certs//localhost.crt \
  		--key-file=./certs/localhost.key.pem
*/

package etcd_test

import (
	"context"
	"testing"

	"github.com/coreos/etcd/client"

	"motify_core_api/godep_libs/etcd"
)

const REAL_RUN = false

func TestNewClient(t *testing.T) {
	if !REAL_RUN {
		return
	}

	etcd.SetCerts()

	c, err := etcd.NewClient([]string{"https://localhost:2379"})
	if err != nil {
		t.Fatal(err)
	}

	kApi := client.NewKeysAPI(c)
	_, err = kApi.Set(context.Background(), "/test", "123", &client.SetOptions{})
	if err != nil {
		t.Fatal(err)
	}
}

func TestNewClientV3(t *testing.T) {
	if !REAL_RUN {
		return
	}

	etcd.SetCerts()

	c, err := etcd.NewClientV3([]string{"https://localhost:2379"})
	if err != nil {
		t.Fatal(err)
	}

	_, err = c.Put(context.Background(), "/test", "123")
	if err != nil {
		t.Fatal(err)
	}
}

func TestUnsecuredNewClient(t *testing.T) {
	if !REAL_RUN {
		return
	}

	etcd.UnsetCerts()

	c, err := etcd.NewClient([]string{"https://localhost:2379"})
	if err != nil {
		t.Fatal(err)
	}

	kApi := client.NewKeysAPI(c)
	_, err = kApi.Set(context.Background(), "/test", "123", &client.SetOptions{})
	if err != nil {
		t.Fatal(err)
	}
}

func TestUnsecuredNewClientV3(t *testing.T) {
	if !REAL_RUN {
		return
	}

	etcd.UnsetCerts()

	c, err := etcd.NewClientV3([]string{"https://localhost:2379"})
	if err != nil {
		t.Fatal(err)
	}

	_, err = c.Put(context.Background(), "/test", "123")
	if err != nil {
		t.Fatal(err)
	}
}
