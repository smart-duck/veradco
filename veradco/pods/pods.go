package pods

import (
	"encoding/json"

	"github.com/smart-duck/veradco"

	"github.com/smart-duck/veradco/cfg"

	v1 "k8s.io/api/core/v1"

	admission "k8s.io/api/admission/v1"

	log "k8s.io/klog/v2"

)

// NewValidationHook creates a new instance of pods validation hook
func NewValidationHook(veradcoCfg *conf.VeradcoCfg) admissioncontroller.Hook {
	return admissioncontroller.Hook{
		Create: validateCreate(veradcoCfg),
	}
}

// NewMutationHook creates a new instance of pods mutation hook
func NewMutationHook(veradcoCfg *conf.VeradcoCfg) admissioncontroller.Hook {
	return admissioncontroller.Hook{
		Create: mutateCreate(veradcoCfg),
	}
}

func ParsePod(r *admission.AdmissionRequest) (*v1.Pod, error) {
	var pod v1.Pod

	if err := json.Unmarshal(r.Object.Raw, &pod); err != nil {

		// Try with OldObject
		if errOldObj := json.Unmarshal(r.OldObject.Raw, &pod); errOldObj != nil {

			// Use previous error on Object
			log.Errorf("Failed to parse (parsePod): %v", err)
			return nil, err
		}

	}

	return &pod, nil
}
