package main

import (
	"context"

	"github.com/coreos/etcd/client"

	"godep.lzd.co/etcd"
)

func main() {
	// Etcd V2
	c2, err := etcd.NewClient([]string{"https://localhost:2379"})
	if err != nil {
		panic(err)
	}

	kApi := client.NewKeysAPI(c2)
	_, err = kApi.Set(context.Background(), "/test", "123", &client.SetOptions{})
	if err != nil {
		panic(err)
	}

	// Etcd V3
	c3, err := etcd.NewClientV3([]string{"https://localhost:2379"})
	if err != nil {
		panic(err)
	}

	_, err = c3.Put(context.Background(), "/test", "123")
	if err != nil {
		panic(err)
	}
}
