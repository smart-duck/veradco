package http

import (
	"fmt"
	"net/http"
	"os"

	log "k8s.io/klog/v2"

	// "github.com/smart-duck/veradco/deployments"
	"github.com/smart-duck/veradco/pods"

	"github.com/smart-duck/veradco/deployments"

	"github.com/smart-duck/veradco/daemonsets"

	"github.com/smart-duck/veradco/statefulsets"
	
	"github.com/smart-duck/veradco/others"
	"github.com/smart-duck/veradco/cfg"
)

const (
	ENDPOINT_VALIDATE_PODS = "/validate/pods"
	ENDPOINT_MUTATE_PODS = "/mutate/pods"
	ENDPOINT_VALIDATE_DEPLOYMENTS = "/validate/deployments"
	ENDPOINT_MUTATE_DEPLOYMENTS = "/mutate/deployments"
	ENDPOINT_VALIDATE_DAEMONSETS = "/validate/daemonsets"
	ENDPOINT_MUTATE_DAEMONSETS = "/mutate/daemonsets"
	ENDPOINT_VALIDATE_STATEFULSETS = "/validate/statefulsets"
	ENDPOINT_MUTATE_STATEFULSETS = "/mutate/statefulsets"
	ENDPOINT_VALIDATE_OTHERS = "/validate/others"
	ENDPOINT_MUTATE_OTHERS = "/mutate/others"
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
		log.Infof(">> Configuration %s successfully loaded", config)
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
	podsValidation := pods.NewValidationHook(&veradcoCfg, ENDPOINT_VALIDATE_PODS)
	podsMutation := pods.NewMutationHook(&veradcoCfg, ENDPOINT_MUTATE_PODS)
	
	deploymentValidation := deployments.NewValidationHook(&veradcoCfg, ENDPOINT_VALIDATE_DEPLOYMENTS)
	deploymentMutation := deployments.NewMutationHook(&veradcoCfg, ENDPOINT_MUTATE_DEPLOYMENTS)

	daemonsetValidation := daemonsets.NewValidationHook(&veradcoCfg, ENDPOINT_VALIDATE_DAEMONSETS)
	daemonsetMutation := daemonsets.NewMutationHook(&veradcoCfg, ENDPOINT_MUTATE_DAEMONSETS)

	statefulsetValidation := statefulsets.NewValidationHook(&veradcoCfg, ENDPOINT_VALIDATE_STATEFULSETS)
	statefulsetMutation := statefulsets.NewMutationHook(&veradcoCfg, ENDPOINT_MUTATE_STATEFULSETS)


	othersValidation := others.NewValidationHook(&veradcoCfg, ENDPOINT_VALIDATE_OTHERS)
	othersMutation := others.NewMutationHook(&veradcoCfg, ENDPOINT_MUTATE_OTHERS)

	// Routers
	ah := newAdmissionHandler()
	mux := http.NewServeMux()
	mux.Handle("/healthz", healthz())

	mux.Handle(ENDPOINT_VALIDATE_PODS, ah.Serve(podsValidation))
	mux.Handle(ENDPOINT_MUTATE_PODS, ah.Serve(podsMutation))

	mux.Handle(ENDPOINT_VALIDATE_DEPLOYMENTS, ah.Serve(deploymentValidation))
	mux.Handle(ENDPOINT_MUTATE_DEPLOYMENTS, ah.Serve(deploymentMutation))

	mux.Handle(ENDPOINT_VALIDATE_DAEMONSETS, ah.Serve(daemonsetValidation))
	mux.Handle(ENDPOINT_MUTATE_DAEMONSETS, ah.Serve(daemonsetMutation))

	mux.Handle(ENDPOINT_VALIDATE_STATEFULSETS, ah.Serve(statefulsetValidation))
	mux.Handle(ENDPOINT_MUTATE_STATEFULSETS, ah.Serve(statefulsetMutation))

	mux.Handle(ENDPOINT_VALIDATE_OTHERS, ah.Serve(othersValidation))
	mux.Handle(ENDPOINT_MUTATE_OTHERS, ah.Serve(othersMutation))

	return &http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: mux,
	}
}
