package deployments

import (
	"encoding/json"

	"github.com/smart-duck/veradco"

	v1 "k8s.io/api/apps/v1"

	"github.com/smart-duck/veradco/cfg"

	admission "k8s.io/api/admission/v1"

	log "k8s.io/klog/v2"
)

// NewValidationHook creates a new instance of deployment validation hook
func NewValidationHook(veradcoCfg *conf.VeradcoCfg) admissioncontroller.Hook {
	return admissioncontroller.Hook{
		Create: validateCreate(veradcoCfg),
		Delete: validateDelete(veradcoCfg),
	}
}

func parseDeployment(r *admission.AdmissionRequest) (*v1.Deployment, error) {
	var dp v1.Deployment

	if err := json.Unmarshal(r.Object.Raw, &dp); err != nil {

		// Try with OldObject
		if errOldObj := json.Unmarshal(r.OldObject.Raw, &dp); errOldObj != nil {

			// Use previous error on Object
			log.Errorf("Failed to parse (parseDeployment): %v", err)
			return nil, err
		}

	}

	return &dp, nil
}
