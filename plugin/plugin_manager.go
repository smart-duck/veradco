package plugin

import (
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	admission "k8s.io/api/admission/v1"
	"github.com/smart-duck/veradco"
)

type VeradcoPlugin interface {
	Init(configFile string)
	Execute(meta meta.TypeMeta, kobj interface{}, r *admission.AdmissionRequest) (*admissioncontroller.Result, error)
}