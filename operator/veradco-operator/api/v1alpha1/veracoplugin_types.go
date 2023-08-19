/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// VeracoPluginSpec defines the desired state of VeracoPlugin
type VeracoPluginSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of VeracoPlugin. Edit veracoplugin_types.go to remove/update
	Name          string              `json:"name"`
	Path          string              `json:"path"`
	Code          string              `json:"code"`
	CodeSignature string              `json:"codeSignature,omitempty"`
	Kinds         string              `json:"kinds"`
	Operations    string              `json:"operations"`
	Namespaces    string              `json:"namespaces"`
	Labels        []map[string]string `json:"labels,omitempty"`
	Annotations   []map[string]string `json:"annotations,omitempty"`
	DryRun        bool                `json:"dryRun"`
	Configuration string              `json:"configuration"`
	Scope         string              `json:"scope"`
	Endpoints     string              `json:"endpoints,omitempty"`
}

// VeracoPluginStatus defines the observed state of VeracoPlugin
type VeracoPluginStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// How many executions
	PluginCalls int32 `json:"pluginCalls"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// VeracoPlugin is the Schema for the veracoplugins API
type VeracoPlugin struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VeracoPluginSpec   `json:"spec,omitempty"`
	Status VeracoPluginStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// VeracoPluginList contains a list of VeracoPlugin
type VeracoPluginList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VeracoPlugin `json:"items"`
}

func init() {
	SchemeBuilder.Register(&VeracoPlugin{}, &VeracoPluginList{})
}
