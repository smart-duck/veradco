package deployments

import (
	"github.com/smart-duck/veradco"

	admission "k8s.io/api/admission/v1"

	"github.com/smart-duck/veradco/cfg"
)

func validateCreate(veradcoCfg *conf.VeradcoCfg) admissioncontroller.AdmitFunc {
	return func(r *admission.AdmissionRequest) (*admissioncontroller.Result, error) {
		dp, err := parseDeployment(r)
		if err != nil {
			return &admissioncontroller.Result{Msg: err.Error()}, nil
		}

		if dp.Namespace == "special" {
			return &admissioncontroller.Result{Msg: "You cannot create a deployment in `special` namespace."}, nil
		}

		return &admissioncontroller.Result{Allowed: true}, nil
	}
}
