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
	conf, err := conf.ReadConf(config)
	if err != nil {
		log.Errorf("Error loading configuration %s: %v", config, err)
		os.Exit(3)
	} else {
		fmt.Printf("Conf: %v\n", conf)
	}

	// Instances hooks
	podsValidation := pods.NewValidationHook()
	podsMutation := pods.NewMutationHook()
	deploymentValidation := deployments.NewValidationHook()

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
