package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	configFile := flag.String("config", "", "Path to config file")
	flag.Parse()

	config, err := LoadConfig(*configFile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	onErrorCallbacks := make([]AlertFunc, 0)
	onRecoverCallbacks := make([]AlertFunc, 0)

	onErrorCallbacks = append(onErrorCallbacks, NewLogAlert())
	onRecoverCallbacks = append(onRecoverCallbacks, NewLogResolve())

	if config.SMTP != nil {
		onErrorCallbacks = append(onErrorCallbacks, NewEmailAlert(*config.SMTP))
		onRecoverCallbacks = append(onRecoverCallbacks, NewEmailAlert(*config.SMTP))
	}

	checker := NewHealthChecker(time.Duration(config.CheckTimeoutInSec) * time.Second)
	for _, target := range config.Targets {
		url, err := url.Parse(target.Url)
		if err != nil {
			log.Fatalf("Failed to parse URL: %v", err)
		}
		checker.AddTarget(HealthTarget{ID: target.Id, URL: url, URLString: target.Url})
	}
	monitor := NewHealthMonitor(checker, time.Duration(config.CheckIntervalInSec)*time.Second, onErrorCallbacks, onRecoverCallbacks)
	go monitor.Start()

	// Create and setup server
	router := http.NewServeMux()
	server := &Server{checker: checker}
	HandlerFromMux(server, router)
	router.Handle("/metrics", promhttp.Handler())

	// Start server
	log.Printf("Starting server on :%d", config.Port)
	log.Printf("Prometheus metrics available at: /metrics")
	if err := http.ListenAndServe(fmt.Sprintf(":%d", config.Port), router); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
