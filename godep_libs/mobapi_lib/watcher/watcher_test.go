package watcher_test

import (
	"testing"
	"time"

	"motify_core_api/godep_libs/mobapi_lib/watcher"
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
	watcher.Watch(func() { channel <- true }, stop, time.Nanosecond)

	<-channel
	<-channel

	close(stop)

	select {
	case <-channel:
	default:
	}

	byTimer := false
	select {
	case <-channel:
	case <-time.After(3 * time.Nanosecond):
		byTimer = true
	}

	if !byTimer {
		t.Error("Watcher hasn't stopped")
	}
}
