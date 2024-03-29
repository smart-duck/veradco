package main

import (
	admission "k8s.io/api/admission/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"github.com/smart-duck/veradco/veradco/admissioncontroller"
	"fmt"
	"gopkg.in/yaml.v3"
	"regexp"
	// meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	name string = "ForbidTag"
)

type ForbidTag struct {
	TagToForbid string `yaml:"tagToForbid"`
	summary string `yaml:"-"`
}

func (plug *ForbidTag) Init(configFile string) error {
	// Load configuration
	err := yaml.Unmarshal([]byte(configFile), plug)
	if err != nil {
		return err
	}
	if len(plug.TagToForbid) == 0 {
		return fmt.Errorf("tagToForbid parameter shall be defined for plugin %s", name)
	}
	return nil
}


func (plug *ForbidTag) Execute(kobj runtime.Object, operation string, dryRun bool, r *admission.AdmissionRequest) (*admissioncontroller.Result, error) {

	plug.summary = fmt.Sprintf("Plugin %s tag to forbid is %s", name, plug.TagToForbid)

	// kobj is supposed to be a pod...
	pod, ok := kobj.(*v1.Pod)
	if !ok {
		plug.summary += "\n" + fmt.Sprintf("Kubernetes resource is not a pod as expected (%s)", kobj.GetObjectKind().GroupVersionKind().Kind)
		return nil, fmt.Errorf("Kubernetes resource is not as expected (%s)", kobj.GetObjectKind().GroupVersionKind().Kind)
	}

	// Browse containers
	for _, c := range pod.Spec.Containers {


		matched, err := regexp.MatchString(plug.TagToForbid, c.Image)
		if err != nil {
			plug.summary += "\n" + fmt.Sprintf("Evaluate regex %s on %s failed: %v\n", plug.TagToForbid, c.Image, err)
			return nil, err
		}
		if matched {
			msg := fmt.Sprintf("Container %s is rejected because its image does not fit the regex pattern %s", c.Name, plug.TagToForbid)
			plug.summary += "\n" + msg
			return &admissioncontroller.Result{Msg: msg}, nil
		}
	}

	plug.summary += "\n" + fmt.Sprintf("Pod %s, accepted", pod.ObjectMeta.Name)

	return &admissioncontroller.Result{Allowed: true}, nil
}

func (plug *ForbidTag) Summary() string {
	return plug.summary
}

// exported as symbol named "VeradcoPlugin"
var VeradcoPlugin ForbidTag