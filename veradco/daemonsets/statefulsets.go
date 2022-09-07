package daemonsets

import (
	"github.com/smart-duck/veradco"

	"github.com/smart-duck/veradco/cfg"
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

