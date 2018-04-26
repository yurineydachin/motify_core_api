package watcher

import "time"

func Watch(function func(), stop chan struct{}, period time.Duration) {
	go func() {
		function()

		for {
			select {
			case <-stop:
				return
			case <-time.After(period):
				function()
			}
		}
	}()
}

func WatchForever(function func(), period time.Duration) {
	Watch(function, nil, period)
}
