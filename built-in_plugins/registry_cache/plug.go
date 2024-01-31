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
	name string = "RegistryCache"
)

type FindReplace struct {
	Find string `yaml:"find"`
	Replace string `yaml:"replace"`
}

type RegistryCache struct {
	Replacements []FindReplace `yaml:"replacements"`
	summary string `yaml:"-"`
}

func (plug *RegistryCache) Init(configFile string) error {
	// Load configuration
	err := yaml.Unmarshal([]byte(configFile), plug)
	if err != nil {
		return err
	}
	if len(plug.Replacements) == 0 {
		return fmt.Errorf("Replacements list shall contain at least one element for plugin %s", name)
	}
	return nil
}


func (plug *RegistryCache) Execute(kobj runtime.Object, operation string, dryRun bool, r *admission.AdmissionRequest) (*admissioncontroller.Result, error) {

	// kobj is supposed to be a pod...
	pod, ok := kobj.(*v1.Pod)
	if !ok {
		plug.summary += fmt.Sprintf("Kubernetes resource is not a pod as expected (%s)", kobj.GetObjectKind().GroupVersionKind().Kind)
		return nil, fmt.Errorf("Kubernetes resource is not a pod as expected (%s)", kobj.GetObjectKind().GroupVersionKind().Kind)
	}

	var operations []admissioncontroller.PatchOperation

	plug.summary = fmt.Sprintf("Execute plugin %s", name)

	// Browse containers
	for i, c := range pod.Spec.Containers {
		// Browse replacements
		for _, op := range plug.Replacements {
			find := op.Find
			replace := op.Replace
			
			re := regexp.MustCompile(find)
			img := c.Image
			if re.MatchString(img) {
				imgNew := re.ReplaceAllString(img, replace)
				// add replace patch
				replaceOp := admissioncontroller.ReplacePatchOperation(fmt.Sprintf("/spec/containers/%d/image", i), imgNew)
				operations = append(operations, replaceOp)
				plug.summary += "\n" + fmt.Sprintf("Add repacement operation %v", replaceOp)
				break
			}
		}
	}

	return &admissioncontroller.Result{
		Allowed:  true,
		PatchOps: operations,
	}, nil
}

func (plug *RegistryCache) Summary() string {
	return plug.summary
}

// exported as symbol named "VeradcoPlugin"
var VeradcoPlugin RegistryCache