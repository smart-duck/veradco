package main

import (
	admission "k8s.io/api/admission/v1"
	// v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"github.com/smart-duck/veradco/admissioncontroller"
	"github.com/smart-duck/veradco/grpc"
	"github.com/smart-duck/veradco/kres"
	"fmt"
)

type Basic struct {
	summary string
}

func (plug *Basic) Init(configFile string) error {
	// plug.summary = fmt.Sprintf("Configuration of plugin %s is: %s", name, configFile)
	fmt.Printf("Init GRPC plugin\n")
	return nil
}


func (plug *Basic) Execute(kobj runtime.Object, operation string, dryRun bool, r *admission.AdmissionRequest) (*admissioncontroller.Result, error) {
	plug.summary = ""

	pod, err := kres.ParsePod(r)
	if err != nil {
		fmt.Printf("Error parsing pod: %v\n", err)
		return &admissioncontroller.Result{Msg: err.Error()}, nil
	} else {
		for i, c := range pod.Spec.Containers {
			fmt.Printf("Container %d: name %s, image %s\n", i, c.Name, c.Image)
			plug.summary += "\n" + fmt.Sprintf("Container %d: name %s, image %s", i, c.Name, c.Image)
		}
	}

	// pod, ok := kobj.(*v1.Pod)
	// if !ok {
	// 	fmt.Printf("Kubernetes resource is not a pod as expected (%s)\n", kobj.GetObjectKind().GroupVersionKind().Kind)
	// 	plug.summary += "\n" + fmt.Sprintf("Kubernetes resource is not a pod as expected (%s)", kobj.GetObjectKind().GroupVersionKind().Kind)
	// 	return nil, fmt.Errorf("Kubernetes resource is not a pod as expected (%s)", kobj.GetObjectKind().GroupVersionKind().Kind)
	// } else {
	// 	fmt.Printf("As expected it is a pod: %s, %s, %s, %s\n", pod.TypeMeta.Kind,
	// 		pod.TypeMeta.APIVersion, pod.ObjectMeta.Name, pod.ObjectMeta.Namespace)
	// 	plug.summary += "\n" + fmt.Sprintf("As expected it is a pod: %s, %s, %s, %s", pod.TypeMeta.Kind,
	// 		pod.TypeMeta.APIVersion, pod.ObjectMeta.Name, pod.ObjectMeta.Namespace)

	// 	for i, c := range pod.Spec.Containers {
	// 		fmt.Printf("Container %d: name %s, image %s\n", i, c.Name, c.Image)
	// 		plug.summary += "\n" + fmt.Sprintf("Container %d: name %s, image %s", i, c.Name, c.Image)
	// 	}
	// }

	fmt.Printf("Pod %s, accepted\n", pod.ObjectMeta.Name)
	plug.summary += "\n" + fmt.Sprintf("Pod %s, accepted", pod.ObjectMeta.Name)

	return &admissioncontroller.Result{Allowed: true}, nil
}

func (plug *Basic) Summary() string {
	return plug.summary
}

func main() {
	plugin := grpc.GrpcPlugin {
		Port: 50053,
		VeradcoPlugin: &Basic{},
	}
	err := plugin.StartServer()
	if err != nil {
		fmt.Printf("Error starting Basic plugin: %v\n", err)
	}
}