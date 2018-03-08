package main

import (
	"context"
	"flag"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/coreos/etcd/clientv3"
)

const (
	maxTxnLen = 128
)

func newClient(endpoints string) *clientv3.Client {
	parsedEndpoints := strings.Split(endpoints, ",")

	c, err := clientv3.New(clientv3.Config{
		Endpoints:   parsedEndpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		log.Fatalf("etcd connection error: %s", err)
	}
	return c
}

// updateReleases updates release version for defined segregationID interval
func updateReleases(ctx context.Context, kv clientv3.KV, releaseName string, from, to int) {
	defer deferedTiming("updateReleases", time.Now())

	wg := &sync.WaitGroup{}
	for i := from; i < to; i += maxTxnLen {
		wg.Add(1)
		go func(f int) {
			defer wg.Done()
			t := f + maxTxnLen - 1
			if t > to {
				t = to
			}

			err := setRelease(ctx, kv, f, t, releaseName)
			if err != nil {
				log.Printf("error setting %s version, %d - %d: %s", releaseName, f, t, err)
			}
		}(i)
	}
	wg.Wait()
}

func main() {
	var verbose bool
	var endpoints string
	var from, to int
	var rolloutType string

	flag.StringVar(&endpoints, "endpoints", "localhost:2379,localhost:4001", "Etcd gRPC endpoints")
	flag.BoolVar(&verbose, "verbose", false, "If true - increases logging verbosity")
	flag.IntVar(&from, "from", 1, "segregation ID interval start")
	flag.IntVar(&to, "to", 1000, "segregation ID interval end")
	flag.StringVar(&rolloutType, "rollout", "stable", "rollout type: (stable | unstable1 | ... | unstable20)")
	flag.Parse()

	client := newClient(endpoints)
	defer client.Close()

	log.Printf("Setting rollout %q from for %d-%d segregation IDs", rolloutType, from, to)
	ctx, cancel := context.WithTimeout(newContext(verbose), 2*time.Second)
	defer cancel()

	updateReleases(ctx, client, rolloutType, from, to)
	log.Println("Profit!")
}
