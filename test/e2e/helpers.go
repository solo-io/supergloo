package e2e

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1alpha1types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	v1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func StatusOf(something interface{}, kc KubeContext) func() v1alpha1types.Status_State {
	switch obj := something.(type) {
	case v1alpha1.TrafficPolicy:
		return func() v1alpha1types.Status_State { return StatusOfTrafficPolicy(obj, kc) }
	case *v1alpha1.TrafficPolicy:
		return func() v1alpha1types.Status_State { return StatusOfTrafficPolicy(*obj, kc) }
	default:
		Fail("unknown object")
	}
	panic("never happens")
}

func StatusOfTrafficPolicy(tp v1alpha1.TrafficPolicy, kc KubeContext) v1alpha1types.Status_State {
	key := client.ObjectKey{
		Name:      tp.Name,
		Namespace: tp.Namespace,
	}
	newtp, err := kc.TrafficPolicyClient.GetTrafficPolicy(context.Background(), key)
	Expect(err).NotTo(HaveOccurred())
	return newtp.Status.GetTranslationStatus().GetState()
}
