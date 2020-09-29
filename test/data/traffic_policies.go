package data

import (
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha2"
	v1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// shifts traffic to a subset in the local cluster
func LocalTrafficShiftPolicy(
	name, namespace string,
	destinationService *v1.ClusterObjectRef,
	subset map[string]string,
	port uint32) *v1alpha2.TrafficPolicy {
	return &v1alpha2.TrafficPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "TrafficPolicy",
			APIVersion: v1alpha2.SchemeGroupVersion.String(),
		},
		Spec: v1alpha2.TrafficPolicySpec{
			SourceSelector: nil,
			DestinationSelector: []*v1alpha2.TrafficTargetSelector{{
				KubeServiceRefs: &v1alpha2.TrafficTargetSelector_KubeServiceRefs{
					Services: []*v1.ClusterObjectRef{destinationService},
				},
			}},
			TrafficShift: &v1alpha2.TrafficPolicySpec_MultiDestination{
				Destinations: []*v1alpha2.TrafficPolicySpec_MultiDestination_WeightedDestination{{
					DestinationType: &v1alpha2.TrafficPolicySpec_MultiDestination_WeightedDestination_KubeService{
						KubeService: &v1alpha2.TrafficPolicySpec_MultiDestination_WeightedDestination_KubeDestination{
							Namespace:   destinationService.GetNamespace(),
							Name:        destinationService.GetName(),
							ClusterName: destinationService.GetClusterName(),
							Subset:      subset,
							Port:        port,
						},
					},
				}},
			},
		},
	}
}

// shifts traffic to a subset in the remote cluster
func RemoteTrafficShiftPolicy(
	name, namespace string,
	destinationService *v1.ClusterObjectRef,
	subsetCluster string,
	subset map[string]string,
	port uint32) *v1alpha2.TrafficPolicy {
	return &v1alpha2.TrafficPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "TrafficPolicy",
			APIVersion: v1alpha2.SchemeGroupVersion.String(),
		},
		Spec: v1alpha2.TrafficPolicySpec{
			SourceSelector: nil,
			DestinationSelector: []*v1alpha2.TrafficTargetSelector{{
				KubeServiceRefs: &v1alpha2.TrafficTargetSelector_KubeServiceRefs{
					Services: []*v1.ClusterObjectRef{destinationService},
				},
			}},
			TrafficShift: &v1alpha2.TrafficPolicySpec_MultiDestination{
				Destinations: []*v1alpha2.TrafficPolicySpec_MultiDestination_WeightedDestination{{
					DestinationType: &v1alpha2.TrafficPolicySpec_MultiDestination_WeightedDestination_KubeService{
						KubeService: &v1alpha2.TrafficPolicySpec_MultiDestination_WeightedDestination_KubeDestination{
							Namespace:   destinationService.GetNamespace(),
							Name:        destinationService.GetName(),
							ClusterName: subsetCluster,
							Subset:      subset,
							Port:        port,
						},
					},
				}},
			},
		},
	}
}
