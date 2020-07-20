package labelutils

import (
	"fmt"

	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/common/defaults"
)

// the key used to differentiate discovery resources by
// the cluster in which they were discovered
var ClusterLabelKey = fmt.Sprintf("cluster.%s", v1alpha1.SchemeGroupVersion.Group)

// Create a label that identifies the cluster used to discover a resource.
func ClusterLabels(cluster string) map[string]string {
	clusterK, clusterV := ClusterLabel(cluster)
	ownerK, ownerV := OwnershipLabel()
	return map[string]string{
		clusterK: clusterV,
		ownerK:   ownerV,
	}
}

func ClusterLabel(cluster string) (string, string) {
	return ClusterLabelKey,
		fmt.Sprintf("%s", cluster)
}

// identifies the instance of service-mesh-hub discovery that produced the resource.
// uses pod namespace to identify the instance
func OwnershipLabel() (string, string) {
	return fmt.Sprintf("owner.%s", v1alpha1.SchemeGroupVersion.Group),
		defaults.GetPodNamespace()
}
