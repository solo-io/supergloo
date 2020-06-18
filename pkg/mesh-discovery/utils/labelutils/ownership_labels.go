package labelutils

import (
	"fmt"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
)

// Create a label that identifies the cluster used to discover a resource.
func ClusterLabels(cluster string) map[string]string {
	k, v := ClusterLabel(cluster)
	return map[string]string{k: v}
}

func ClusterLabel(cluster string) (string, string) {
	return fmt.Sprintf("cluster.%s", v1alpha1.SchemeGroupVersion.Group),
		fmt.Sprintf("%s", cluster)
}
