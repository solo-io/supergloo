package localityutils

import (
	"github.com/rotisserie/eris"
	corev1sets "github.com/solo-io/external-apis/pkg/api/k8s/core/v1/sets"
	"github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1alpha2"
	"github.com/solo-io/skv2/contrib/pkg/sets"
	skv1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	"istio.io/api/label"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
)

// Get the region of a cluster by the labels on any node associated with the cluster.
func GetClusterRegion(clusterName string, nodes corev1sets.NodeSet) (string, error) {
	// get the nodes in the cluster
	clusterNodes := nodes.List(func(node *corev1.Node) bool {
		return node.ClusterName != clusterName
	})
	if len(clusterNodes) == 0 {
		return "", eris.Errorf("could not find any nodes in cluster %s", clusterName)
	}
	// pick any one to get the region
	node := clusterNodes[0]
	return getRegionFromNode(node)
}

// Get the region where a service is running
func GetServiceRegion(service *corev1.Service, pods corev1sets.PodSet, nodes corev1sets.NodeSet) (string, error) {
	if len(service.Spec.Selector) == 0 {
		return "", eris.Errorf("service %s has no selector", sets.Key(service))
	}
	// get all the pods matching the service selector
	matchingPods := pods.List(func(pod *corev1.Pod) bool {
		return pod.ClusterName != service.GetClusterName() ||
			pod.Namespace != service.GetNamespace() ||
			!labels.SelectorFromSet(service.Spec.Selector).Matches(labels.Set(pod.Labels))
	})
	if len(matchingPods) == 0 {
		return "", eris.Errorf("failed to find matching pod for service with selector %v", service.Spec.Selector)
	}
	// pick any pod; all of the pods' nodes should be in the same region
	pod := matchingPods[0]
	// get the node that the pod is running on
	node, err := nodes.Find(&skv1.ClusterObjectRef{
		ClusterName: pod.ClusterName,
		Name:        pod.Spec.NodeName,
	})
	if err != nil {
		return "", eris.Wrapf(err, "failed to find node for pod %s", sets.Key(pod))
	}

	return getRegionFromNode(node)
}

// Get the zone and sub-zone (if it exists) of a node
func GetSubLocality(
	clusterName string,
	nodeName string,
	nodes corev1sets.NodeSet,
) (*v1alpha2.SubLocality, error) {
	node, err := nodes.Find(&skv1.ClusterObjectRef{
		ClusterName: clusterName,
		Name:        nodeName,
	})
	if err != nil {
		return nil, eris.Wrapf(err, "failed to find node with name %s on cluster %s", nodeName, clusterName)
	}

	// get the zone labels from the node. check both the stable and deprecated labels
	var zone string
	if zoneStable, ok := node.Labels[corev1.LabelZoneFailureDomainStable]; ok {
		zone = zoneStable
	} else if zoneDeprecated, ok := node.Labels[corev1.LabelZoneFailureDomain]; ok {
		zone = zoneDeprecated
	} else {
		return nil, eris.Errorf("failed to find zone label on node %s", node.GetName())
	}

	subLocality := &v1alpha2.SubLocality{
		Zone: zone,
	}

	// get the sub-zone (Istio-specific)
	if subZone, ok := node.Labels[label.TopologySubzone.Name]; ok {
		subLocality.SubZone = subZone
	}

	return subLocality, nil
}

func getRegionFromNode(node *corev1.Node) (string, error) {
	// get the region labels from the node. check both the stable and deprecated labels
	if regionStable, ok := node.Labels[corev1.LabelZoneRegionStable]; ok {
		return regionStable, nil
	} else if regionDeprecated, ok := node.Labels[corev1.LabelZoneRegion]; ok {
		return regionDeprecated, nil
	}
	return "", eris.Errorf("failed to find region label on node %s", node.GetName())
}
