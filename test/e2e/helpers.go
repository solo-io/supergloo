package e2e

import (
	"context"
	"os/exec"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1alpha1types "github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1/types"
	v1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
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

func MeshCtl(args string) error {
	return MeshCtlCommand(strings.Fields(args)...).Run()
}

func MeshCtlCommand(arg ...string) *exec.Cmd {
	arg = append([]string{"--context", GetEnv().Management.Context}, arg...)
	return exec.Command("../../_output/meshctl", arg...)
}
