package main

import (
	"sync"
	"time"

	"motify_core_api/godep_libs/go-dconfig"

	etcdcl "github.com/coreos/etcd/client"
	"context"
)

func main() {
	wg := sync.WaitGroup{}
	wg.Add(1)

	// Register options must be before manager.Run
	dconfig.RegisterString("string-value", "Descrition", "default value", func(val string) {
		// Callback code
		println("New value callback:", val)
		wg.Done()
	})

	// Run manager
	etcdClient, err := etcdcl.New(etcdcl.Config{
		Endpoints: []string{"http://localhost:4001"},
	})
	if err != nil {
		panic(err)
	}

	manager := dconfig.NewManager("example-service")
	manager.Run(etcdClient, "example-ns", "test-venture", "dev-env")

	time.Sleep(time.Second) // ToDo: Return from Run() afer all goroutins is initialized

	// Edit value
	manager.EditSetting(context.Background(), "name@example.com", "string-value", "newValue")
	wg.Wait()

	// Get current value
	val, _ := dconfig.GetString("string-value")
	println("Current value:", val)
}
