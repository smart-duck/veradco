package pods

import (
	"encoding/json"

	"github.com/smart-duck/veradco"

	"github.com/smart-duck/veradco/cfg"

	// "github.com/smart-duck/veradco/kres"

	v1 "k8s.io/api/core/v1"

	// log "k8s.io/klog/v2"

	// "errors"
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

// func parseMeta(object []byte) (*meta.TypeMeta, error) {
// 	var meta meta.TypeMeta
// 	if err := json.Unmarshal(object, &meta); err != nil {
// 		return nil, err
// 	}

// 	return &meta, nil
// }

func parsePod(object []byte) (*v1.Pod, error) {
	// meta, err := kres.parseMeta(object)

	// if err != nil {
	// 	log.Infof("Error unmarshal meta: %v", err)
	// } else {
	// 	log.Infof("Meta: %v", meta)
	// 	if meta.Kind != "Pod" {
	// 		return nil, errors.New("Not a Pod")
	// 	}
		
	// }

	var pod v1.Pod
	if err := json.Unmarshal(object, &pod); err != nil {
		return nil, err
	}

	return &pod, nil
}
