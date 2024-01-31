package http

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/smart-duck/veradco/veradco/admissioncontroller"

	// "k8s.io/api/admission/v1beta1"
	admission "k8s.io/api/admission/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	log "k8s.io/klog/v2"
)

// admissionHandler represents the HTTP handler for an admission webhook
type admissionHandler struct {
	decoder runtime.Decoder
}

// newAdmissionHandler returns an instance of AdmissionHandler
func newAdmissionHandler() *admissionHandler {
	return &admissionHandler{
		decoder: serializer.NewCodecFactory(runtime.NewScheme()).UniversalDeserializer(),
	}
}

// Serve returns a http.HandlerFunc for an admission webhook
func (h *admissionHandler) Serve(hook admissioncontroller.Hook) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			http.Error(w, fmt.Sprint("invalid method only POST requests are allowed"), http.StatusMethodNotAllowed)
			return
		}

		if contentType := r.Header.Get("Content-Type"); contentType != "application/json" {
			http.Error(w, fmt.Sprint("only content type 'application/json' is supported"), http.StatusBadRequest)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, fmt.Sprintf("could not read request body: %v", err), http.StatusBadRequest)
			return
		}

		var review admission.AdmissionReview
		if _, _, err := h.decoder.Decode(body, nil, &review); err != nil {
			http.Error(w, fmt.Sprintf("could not deserialize request: %v", err), http.StatusBadRequest)
			return
		}

		if review.Request == nil {
			http.Error(w, "malformed admission review: request is nil", http.StatusBadRequest)
			return
		}

		result, err := hook.Execute(&body, review.Request)
		if err != nil {
			log.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		var msgResult string
		var allowResult bool

		if result == nil {
			msgResult = ""
			allowResult = true
		} else {
			msgResult = result.Msg
			allowResult = result.Allowed
		}

		admissionResponse := admission.AdmissionReview{
			TypeMeta: meta.TypeMeta{APIVersion: "admission.k8s.io/v1", Kind: "AdmissionReview"},
			Response: &admission.AdmissionResponse{
				UID:     review.Request.UID,
				Allowed: allowResult,
				Result:  &meta.Status{Message: msgResult},
			},
		}

		if result != nil {
			// set the patch operations for mutating admission
			if len(result.PatchOps) > 0 {
				log.V(2).Infof("Set patch operations: %v", result.PatchOps)
				patchBytes, err := json.Marshal(result.PatchOps)
				if err != nil {
					log.Error(err)
					http.Error(w, fmt.Sprintf("could not marshal JSON patch: %v", err), http.StatusInternalServerError)
				}
				admissionResponse.Response.Patch = patchBytes
				v1JSONPatch := admission.PatchTypeJSONPatch
				admissionResponse.Response.PatchType = &v1JSONPatch
				log.V(3).Infof("admissionResponse.Response.Patch: %s", string(admissionResponse.Response.Patch))
			}
		}

		res, err := json.Marshal(admissionResponse)
		if err != nil {
			log.Error(err)
			http.Error(w, fmt.Sprintf("could not marshal response: %v", err), http.StatusInternalServerError)
			return
		}

		log.V(4).Infof("Webhook %s response: %c", r.URL.Path, res)

		log.V(1).Infof("Webhook [%s - %s] - Allowed: %t", r.URL.Path, review.Request.Operation, allowResult)
		w.WriteHeader(http.StatusOK)
		w.Write(res)
	}
}

func healthz() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}
}
