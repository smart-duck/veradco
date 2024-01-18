package daemonsets

import (
	"github.com/smart-duck/veradco/admissioncontroller"

	"github.com/smart-duck/veradco/kres"

	"github.com/smart-duck/veradco/cfg"

	log "k8s.io/klog/v2"

	admission "k8s.io/api/admission/v1"

	"k8s.io/apimachinery/pkg/runtime"
)

func validateCreate(veradcoCfg *conf.VeradcoCfg, endpoint string) admissioncontroller.AdmitFunc {
	return operation(admission.Create, veradcoCfg, "Validating", endpoint)
}

func validateUpdate(veradcoCfg *conf.VeradcoCfg, endpoint string) admissioncontroller.AdmitFunc {
	return operation(admission.Update, veradcoCfg, "Validating", endpoint)
}

func validateDelete(veradcoCfg *conf.VeradcoCfg, endpoint string) admissioncontroller.AdmitFunc {
	return operation(admission.Delete, veradcoCfg, "Validating", endpoint)
}

func validateConnect(veradcoCfg *conf.VeradcoCfg, endpoint string) admissioncontroller.AdmitFunc {
	return operation(admission.Connect, veradcoCfg, "Validating", endpoint)
}

func mutateCreate(veradcoCfg *conf.VeradcoCfg, endpoint string) admissioncontroller.AdmitFunc {
	return operation(admission.Create, veradcoCfg, "Mutating", endpoint)
}

func mutateUpdate(veradcoCfg *conf.VeradcoCfg, endpoint string) admissioncontroller.AdmitFunc {
	return operation(admission.Update, veradcoCfg, "Mutating", endpoint)
}

func mutateDelete(veradcoCfg *conf.VeradcoCfg, endpoint string) admissioncontroller.AdmitFunc {
	return operation(admission.Delete, veradcoCfg, "Mutating", endpoint)
}

func mutateConnect(veradcoCfg *conf.VeradcoCfg, endpoint string) admissioncontroller.AdmitFunc {
	return operation(admission.Connect, veradcoCfg, "Mutating", endpoint)
}

func operation(op admission.Operation, veradcoCfg *conf.VeradcoCfg, scope string, endpoint string) admissioncontroller.AdmitFunc {
	return func(body *[]byte, r *admission.AdmissionRequest) (*admissioncontroller.Result, error) {

		log.V(1).Infof(">>>> daemonsets / %s operation, Kind: %s, Version: %s, Group: %s, Name: %s, Namespace: %s", string(op), r.Kind.Kind, r.Kind.Version, r.Kind.Group, r.Name, r.Namespace)

		var ds runtime.Object
		var err error

		// Should be a pod
		ds, err = kres.ParseDaemonSet(r)
		if err != nil {
			return &admissioncontroller.Result{Msg: err.Error()}, nil
		}

		// Apply the plugins
		return veradcoCfg.ProceedPlugins(body, ds, r, scope, endpoint)
	}
}