package main

import (
	admission "k8s.io/api/admission/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"github.com/smart-duck/veradco/veradco/admissioncontroller"
	"github.com/smart-duck/veradco/veradco/kres"
	v1 "k8s.io/api/core/v1"
	// "gopkg.in/yaml.v3"
	// "regexp"
)

var (
	name string = "AddSidecar"
)

type AddSidecar struct {
	summary string `yaml:"-"`
}

func (plug *AddSidecar) Init(configFile string) error {
	// Load configuration
	// err := yaml.Unmarshal([]byte(configFile), plug)
	// if err != nil {
	// 	// plug.summary = fmt.Sprintf("Cannot unmarshal configuration: %v", err)
	// 	return err
	// }
	return nil
}


func (plug *AddSidecar) Execute(kobj runtime.Object, operation string, dryRun bool, r *admission.AdmissionRequest) (*admissioncontroller.Result, error) {

	plug.summary = "Inject simple side car"

	var operations []admissioncontroller.PatchOperation
	pod, err := kres.ParsePod(r)
	if err != nil {
		return &admissioncontroller.Result{Msg: err.Error()}, nil
	}

	// Very simple logic to inject a new "sidecar" container.
	if pod.Namespace == "special" {
		var containers []v1.Container
		containers = append(containers, pod.Spec.Containers...)
		sideC := v1.Container{
			Name:    "test-sidecar",
			Image:   "busybox:stable",
			Command: []string{"sh", "-c", "while true; do echo 'I am a container injected by mutating webhook'; sleep 2; done"},
		}
		containers = append(containers, sideC)
		operations = append(operations, admissioncontroller.ReplacePatchOperation("/spec/containers", containers))
	}

	// Add a simple annotation using `AddPatchOperation`
	if len(pod.ObjectMeta.Annotations) > 0 {
		operations = append(operations, admissioncontroller.AddPatchOperation("/metadata/annotations/origin", "fromMutation"))
	} else {
		metadata := map[string]string{"origin": "fromMutation"}
		operations = append(operations, admissioncontroller.AddPatchOperation("/metadata/annotations", metadata))
	}

	return &admissioncontroller.Result{
		Allowed:  true,
		PatchOps: operations,
	}, nil
}

func (plug *AddSidecar) Summary() string {
	return plug.summary
}

// exported as symbol named "VeradcoPlugin"
var VeradcoPlugin AddSidecar