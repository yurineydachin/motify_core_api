package watcher_test

import (
	"motify_core_api/godep_libs/service/watcher"
	"testing"
	"time"
)

func TestWatchForever(t *testing.T) {
	channel := make(chan bool)
	watcher.WatchForever(func() { channel <- true }, time.Nanosecond)

	<-channel
	<-channel
}

func TestWatch(t *testing.T) {
	stop := make(chan struct{})
	channel := make(chan bool)
	watcher.Watch(func() { channel <- true }, stop, time.Millisecond)

	<-channel
	<-channel

	close(stop)

	byTimer := false
	select {
	case <-channel:
	case <-time.After(5 * time.Nanosecond):
		byTimer = true
	}

	if !byTimer {
		t.Error("Watcher hasn't stopped")
	}
}
