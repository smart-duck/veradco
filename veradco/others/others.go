package others

import (
	"encoding/json"

	"github.com/smart-duck/veradco"

	"github.com/smart-duck/veradco/cfg"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	admission "k8s.io/api/admission/v1"

	log "k8s.io/klog/v2"
)

// NewValidationHook creates a new instance of others validation hook
func NewValidationHook(veradcoCfg *conf.VeradcoCfg) admissioncontroller.Hook {
	return admissioncontroller.Hook{
		Create: validateCreate(veradcoCfg),
		Delete: validateDelete(veradcoCfg),
		Update: validateUpdate(veradcoCfg),
		Connect: validateConnect(veradcoCfg),
	}
}

func NewMutationHook(veradcoCfg *conf.VeradcoCfg) admissioncontroller.Hook {
	return admissioncontroller.Hook{
		Create: mutateCreate(veradcoCfg),
		Delete: mutateDelete(veradcoCfg),
		Update: mutateUpdate(veradcoCfg),
		Connect: mutateConnect(veradcoCfg),
	}
}
 
func parseOther(r *admission.AdmissionRequest) (*meta.PartialObjectMetadata, error) {
	var other meta.PartialObjectMetadata

	if err := json.Unmarshal(r.Object.Raw, &other); err != nil {

		// Try with OldObject
		if errOldObj := json.Unmarshal(r.OldObject.Raw, &other); errOldObj != nil {

			// Use previous error on Object
			log.Errorf("Failed to parse (parseOther): %v", err)
			return nil, err
		}

	}

	return &other, nil
}
