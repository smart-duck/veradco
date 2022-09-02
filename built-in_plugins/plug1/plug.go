package main

import (
	// meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	admission "k8s.io/api/admission/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"github.com/smart-duck/veradco"
	// log "k8s.io/klog/v2"
	"fmt"
	"strings"
	"encoding/json"
)

var (
	name string = "Plug1"
)

type Plug1 struct {
	configFile string
	summary string
}

func (plug *Plug1) Init(configFile string) {
	plug.configFile = configFile
	// log.Infof("Configuration file of plugin %s: %s", name, configFile)
	plug.summary = fmt.Sprintf("Configuration file of plugin %s: %s", name, configFile)
}


func (plug *Plug1) Execute(kobj runtime.Object, operation string, dryRun bool, r *admission.AdmissionRequest) (*admissioncontroller.Result, error) {

	plug.summary += "\n" + fmt.Sprintf("Operation: %s, dryRun: %t", operation, dryRun)

	plug.summary += "\n" + fmt.Sprintf("Kind: %s, Version: %s, Group: %s", kobj.GetObjectKind().GroupVersionKind().Kind, kobj.GetObjectKind().GroupVersionKind().Version, kobj.GetObjectKind().GroupVersionKind().Group)

	pod, ok := kobj.(*v1.Pod)
	if !ok {
		plug.summary += "\n" + fmt.Sprintf("Kubernetes resource is not a pod as expected (%s)", kobj.GetObjectKind().GroupVersionKind().Kind)
		return nil, fmt.Errorf("Kubernetes resource is not a pod as expected (%s)", kobj.GetObjectKind().GroupVersionKind().Kind)
	} else {
		// https://pkg.go.dev/k8s.io/kubernetes/pkg/apis/admission#AdmissionRequest
		jsonData, err := json.Marshal(r)
		if err != nil {
			fmt.Printf("could not marshal json: %s\n", err)
		}

		// fmt.Printf("json data: %s\n", jsonData)
		plug.summary += "\n" + fmt.Sprintf("json data: %s\n", jsonData)

		for _, c := range pod.Spec.Containers {
			if strings.HasSuffix(c.Image, ":latest") {
				plug.summary += "\n" + fmt.Sprintf("Container %s is rejected", c.Name)
				if ! dryRun {
					return &admissioncontroller.Result{Msg: "You cannot use the tag 'latest' in a container."}, nil
				}
				
			}
		}
	}

	plug.summary += "\n" + fmt.Sprintf("Pod %s, accepted", pod.ObjectMeta.Name)

	return &admissioncontroller.Result{Allowed: true}, nil
}

func (plug *Plug1) Summary() string {
	return plug.summary
}

// exported as symbol named "VeradcoPlugin"
var VeradcoPlugin Plug1