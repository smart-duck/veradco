package pods

import (
	// "strings"

	// "plugin"

	"github.com/smart-duck/veradco"

	"github.com/smart-duck/veradco/cfg"

	// veradcoplugin "github.com/smart-duck/veradco/plugin"

	log "k8s.io/klog/v2"

	admission "k8s.io/api/admission/v1"
	v1 "k8s.io/api/core/v1"

	// meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func validateCreate(veradcoCfg *conf.VeradcoCfg) admissioncontroller.AdmitFunc {
	return func(r *admission.AdmissionRequest) (*admissioncontroller.Result, error) {

		log.Infof(">>>> validateCreate")

		pod, err := ParsePod(r)
		if err != nil {
			return &admissioncontroller.Result{Msg: err.Error()}, nil
		}




		// Ask for relative plugins
		// plugins, err := veradcoCfg.GetPlugins(r, pod.TypeMeta.Kind, "CREATE", pod.ObjectMeta.Namespace, pod.ObjectMeta.Labels, pod.ObjectMeta.Annotations, "Validating")
		// plugins, err := veradcoCfg.GetPlugins(r, "Validating")

		// if err != nil {
		// 	log.Errorf("Failed to load plugins: %v", err)
		// 	return &admissioncontroller.Result{Allowed: true}, nil
		// }

		// for _, plug := range *plugins {
		// 	log.Infof(">> Init plugin %s\n", plug.Name)
		// 	// log.Infof("Plug: %v\n", plug)
		// 	// log.Infof("[%T] %+v\n", plug.VeradcoPlugin, plug.VeradcoPlugin)
		// 	plug.VeradcoPlugin.Init(plug.Configuration)
		// 	log.Infof(">> Execute plugin %s\n", plug.Name)
		// 	// Execute(meta meta.TypeMeta, kobj interface{}, r *admission.AdmissionRequest) (*admissioncontroller.Result, error)
		// 	// veradcoPlugin.Execute(meta.TypeMeta{}, pod, r)
		// 	result, err := plug.VeradcoPlugin.Execute(pod, string(r.Operation), *r.DryRun || plug.DryRun, r)
		// 	if err == nil {
		// 		log.Infof(">> Plugin execution summary: %s\n", plug.VeradcoPlugin.Summary())
		// 		if ! result.Allowed {
		// 			return result, err
		// 		}
		// 	} else {
		// 		return result, err
		// 	}
		// }

		// log.Infof("Loading plugin  %s\n", "/app/external_plugins/extplug1.so")

		// // Try to execute plugins
		// plug, err := plugin.Open("/app/external_plugins/extplug1.so")
		// if err != nil {
		// 	log.Errorf("Unable to load plugin: %v\n", err)
		// }

		// pluginHandler, err := plug.Lookup("VeradcoPlugin")
		// if err != nil {
		// 	log.Errorf("Unable to find handler for plugin: %v\n", err)
		// }

		// var veradcoPlugin veradcoplugin.VeradcoPlugin

		// veradcoPlugin, ok := pluginHandler.(veradcoplugin.VeradcoPlugin)
		// if !ok {
		// 	log.Infof("Plugin does not implement awaited interface\n")
		// } else {
		// 	log.Infof("Init plugin\n")
		// 	veradcoPlugin.Init("Path to config file")
		// 	log.Infof("Execute plugin\n")
		// 	// Execute(meta meta.TypeMeta, kobj interface{}, r *admission.AdmissionRequest) (*admissioncontroller.Result, error)
		// 	// veradcoPlugin.Execute(meta.TypeMeta{}, pod, r)
		// 	result, err := veradcoPlugin.Execute(pod, string(r.Operation), *r.DryRun, r)
		// 	if err == nil {
		// 		log.Infof("Plugin execution summary: %s\n", veradcoPlugin.Summary())
		// 	}

		// 	return result, err
			
		// }















		

		// for _, c := range pod.Spec.Containers {
		// 	if strings.HasSuffix(c.Image, ":latest") {
		// 		return &admissioncontroller.Result{Msg: "You cannot use the tag 'latest' in a container."}, nil
		// 	}
		// }

		// return &admissioncontroller.Result{Allowed: true}, nil

		return veradcoCfg.ProceedPlugins(pod, r, "Validating")
	}
}

func mutateCreate(veradcoCfg *conf.VeradcoCfg) admissioncontroller.AdmitFunc {
	return func(r *admission.AdmissionRequest) (*admissioncontroller.Result, error) {
		var operations []admissioncontroller.PatchOperation
		pod, err := ParsePod(r)
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
