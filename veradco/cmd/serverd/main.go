package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/smart-duck/veradco/veradco/http"
	"github.com/smart-duck/veradco/veradco/monitoring"

	log "k8s.io/klog/v2"
)

var (
	tlscert, tlskey, port, conf string
)

func main() {
	flag.StringVar(&tlscert, "tlscert", "/etc/certs/tls.crt", "Path to the TLS certificate")
	flag.StringVar(&tlskey, "tlskey", "/etc/certs/tls.key", "Path to the TLS key")
	flag.StringVar(&port, "port", "8443", "The port to listen")
	flag.StringVar(&conf, "conf", "/conf/veradco.yaml", "Configuration of veradco")
	log.InitFlags(nil)
	flag.Parse()

	log.Infof(">>>>>> Starting veradco")

	log.V(1).Info("log verbosity 1 enabled")
	log.V(2).Info("log verbosity 2 enabled")
	log.V(3).Info("log verbosity 3 enabled")
	log.V(4).Info("log verbosity 4 enabled")

	go func() {
	monitoring.Init()
	monitoring.StartMonitoringSvr()
	}()

	server := http.NewServer(port, conf)
	go func() {
		if err := server.ListenAndServeTLS(tlscert, tlskey); err != nil {
			log.Errorf("Failed to listen and serve: %v", err)
		}
	}()

	log.Infof(">> Server running on port: %s", port)

	// listen shutdown signal
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

	log.Infof(">> Shutdown gracefully...")
	if err := server.Shutdown(context.Background()); err != nil {
		log.Error(err)
	}
}
