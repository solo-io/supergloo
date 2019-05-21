package metrics

import (
	"path"
)

const APIVersion = "metrics.smi-spec.io/v1alpha1"

func baseURL() string {
	return path.Join("/apis", APIVersion)
}
