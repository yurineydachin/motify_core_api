# godep.lzd.co/go-dconfig
## Description
The library uses etcd as config storage. It can add options to etcd, edit their values
with saving history (last 10 records) and listen changes. If one node change value all
other nodes will receive notify.

## Example
**main.go**
```go
package main

import (
	"sync"
	"time"

	"godep.lzd.co/go-dconfig"

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
```