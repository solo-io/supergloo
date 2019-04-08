package util

import (
	"strings"

	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/supergloo/pkg/translator/utils"
)

func LabelsForResource(res resources.Resource) map[string]string {
	labelKey := strings.TrimPrefix(utils.SanitizeName(resources.Kind(res)), "-")
	return map[string]string{labelKey: res.GetMetadata().Ref().Key()}
}
