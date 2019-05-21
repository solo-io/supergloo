package utils

import (
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
)

// set labels on all resources, required for reconcilers
func SetLabels(ownerLabels map[string]string, rcs ...resources.Resource) {
	for _, res := range rcs {
		resources.UpdateMetadata(res, func(meta *core.Metadata) {
			if meta.Labels == nil {
				meta.Labels = make(map[string]string)
			}
			for k, v := range ownerLabels {
				meta.Labels[k] = v
			}
		})
	}
}
