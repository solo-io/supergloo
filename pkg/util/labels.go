package util

import (
	"strings"

	"github.com/solo-io/go-utils/kubeutils"

	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
)

func LabelsForResource(res resources.Resource) map[string]string {
	labelKey := strings.TrimPrefix(kubeutils.SanitizeName(resources.Kind(res)), "-")
	return map[string]string{labelKey: res.GetMetadata().Ref().Key()}
}
