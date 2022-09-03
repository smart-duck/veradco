package deployments

import (
	"github.com/smart-duck/veradco"

	"github.com/smart-duck/veradco/cfg"

	admission "k8s.io/api/admission/v1"
)

func validateDelete(veradcoCfg *conf.VeradcoCfg) admissioncontroller.AdmitFunc {
	return func(r *admission.AdmissionRequest) (*admissioncontroller.Result, error) {
		dp, err := parseDeployment(r.OldObject.Raw)
		if err != nil {
			return &admissioncontroller.Result{Msg: err.Error()}, nil
		}

		if dp.Namespace == "special-system" && dp.Annotations["skip"] == "false" {
			return &admissioncontroller.Result{Msg: "You cannot remove a deployment from `special-system` namespace."}, nil
		}

		return &admissioncontroller.Result{Allowed: true}, nil
	}
}
