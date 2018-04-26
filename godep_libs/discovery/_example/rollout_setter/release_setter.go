package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/coreos/etcd/clientv3"
)

const (
	keyTmpl = "/rollout/segregation/%03d"
)

func getKey(segregationID int) string {
	return fmt.Sprintf(keyTmpl, segregationID)
}

func deferedTiming(name string, start time.Time) {
	log.Printf("%s executed in %s", name, time.Now().Sub(start))
}

// setRelease implements releaseSetter using etcd3 transactions
func setRelease(ctx context.Context, kv clientv3.KV, from, to int, releaseVersion string) error {
	if getVerbose(ctx) {
		defer deferedTiming(fmt.Sprintf("setRelease '%s'", releaseVersion), time.Now())
	}

	count := to - from + 1
	if count <= 0 {
		return fmt.Errorf("to '%d' must be >= from '%d'", to, from)
	}
	ops := make([]clientv3.Op, 0, count)
	for i := 0; i < count; i++ {
		ops = append(ops, clientv3.OpPut(getKey(from+i), releaseVersion))
	}

	_, err := kv.Txn(ctx).Then(ops...).Commit()
	return err
}
