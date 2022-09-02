package kres

import (
	"encoding/json"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func parseMeta(object []byte) (*meta.TypeMeta, error) {
	var meta meta.TypeMeta
	if err := json.Unmarshal(object, &meta); err != nil {
		return nil, err
	}

	return &meta, nil
}