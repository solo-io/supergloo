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

// Get the region of a cluster
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

// UniqueSubLocalities is a list of UniqueSubLocalities
// Exists as a sub component so it can be used in enterprise
// Not thread safe
type UniqueSubLocalities struct {
	subLocalities map[string]map[string]*v1alpha2.SubLocality
}

func (u *UniqueSubLocalities) Add(subLocality *v1alpha2.SubLocality) {
	if u.subLocalities == nil {
		u.subLocalities = map[string]map[string]*v1alpha2.SubLocality{}
	}

	// We do not care to add sub localities with empty zones to the list.
	// Empty sub-zones are okay however.
	if subLocality.GetZone() == "" {
		return
	}

	if zone, foundZone := u.subLocalities[subLocality.GetZone()]; foundZone {
		if _, foundSubZone := u.subLocalities[subLocality.GetSubZone()]; foundSubZone {
			// entry already exists
			return
		}
		// Add the new sublocality in the existing zone
		zone[subLocality.GetSubZone()] = subLocality
	} else {
		u.subLocalities[subLocality.GetZone()] = map[string]*v1alpha2.SubLocality{
			subLocality.GetSubZone(): subLocality,
		}
	}

}

func (u *UniqueSubLocalities) List() []*v1alpha2.SubLocality {

	var result []*v1alpha2.SubLocality

	for _, byZone := range u.subLocalities {
		for _, bySubZone := range byZone {
			result = append(result, bySubZone)
		}
	}

	return result
}

func GetUniqueClusterSubLocalities(
	clusterName string,
	nodes corev1sets.NodeSet,
) ([]*v1alpha2.SubLocality, error) {

	// Map to ensure uniqueness of sub_localities being added to
	localities := &UniqueSubLocalities{}
	for _, node := range nodes.List(func(node *corev1.Node) bool {
		return node.GetClusterName() == clusterName
	}) {

		subLocality, err := getSubLocality(node)
		if err != nil {
			return nil, err
		}

		localities.Add(subLocality)
	}

	return localities.List(), nil
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

	return getSubLocality(node)
}

func getSubLocality(
	node *corev1.Node,
) (*v1alpha2.SubLocality, error) {

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
	if subzone, ok := node.Labels[label.TopologySubzone.Name]; ok {
		subLocality.SubZone = subzone
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
