package data

import (
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	v1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TrafficShiftPolicy(name, namespace string, destinationService *v1.ClusterObjectRef, subset map[string]string, port uint32) *v1alpha1.TrafficPolicy {
	return &v1alpha1.TrafficPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "TrafficPolicy",
			APIVersion: v1alpha1.SchemeGroupVersion.String(),
		},
		Spec: v1alpha1.TrafficPolicySpec{
			SourceSelector: nil,
			DestinationSelector: []*v1alpha1.ServiceSelector{{
				KubeServiceRefs: &v1alpha1.ServiceSelector_KubeServiceRefs{
					Services: []*v1.ClusterObjectRef{destinationService},
				},
			}},
			TrafficShift: &v1alpha1.TrafficPolicySpec_MultiDestination{
				Destinations: []*v1alpha1.TrafficPolicySpec_MultiDestination_WeightedDestination{{
					DestinationType: &v1alpha1.TrafficPolicySpec_MultiDestination_WeightedDestination_KubeService{
						KubeService: &v1alpha1.TrafficPolicySpec_MultiDestination_WeightedDestination_KubeDestination{
							Namespace: destinationService.Namespace,
							Name:      destinationService.Name,
							Cluster:   destinationService.ClusterName,
							Subset:    subset,
							Port:      port,
						},
					},
				}},
			},
		},
	}
}
