package main

import (
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	gologger "motify_core_api/godep_libs/go-logger"
	"motify_core_api/godep_libs/loggo"
)

var logger = initLogger()
var app *Application

func initLogger() gologger.ILogger {
	logLevel := gologger.LevelDebug
	loggerFormatter := loggo.NewTextFormatter(":time: | :level: | :_package: | :_file: | :message: ")
	loggerHandler := loggo.NewStreamHandler(logLevel, loggerFormatter, os.Stdout)

	l := loggo.New("example", loggerHandler)
	l.AddProcessor(loggo.NewCalleeProcessor(0))
	return l
}

func handleStopSignals() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	s := <-ch
	signal.Stop(ch)
	logger.Infof("%s recieved", s)

	switch s {
	case syscall.SIGINT, syscall.SIGTERM:
		logger.Info("stopping app")
		app.Shutdown()
	}
	os.Exit(0)
}

func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	config := NewConfig()
	var err error
	app, err = NewApplication(config, logger)
	if err != nil {
		logger.Critical(err)
		os.Exit(1)
	}

	go handleStopSignals()
	if err := app.Start(); err != nil {
		logger.Criticalf("Failed to start application: %s", err)
		os.Exit(1)
	}
	if err := app.Serve(); err != nil {
		logger.Critical(err)
		os.Exit(1)
	}
}
