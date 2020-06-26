package utils

import (
	skv1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	"github.com/solo-io/skv2/pkg/ezkube"
)

func MakeResourceRef(id ezkube.ResourceId) *skv1.ObjectRef {
	return &skv1.ObjectRef{
		Name:      id.GetName(),
		Namespace: id.GetNamespace(),
	}
}

func MakeClusterResourceRef(id ezkube.ResourceId) *skv1.ClusterObjectRef {
	return &skv1.ClusterObjectRef{
		Name:        id.GetName(),
		Namespace:   id.GetNamespace(),
		ClusterName: id.GetClusterName(),
	}
}
