package dns

import (
	"fmt"

	core_types "github.com/solo-io/mesh-projects/pkg/api/core.zephyr.solo.io/v1alpha1/types"
)

// Given a ref to a *KUBE* service, produce a DNS name for it that can resolve across federated clusters
func BuildMulticlusterDnsName(kubeServiceRef *core_types.ResourceRef, serviceClusterName string) string {
	return fmt.Sprintf("%s.%s.%s", kubeServiceRef.GetName(), kubeServiceRef.GetNamespace(), serviceClusterName)
}
