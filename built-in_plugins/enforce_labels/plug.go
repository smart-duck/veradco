package main

import (
	admission "k8s.io/api/admission/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"github.com/smart-duck/veradco/veradco/admissioncontroller"
	"github.com/smart-duck/veradco/veradco/kres"
	"fmt"
	"gopkg.in/yaml.v3"
	"regexp"
	// meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	name string = "EnforceLabels"
)

type EnforceLabels struct {
	Annotations map[string]string `yaml:"annotations,omitempty"`
	Labels map[string]string `yaml:"labels,omitempty"`
	summary string `yaml:"-"`
}

func (plug *EnforceLabels) Init(configFile string) error {
	// Load configuration
	err := yaml.Unmarshal([]byte(configFile), plug)
	if err != nil {
		// plug.summary = fmt.Sprintf("Cannot unmarshal configuration: %v", err)
		return err
	}
	return nil
}


func (plug *EnforceLabels) Execute(kobj runtime.Object, operation string, dryRun bool, r *admission.AdmissionRequest) (*admissioncontroller.Result, error) {

	plug.summary = ""

	obj, err := kres.ParseOther(r)

	if err != nil {
		return nil, err
	}

	// obj, ok := kobj.(*meta.PartialObjectMetadata)
	// if !ok {
	// 	plug.summary += "\n" + fmt.Sprintf("Kubernetes resource is not as expected (%s)", kobj.GetObjectKind().GroupVersionKind().Kind)
	// 	return nil, fmt.Errorf("Kubernetes resource is not as expected (%s)", kobj.GetObjectKind().GroupVersionKind().Kind)
	// }

	for k := range plug.Annotations {
		annot, exists := obj.ObjectMeta.Annotations[k]
		if ! exists {
			msg := fmt.Sprintf("Resource %s: annotation %s shall be defined!", obj.ObjectMeta.Name, k)
			plug.summary += "\n" + msg
			return &admissioncontroller.Result{Msg: msg}, nil
		} else {
			matched, err := regexp.MatchString(plug.Annotations[k], annot)
			if err != nil {
				plug.summary += "\n" + fmt.Sprintf("Evaluate regex %s on %s failed: %v\n", plug.Annotations[k], annot, err)
				return nil, err
			}
			if ! matched {
				msg := fmt.Sprintf("Resource %s: annotation %s shall match the regex %s!", obj.ObjectMeta.Name, k, plug.Annotations[k])
				plug.summary += "\n" + msg
				return &admissioncontroller.Result{Msg: msg}, nil
			}
		}
	}

	for k := range plug.Labels {
		label, exists := obj.ObjectMeta.Labels[k]
		if ! exists {
			msg := fmt.Sprintf("Resource %s: label %s shall be defined!", obj.ObjectMeta.Name, k)
			plug.summary += "\n" + msg
			return &admissioncontroller.Result{Msg: msg}, nil
		} else {
			matched, err := regexp.MatchString(plug.Labels[k], label)
			if err != nil {
				plug.summary += "\n" + fmt.Sprintf("Evaluate regex %s on %s failed: %v\n", plug.Labels[k], label, err)
				return nil, err
			}
			if ! matched {
				msg := fmt.Sprintf("Resource %s: label %s shall match the regex %s!", obj.ObjectMeta.Name, k, plug.Labels[k])
				plug.summary += "\n" + msg
				return &admissioncontroller.Result{Msg: msg}, nil
			}
		}
	}	

	plug.summary += "\n" + fmt.Sprintf("Resource %s, accepted", obj.ObjectMeta.Name)

	return &admissioncontroller.Result{Allowed: true}, nil
}

func (plug *EnforceLabels) Summary() string {
	return plug.summary
}

// exported as symbol named "VeradcoPlugin"
var VeradcoPlugin EnforceLabels