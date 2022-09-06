package kres

import (
	"encoding/json"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	admission "k8s.io/api/admission/v1"
	v1 "k8s.io/api/core/v1"
	log "k8s.io/klog/v2"
)

func parseTypeMeta(object []byte) (*meta.TypeMeta, error) {
	var meta meta.TypeMeta
	if err := json.Unmarshal(object, &meta); err != nil {
		return nil, err
	}

	return &meta, nil
}

func parseObjectMeta(object []byte) (*meta.ObjectMeta, error) {
	var meta meta.ObjectMeta
	if err := json.Unmarshal(object, &meta); err != nil {
		return nil, err
	}

	return &meta, nil
}

func ParsePod(r *admission.AdmissionRequest) (*v1.Pod, error) {
	var pod v1.Pod

	if err := json.Unmarshal(r.Object.Raw, &pod); err != nil {

		// Try with OldObject
		if errOldObj := json.Unmarshal(r.OldObject.Raw, &pod); errOldObj != nil {

			// Use previous error on Object
			log.Errorf("Failed to parse (parsePod): %v", err)
			return nil, err
		}

	}

	return &pod, nil
}