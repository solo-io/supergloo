package labelutils

import (
	"fmt"

	"github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1alpha2"
	"github.com/solo-io/gloo-mesh/pkg/common/defaults"
)

// the key used to differentiate discovery resources by
// the cluster in which they were discovered
var ClusterLabelKey = fmt.Sprintf("cluster.%s", v1alpha2.SchemeGroupVersion.Group)
var OwnershipLabelKey = fmt.Sprintf("owner.%s", v1alpha2.SchemeGroupVersion.Group)

// Create a label that identifies the cluster used to discover a resource.
func ClusterLabels(cluster string) map[string]string {
	clusterK, clusterV := ClusterLabel(cluster)
	labels := OwnershipLabels()
	labels[clusterK] = clusterV
	return labels
}

func ClusterLabel(cluster string) (string, string) {
	return ClusterLabelKey,
		fmt.Sprintf("%s", cluster)
}

// identifies the instance of gloo-mesh discovery that produced the resource.
// uses pod namespace to identify the instance.
// if running in an agent, (AGENT_CLUSTER is set)
func OwnershipLabels() map[string]string {
	ownerId := defaults.GetPodNamespace()
	if agentCluster := defaults.GetAgentCluster(); agentClsuter != "" {
		ownerId += "-" + agentClsuter
	}
	return map[string]string{
		fmt.Sprintf(OwnershipLabelKey): ownerId,
	}
}
