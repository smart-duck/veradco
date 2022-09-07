package kres

import (
	"encoding/json"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	admission "k8s.io/api/admission/v1"
	batch "k8s.io/kubernetes/pkg/apis/batch"
	v1 "k8s.io/api/core/v1"
	appsv1 "k8s.io/api/apps/v1"
	log "k8s.io/klog/v2"
)

func ParseTypeMeta(object []byte) (*meta.TypeMeta, error) {
	var meta meta.TypeMeta
	if err := json.Unmarshal(object, &meta); err != nil {
		return nil, err
	}

	return &meta, nil
}

func ParseObjectMeta(object []byte) (*meta.ObjectMeta, error) {
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

func ParseJob(r *admission.AdmissionRequest) (*batch.Job, error) {
	var job batch.Job

	if err := json.Unmarshal(r.Object.Raw, &job); err != nil {

		// Try with OldObject
		if errOldObj := json.Unmarshal(r.OldObject.Raw, &job); errOldObj != nil {

			// Use previous error on Object
			log.Errorf("Failed to parse (parseJob): %v", err)
			return nil, err
		}

	}

	return &job, nil
}

func ParseCronJob(r *admission.AdmissionRequest) (*batch.CronJob, error) {
	var cronJob batch.CronJob

	if err := json.Unmarshal(r.Object.Raw, &cronJob); err != nil {

		// Try with OldObject
		if errOldObj := json.Unmarshal(r.OldObject.Raw, &cronJob); errOldObj != nil {

			// Use previous error on Object
			log.Errorf("Failed to parse (parseCronJob): %v", err)
			return nil, err
		}

	}

	return &cronJob, nil
}

func ParseDeployment(r *admission.AdmissionRequest) (*appsv1.Deployment, error) {
	var deployment appsv1.Deployment

	if err := json.Unmarshal(r.Object.Raw, &deployment); err != nil {

		// Try with OldObject
		if errOldObj := json.Unmarshal(r.OldObject.Raw, &deployment); errOldObj != nil {

			// Use previous error on Object
			log.Errorf("Failed to parse (parseDeployment): %v", err)
			return nil, err
		}

	}

	return &deployment, nil
}

func ParseDaemonSet(r *admission.AdmissionRequest) (*appsv1.DaemonSet, error) {
	var daemonSet appsv1.DaemonSet

	if err := json.Unmarshal(r.Object.Raw, &daemonSet); err != nil {

		// Try with OldObject
		if errOldObj := json.Unmarshal(r.OldObject.Raw, &daemonSet); errOldObj != nil {

			// Use previous error on Object
			log.Errorf("Failed to parse (parseDaemonSet): %v", err)
			return nil, err
		}

	}

	return &daemonSet, nil
}

func ParseStatefulSet(r *admission.AdmissionRequest) (*appsv1.StatefulSet, error) {
	var statefulSet appsv1.StatefulSet

	if err := json.Unmarshal(r.Object.Raw, &statefulSet); err != nil {

		// Try with OldObject
		if errOldObj := json.Unmarshal(r.OldObject.Raw, &statefulSet); errOldObj != nil {

			// Use previous error on Object
			log.Errorf("Failed to parse (parseDeployment): %v", err)
			return nil, err
		}

	}

	return &statefulSet, nil
}

func ParseOther(r *admission.AdmissionRequest) (*meta.PartialObjectMetadata, error) {
	var other meta.PartialObjectMetadata

	if err := json.Unmarshal(r.Object.Raw, &other); err != nil {

		// Try with OldObject
		if errOldObj := json.Unmarshal(r.OldObject.Raw, &other); errOldObj != nil {

			// Use previous error on Object
			log.Errorf("Failed to parse (parseOther): %v", err)
			return nil, err
		}

	}

	return &other, nil
}