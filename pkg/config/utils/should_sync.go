package utils

import (
	"context"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/common/kubernetes"
)

func ShouldSync(ctx context.Context, syncerName string, requiredCrds []string, crdList kubernetes.CustomResourceDefinitionList) bool {
	syncerActive := true
	for _, crdName := range requiredCrds {
		if _, err := crdList.Find("", crdName); err != nil {
			syncerActive = false
			break
		}
	}
	if err := RecordActive(ctx, syncerName, syncerActive); err != nil {
		contextutils.LoggerFrom(ctx).Errorf("failed recording active stat")
	}
	return syncerActive
}
