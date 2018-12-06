package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/kelseyhightower/confd/backends"
	"github.com/kelseyhightower/confd/log"
	"github.com/kelseyhightower/confd/resource/template"

	"github.com/kelseyhightower/confd/result"
)

func main() {
	flag.Parse()
	if config.PrintVersion {
		fmt.Printf("confd %s (Git SHA: %s, Go Version: %s)\n", Version, GitSHA, runtime.Version())
		os.Exit(0)
	}

	if err := initConfig(); err != nil {
		log.Fatal(err.Error())
	}

	log.Info("Starting confd")

	storeClient, err := backends.New(config.BackendsConfig)
	if err != nil {
		log.Fatal(err.Error())
	}

	config.TemplateConfig.StoreClient = storeClient
	if config.OneTime {
		if err := template.Process(config.TemplateConfig); err != nil {
			log.Fatal(err.Error())
		}
		os.Exit(0)
	}

	stopChan := make(chan bool)
	doneChan := make(chan bool)
	errChan := make(chan error, 10)

	// add a success msg channel
	succChan := make(chan bool)
	result.InitEtcdClient(config.BackendNodes)

	var processor template.Processor
	switch {
	case config.Watch:
		processor = template.WatchProcessor(config.TemplateConfig, stopChan, doneChan, errChan, succChan)
	default:
		processor = template.IntervalProcessor(config.TemplateConfig, stopChan, doneChan, errChan, config.Interval)
	}


	go processor.Process()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)


	for {
		select {
		case err := <-errChan:
			log.Error(err.Error())
			result.WriteEtcd(config.ResultKey, err.Error())
		case <-succChan:
			result.WriteEtcd(config.ResultKey, "ok")
		case s := <-signalChan:
			log.Info(fmt.Sprintf("Captured %v. Exiting...", s))
			close(doneChan)
		case <-doneChan:
			os.Exit(0)
		}
	}
}
