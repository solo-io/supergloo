package utils

import (
	"context"

	. "github.com/onsi/gomega"

	networkingv1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func AssertTrafficPolicyStatuses(dynamicClient client.Client, namespace string) {
	ctx := context.Background()
	trafficPolicy := networkingv1alpha2.NewTrafficPolicyClient(dynamicClient)

	EventuallyWithOffset(1, func() bool {
		list, err := trafficPolicy.ListTrafficPolicy(ctx, client.InNamespace(namespace))
		ExpectWithOffset(1, err).NotTo(HaveOccurred())
		for _, policy := range list.Items {
			if policy.Status.ObservedGeneration == 0 {
				return false
			}
		}
		return true
	}, "20s").Should(BeTrue())
}

func AssertVirtualMeshStatuses(dynamicClient client.Client, namespace string) {
	ctx := context.Background()
	virtualMesh := networkingv1alpha2.NewVirtualMeshClient(dynamicClient)

	EventuallyWithOffset(1, func() bool {
		list, err := virtualMesh.ListVirtualMesh(ctx, client.InNamespace(namespace))
		ExpectWithOffset(1, err).NotTo(HaveOccurred())
		ExpectWithOffset(1, list.Items).To(HaveLen(1))
		for _, policy := range list.Items {
			if policy.Status.ObservedGeneration == 0 {
				return false
			}
		}
		return true
	}, "2m").Should(BeTrue())
}
