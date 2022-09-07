package others

import (
	"github.com/smart-duck/veradco"

	"github.com/smart-duck/veradco/cfg"

	log "k8s.io/klog/v2"

	admission "k8s.io/api/admission/v1"

	"k8s.io/apimachinery/pkg/runtime"
)

func validateCreate(veradcoCfg *conf.VeradcoCfg) admissioncontroller.AdmitFunc {
	return validatingOperation(admission.Create, veradcoCfg, "Validating")
}

func validateUpdate(veradcoCfg *conf.VeradcoCfg) admissioncontroller.AdmitFunc {
	return validatingOperation(admission.Update, veradcoCfg, "Validating")
}

func validateDelete(veradcoCfg *conf.VeradcoCfg) admissioncontroller.AdmitFunc {
	return validatingOperation(admission.Delete, veradcoCfg, "Validating")
}

func validateConnect(veradcoCfg *conf.VeradcoCfg) admissioncontroller.AdmitFunc {
	return validatingOperation(admission.Connect, veradcoCfg, "Validating")
}

func mutateCreate(veradcoCfg *conf.VeradcoCfg) admissioncontroller.AdmitFunc {
	return validatingOperation(admission.Create, veradcoCfg, "Mutating")
}

func mutateUpdate(veradcoCfg *conf.VeradcoCfg) admissioncontroller.AdmitFunc {
	return validatingOperation(admission.Update, veradcoCfg, "Mutating")
}

func mutateDelete(veradcoCfg *conf.VeradcoCfg) admissioncontroller.AdmitFunc {
	return validatingOperation(admission.Delete, veradcoCfg, "Mutating")
}

func mutateConnect(veradcoCfg *conf.VeradcoCfg) admissioncontroller.AdmitFunc {
	return validatingOperation(admission.Connect, veradcoCfg, "Mutating")
}

func validatingOperation(op admission.Operation, veradcoCfg *conf.VeradcoCfg, scope string) admissioncontroller.AdmitFunc {
	return func(r *admission.AdmissionRequest) (*admissioncontroller.Result, error) {

		log.V(1).Infof(">>>> others / %s operation, Kind: %s, Version: %s, Group: %s, Name: %s, Namespace: %s", string(op), r.Kind.Kind, r.Kind.Version, r.Kind.Group, r.Name, r.Namespace)

		var other runtime.Object
		var err error

		// Should be a *meta.PartialObjectMetadata
		other, err = parseOther(r)
		if err != nil {
			return &admissioncontroller.Result{Msg: err.Error()}, nil
		}

		// Apply the plugins
		return veradcoCfg.ProceedPlugins(other, r, scope)
	}
}