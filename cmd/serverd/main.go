package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/smart-duck/veradco/http"

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
	flag.Parse()

	server := http.NewServer(port, conf)
	go func() {
		if err := server.ListenAndServeTLS(tlscert, tlskey); err != nil {
			log.Errorf("Failed to listen and serve: %v", err)
		}
	}()

	log.Infof("Server running on port: %s", port)

	// listen shutdown signal
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

	log.Infof("Shutdown gracefully...")
	if err := server.Shutdown(context.Background()); err != nil {
		log.Error(err)
	}
}
