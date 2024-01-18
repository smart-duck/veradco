package main

import (
	admission "k8s.io/api/admission/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"github.com/smart-duck/veradco/admissioncontroller"
	"github.com/smart-duck/veradco/kres"
	"fmt"
)

var (
	name string = "Generic"
	calls int = 1
)

type Generic struct {
	summary string
}

func (plug *Generic) Init(configFile string) error {
	// plug.summary = fmt.Sprintf("Configuration of plugin %s is: %s", name, configFile)
	// plug.summary += "\n" + fmt.Sprintf("[%T] %+v %p", plug, plug, plug)
	return nil
}


func (plug *Generic) Execute(kobj runtime.Object, operation string, dryRun bool, r *admission.AdmissionRequest) (*admissioncontroller.Result, error) {
	calls++
	plug.summary = ""
	// plug.summary += "\n" + fmt.Sprintf("Generic: call nb %d", calls)
	
	var (
		obj *meta.PartialObjectMetadata
		ok bool
		err error
	)

	obj, ok = kobj.(*meta.PartialObjectMetadata)

	if !ok {

		obj, err = kres.ParseOther(r)

		if err != nil {
			return nil, err
		}

	}

	plug.summary += "\n" + fmt.Sprintf("Generic resource: %s %s/%s %s ns:%s", operation, obj.TypeMeta.Kind, obj.TypeMeta.APIVersion, obj.ObjectMeta.Name, obj.ObjectMeta.Namespace)

	// if !ok {
	// 	plug.summary += "\n" + fmt.Sprintf("Kubernetes resource is not as expected (%s)", kobj.GetObjectKind().GroupVersionKind().Kind)
	// 	return nil, fmt.Errorf("Kubernetes resource is not as expected (%s)", kobj.GetObjectKind().GroupVersionKind().Kind)
	// } else {
	// 	plug.summary += "\n" + fmt.Sprintf("Generic resource: %s, %s, %s, %s", other.TypeMeta.Kind,
	// 		other.TypeMeta.APIVersion, other.ObjectMeta.Name, other.ObjectMeta.Namespace)
	// }

	// plug.summary += "\n" + fmt.Sprintf("%s %s, accepted", obj.TypeMeta.Kind, obj.ObjectMeta.Name)

	return &admissioncontroller.Result{Allowed: true}, nil
}

func (plug *Generic) Summary() string {
	return plug.summary
}

// exported as symbol named "VeradcoPlugin"
var VeradcoPlugin Generic