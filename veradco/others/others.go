package others

import (
	"github.com/smart-duck/veradco"

	"github.com/smart-duck/veradco/cfg"
)

// NewValidationHook creates a new instance of others validation hook
func NewValidationHook(veradcoCfg *conf.VeradcoCfg, endpoint string) admissioncontroller.Hook {
	return admissioncontroller.Hook{
		Create: validateCreate(veradcoCfg, endpoint),
		Delete: validateDelete(veradcoCfg, endpoint),
		Update: validateUpdate(veradcoCfg, endpoint),
		Connect: validateConnect(veradcoCfg, endpoint),
	}
}

func NewMutationHook(veradcoCfg *conf.VeradcoCfg, endpoint string) admissioncontroller.Hook {
	return admissioncontroller.Hook{
		Create: mutateCreate(veradcoCfg, endpoint),
		Delete: mutateDelete(veradcoCfg, endpoint),
		Update: mutateUpdate(veradcoCfg, endpoint),
		Connect: mutateConnect(veradcoCfg, endpoint),
	}
}

