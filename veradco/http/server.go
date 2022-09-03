package http

import (
	"fmt"
	"net/http"
	"os"

	log "k8s.io/klog/v2"

	"github.com/smart-duck/veradco/deployments"
	"github.com/smart-duck/veradco/pods"
	"github.com/smart-duck/veradco/cfg"
)

// NewServer creates and return a http.Server
func NewServer(port string, config string) *http.Server {

	// Load conf
	// Load conf from yaml

	veradcoCfg := conf.VeradcoCfg{}

	err := veradcoCfg.ReadConf(config)
	if err != nil {
		log.Errorf("Error loading configuration %s: %v", config, err)
		os.Exit(3)
	} else {
		log.Infof("Configuration %s succesfully loaded\n", config)
	}

	err = veradcoCfg.LoadPlugins()
	if err != nil {
		log.Errorf("Error loading plugins: %v", err)
		os.Exit(12)
	} else {
		log.Infof("Plugins succesfully loaded")
	}

	// Instances hooks
	podsValidation := pods.NewValidationHook(&veradcoCfg)
	podsMutation := pods.NewMutationHook(&veradcoCfg)
	deploymentValidation := deployments.NewValidationHook(&veradcoCfg)

	// Routers
	ah := newAdmissionHandler()
	mux := http.NewServeMux()
	mux.Handle("/healthz", healthz())
	mux.Handle("/validate/pods", ah.Serve(podsValidation))
	mux.Handle("/mutate/pods", ah.Serve(podsMutation))
	mux.Handle("/validate/deployments", ah.Serve(deploymentValidation))

	return &http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: mux,
	}
}
