package utils

import (
	"fmt"

	discoveryv1alpha2 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1alpha2"
	v1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
)

// CreateCredentialsName creates the credentials name for Limited Trust
func CreateCredentialsName(virtualMesh *v1.ObjectRef) string {
	return fmt.Sprintf("%s-mtls-credential", virtualMesh.Name)
}

// IsIstioInternal checks if a traffic target is an internal istio mesh service
func IsIstioInternal(target *discoveryv1alpha2.TrafficTarget) bool {
	_, ok := target.Spec.GetKubeService().GetLabels()["istio"]
	return ok
}
