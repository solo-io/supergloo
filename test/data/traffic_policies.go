package data

import (
	commonv1 "github.com/solo-io/gloo-mesh/pkg/api/common.mesh.gloo.solo.io/v1"
	v1 "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1"
	skv2corev1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// shifts traffic to a subset in the local cluster
func LocalTrafficShiftPolicy(
	name, namespace string,
	destinationService *skv2corev1.ClusterObjectRef,
	subset map[string]string,
	port uint32) *v1.TrafficPolicy {
	return &v1.TrafficPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "TrafficPolicy",
			APIVersion: v1.SchemeGroupVersion.String(),
		},
		Spec: v1.TrafficPolicySpec{
			SourceSelector: nil,
			DestinationSelector: []*commonv1.DestinationSelector{
				{
					KubeServiceRefs: &commonv1.DestinationSelector_KubeServiceRefs{
						Services: []*skv2corev1.ClusterObjectRef{destinationService},
					},
				},
			},
			Policy: &v1.TrafficPolicySpec_Policy{
				TrafficShift: &v1.TrafficPolicySpec_Policy_MultiDestination{
					Destinations: []*v1.TrafficPolicySpec_Policy_MultiDestination_WeightedDestination{{
						DestinationType: &v1.TrafficPolicySpec_Policy_MultiDestination_WeightedDestination_KubeService{
							KubeService: &v1.TrafficPolicySpec_Policy_MultiDestination_WeightedDestination_KubeDestination{
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
		},
	}
}

// shifts traffic to a subset in the remote cluster
func RemoteTrafficShiftPolicy(
	name, namespace string,
	destinationService *skv2corev1.ClusterObjectRef,
	subsetCluster string,
	subset map[string]string,
	port uint32) *v1.TrafficPolicy {
	return &v1.TrafficPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "TrafficPolicy",
			APIVersion: v1.SchemeGroupVersion.String(),
		},
		Spec: v1.TrafficPolicySpec{
			SourceSelector: nil,
			DestinationSelector: []*commonv1.DestinationSelector{
				{
					KubeServiceRefs: &commonv1.DestinationSelector_KubeServiceRefs{
						Services: []*skv2corev1.ClusterObjectRef{destinationService},
					},
				},
			},
			Policy: &v1.TrafficPolicySpec_Policy{
				TrafficShift: &v1.TrafficPolicySpec_Policy_MultiDestination{
					Destinations: []*v1.TrafficPolicySpec_Policy_MultiDestination_WeightedDestination{{
						DestinationType: &v1.TrafficPolicySpec_Policy_MultiDestination_WeightedDestination_KubeService{
							KubeService: &v1.TrafficPolicySpec_Policy_MultiDestination_WeightedDestination_KubeDestination{
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
		},
	}
}
