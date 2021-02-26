package utils

import (
	"context"

	. "github.com/onsi/gomega"

	networkingv1 "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func AssertTrafficPolicyStatuses(
	ctx context.Context,
	trafficPolicy networkingv1.TrafficPolicyClient,
	namespace string,
) {

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

func AssertVirtualMeshStatuses(
	ctx context.Context,
	virtualMesh networkingv1.VirtualMeshClient,
	namespace string,
) {
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
