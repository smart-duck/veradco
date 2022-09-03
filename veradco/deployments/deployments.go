package deployments

import (
	"encoding/json"

	"github.com/smart-duck/veradco"

	v1 "k8s.io/api/apps/v1"

	"github.com/smart-duck/veradco/cfg"
)

// NewValidationHook creates a new instance of deployment validation hook
func NewValidationHook(veradcoCfg *conf.VeradcoCfg) admissioncontroller.Hook {
	return admissioncontroller.Hook{
		Create: validateCreate(veradcoCfg),
		Delete: validateDelete(veradcoCfg),
	}
}

func parseDeployment(object []byte) (*v1.Deployment, error) {
	var dp v1.Deployment
	if err := json.Unmarshal(object, &dp); err != nil {
		return nil, err
	}

	return &dp, nil
}
