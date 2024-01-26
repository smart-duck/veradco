package plugin

import (
	// meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	admission "k8s.io/api/admission/v1"
	"github.com/smart-duck/veradco/admissioncontroller"
)

type VeradcoPlugin interface {
	Init(configFile string) error
	Execute(kobj runtime.Object, operation string, dryRun bool, r *admission.AdmissionRequest) (*admissioncontroller.Result, error)
	Discover() ([] byte, error) // For GRPC only
	Summary() string
}