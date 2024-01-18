package admissioncontroller

import (
	"fmt"

	admission "k8s.io/api/admission/v1"

	log "k8s.io/klog/v2"
)

// Result contains the result of an admission request
type Result struct {
	Allowed  bool `json:"allowed"`
	Msg      string `json:"msg"`
	PatchOps []PatchOperation `json:"patchOps,omitempty"`
}

// AdmitFunc defines how to process an admission request
type AdmitFunc func(body *[]byte, request *admission.AdmissionRequest) (*Result, error)

// Hook represents the set of functions for each operation in an admission webhook.
type Hook struct {
	Create  AdmitFunc
	Delete  AdmitFunc
	Update  AdmitFunc
	Connect AdmitFunc
}

// Execute evaluates the request and try to execute the function for operation specified in the request.
func (h *Hook) Execute(body *[]byte, r *admission.AdmissionRequest) (*Result, error) {
	switch r.Operation {
	case admission.Create:
		return wrapperExecution(h.Create, body, r)
	case admission.Update:
		return wrapperExecution(h.Update, body, r)
	case admission.Delete:
		return wrapperExecution(h.Delete, body, r)
	case admission.Connect:
		return wrapperExecution(h.Connect, body, r)
	}

	return &Result{Msg: fmt.Sprintf("Invalid operation: %s", r.Operation)}, nil
}

func wrapperExecution(fn AdmitFunc, body *[]byte, r *admission.AdmissionRequest) (*Result, error) {
	if fn == nil {
		log.Infof(">>>> wrapperExecution: operation %s is not registered", r.Operation)
		return nil, fmt.Errorf("operation %s is NOT registered", r.Operation)
	}

	log.V(1).Infof(">>>> wrapperExecution: operation %s is registered", r.Operation)
	return fn(body, r)
}
