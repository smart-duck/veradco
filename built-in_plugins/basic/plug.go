package main

import (
	admission "k8s.io/api/admission/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"github.com/smart-duck/veradco"
	"fmt"
)

var (
	name string = "Basic"
)

type Basic struct {
	summary string
}

func (plug *Basic) Init(configFile string) error {
	// plug.summary = fmt.Sprintf("Configuration of plugin %s is: %s", name, configFile)
	return nil
}


func (plug *Basic) Execute(kobj runtime.Object, operation string, dryRun bool, r *admission.AdmissionRequest) (*admissioncontroller.Result, error) {
	plug.summary = ""
	pod, ok := kobj.(*v1.Pod)
	if !ok {
		plug.summary += "\n" + fmt.Sprintf("Kubernetes resource is not a pod as expected (%s)", kobj.GetObjectKind().GroupVersionKind().Kind)
		return nil, fmt.Errorf("Kubernetes resource is not a pod as expected (%s)", kobj.GetObjectKind().GroupVersionKind().Kind)
	} else {
		plug.summary += "\n" + fmt.Sprintf("As expected it is a pod: %s, %s, %s, %s", pod.TypeMeta.Kind,
			pod.TypeMeta.APIVersion, pod.ObjectMeta.Name, pod.ObjectMeta.Namespace)

		for i, c := range pod.Spec.Containers {
			plug.summary += "\n" + fmt.Sprintf("Container %d: name %s, image %s", i, c.Name, c.Image)
		}
	}

	plug.summary += "\n" + fmt.Sprintf("Pod %s, accepted", pod.ObjectMeta.Name)

	return &admissioncontroller.Result{Allowed: true}, nil
}

func (plug *Basic) Summary() string {
	return plug.summary
}

// exported as symbol named "VeradcoPlugin"
var VeradcoPlugin Basic