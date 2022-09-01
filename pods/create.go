package pods

import (
	"strings"

	"plugin"

	"github.com/smart-duck/veradco"

	veradcoplugin "github.com/smart-duck/veradco/plugin"

	log "k8s.io/klog/v2"

	admission "k8s.io/api/admission/v1"
	v1 "k8s.io/api/core/v1"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func validateCreate() admissioncontroller.AdmitFunc {
	return func(r *admission.AdmissionRequest) (*admissioncontroller.Result, error) {

		pod, err := parsePod(r.Object.Raw)
		if err != nil {
			return &admissioncontroller.Result{Msg: err.Error()}, nil
		}











		// Try to execute plugins
		plug, err := plugin.Open("/plugs/plug1/plug.so")
		if err != nil {
			log.Infof("Unable to load plugin: %v\n", err)
		}

		pluginHandler, err := plug.Lookup("VeradcoPlugin")
		if err != nil {
			log.Infof("Unable to find handler for plugin: %v\n", err)
		}

		var veradcoPlugin veradcoplugin.VeradcoPlugin

		veradcoPlugin, ok := pluginHandler.(veradcoplugin.VeradcoPlugin)
		if !ok {
			log.Infof("Plugin does not implement awaited interface\n")
		} else {
			log.Infof("Init plugin\n")
			veradcoPlugin.Init("Path to config file")
			log.Infof("Execute plugin\n")
			// Execute(meta meta.TypeMeta, kobj interface{}, r *admission.AdmissionRequest) (*admissioncontroller.Result, error)
			veradcoPlugin.Execute(meta.TypeMeta{}, pod, r)
		}















		

		for _, c := range pod.Spec.Containers {
			if strings.HasSuffix(c.Image, ":latest") {
				return &admissioncontroller.Result{Msg: "You cannot use the tag 'latest' in a container."}, nil
			}
		}

		return &admissioncontroller.Result{Allowed: true}, nil
	}
}

func mutateCreate() admissioncontroller.AdmitFunc {
	return func(r *admission.AdmissionRequest) (*admissioncontroller.Result, error) {
		var operations []admissioncontroller.PatchOperation
		pod, err := parsePod(r.Object.Raw)
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
		metadata := map[string]string{"origin": "fromMutation"}
		operations = append(operations, admissioncontroller.AddPatchOperation("/metadata/annotations", metadata))
		return &admissioncontroller.Result{
			Allowed:  true,
			PatchOps: operations,
		}, nil
	}
}
