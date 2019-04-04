package util

import (
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
)

func LabelsForResource(res resources.Resource) map[string]string {
	return map[string]string{resources.Kind(res): res.GetMetadata().Ref().Key()}
}
