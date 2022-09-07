package http

import (
	"fmt"
	"net/http"
	"os"

	log "k8s.io/klog/v2"

	// "github.com/smart-duck/veradco/deployments"
	// "github.com/smart-duck/veradco/pods"
	
	"github.com/smart-duck/veradco/others"
	"github.com/smart-duck/veradco/cfg"
)

// NewServer creates and return a http.Server
func NewServer(port string, config string) *http.Server {

	// Load conf
	// Load conf from yaml

	log.Infof(">>>> NewServer")

	veradcoCfg := conf.VeradcoCfg{FailOnPluginLoadingFails: true}

	err := veradcoCfg.ReadConf(config)
	if err != nil {
		log.Errorf("Error loading configuration %s: %v", config, err)
		os.Exit(3)
	} else {
		log.Infof(">> Configuration %s successfully loaded\n", config)
	}

	var numberOfPluginsLoaded int
	numberOfPluginsLoaded, err = veradcoCfg.LoadPlugins()
	if err != nil {
		log.Errorf("Error loading plugins: %v", err)
		if veradcoCfg.FailOnPluginLoadingFails {
			log.Errorf("According to the configuration, exit on fail")
			os.Exit(12)
		}
	} else {
		log.Infof(">> %d plugins successfully loaded", numberOfPluginsLoaded)
	}

	// Instances hooks
	// podsValidation := pods.NewValidationHook(&veradcoCfg)
	// podsMutation := pods.NewMutationHook(&veradcoCfg)
	// deploymentValidation := deployments.NewValidationHook(&veradcoCfg)

	othersValidation := others.NewValidationHook(&veradcoCfg)
	othersMutation := others.NewMutationHook(&veradcoCfg)

	// Routers
	ah := newAdmissionHandler()
	mux := http.NewServeMux()
	mux.Handle("/healthz", healthz())
	// mux.Handle("/validate/pods", ah.Serve(podsValidation))
	// mux.Handle("/mutate/pods", ah.Serve(podsMutation))
	// mux.Handle("/validate/deployments", ah.Serve(deploymentValidation))

	mux.Handle("/validate/others", ah.Serve(othersValidation))
	mux.Handle("/mutate/others", ah.Serve(othersMutation))

	return &http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: mux,
	}
}
