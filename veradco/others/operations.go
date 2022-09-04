package others

import (
	"github.com/smart-duck/veradco"

	"github.com/smart-duck/veradco/cfg"

	log "k8s.io/klog/v2"

	admission "k8s.io/api/admission/v1"

	"k8s.io/apimachinery/pkg/runtime"
)

func validateCreate(veradcoCfg *conf.VeradcoCfg) admissioncontroller.AdmitFunc {

	return defineOperation(admission.Create, veradcoCfg)

	// return func(r *admission.AdmissionRequest) (*admissioncontroller.Result, error) {

	// 	log.Infof(">>>> others / validateCreate")

	// 	var other runtime.Object
	// 	var err error

	// 	// Should be a *meta.PartialObjectMetadata
	// 	other, err = parseOther(r.Object.Raw)
	// 	if err != nil {
	// 		return &admissioncontroller.Result{Msg: err.Error()}, nil
	// 	}

	// 	// Apply the plugins
	// 	return veradcoCfg.ProceedPlugins(other, r, "Validating")
	// }
}

func validateUpdate(veradcoCfg *conf.VeradcoCfg) admissioncontroller.AdmitFunc {
	return defineOperation(admission.Update, veradcoCfg)
}

func validateDelete(veradcoCfg *conf.VeradcoCfg) admissioncontroller.AdmitFunc {
	return defineOperation(admission.Delete, veradcoCfg)
}

func validateConnect(veradcoCfg *conf.VeradcoCfg) admissioncontroller.AdmitFunc {
	return defineOperation(admission.Connect, veradcoCfg)
}

func defineOperation(op admission.Operation, veradcoCfg *conf.VeradcoCfg) admissioncontroller.AdmitFunc {
	return func(r *admission.AdmissionRequest) (*admissioncontroller.Result, error) {

		log.Infof(">>>> others / %s operation, Kind: %s, Version: %s, Group: %s, Name: %s, Namespace: %s", string(op), r.Kind.Kind, r.Kind.Version, r.Kind.Group, r.Name, r.Namespace)

		var other runtime.Object
		var err error

		// Should be a *meta.PartialObjectMetadata
		other, err = parseOther(r)
		if err != nil {
			return &admissioncontroller.Result{Msg: err.Error()}, nil
		}

		// Apply the plugins
		return veradcoCfg.ProceedPlugins(other, r, "Validating")
	}
}