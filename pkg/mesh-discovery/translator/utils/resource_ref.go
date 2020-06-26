package utils

import (
	"github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1/types"
	"github.com/solo-io/skv2/pkg/ezkube"
)

func MakeResourceRef(id ezkube.ResourceId) *types.ResourceRef {
	return &types.ResourceRef{
		Name:      id.GetName(),
		Namespace: id.GetNamespace(),
		Cluster:   id.GetClusterName(),
	}
}
