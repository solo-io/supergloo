package e2e

import (
	"context"
	"os/exec"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	zephyr_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func StatusOf(something interface{}, kc KubeContext) func() zephyr_core_types.Status_State {
	switch obj := something.(type) {
	case zephyr_networking.TrafficPolicy:
		return func() zephyr_core_types.Status_State { return StatusOfTrafficPolicy(obj, kc) }
	case *zephyr_networking.TrafficPolicy:
		return func() zephyr_core_types.Status_State { return StatusOfTrafficPolicy(*obj, kc) }
	default:
		Fail("unknown object")
	}
	panic("never happens")
}

func StatusOfTrafficPolicy(tp zephyr_networking.TrafficPolicy, kc KubeContext) zephyr_core_types.Status_State {
	key := client.ObjectKey{
		Name:      tp.Name,
		Namespace: tp.Namespace,
	}
	newtp, err := kc.TrafficPolicyClient.GetTrafficPolicy(context.Background(), key)
	Expect(err).NotTo(HaveOccurred())
	return newtp.Status.GetTranslationStatus().GetState()
}

func KubeClusterShouldExist(key client.ObjectKey, kc KubeContext) func() *zephyr_discovery.KubernetesCluster {
	return func() *zephyr_discovery.KubernetesCluster {
		kubeCluster, _ := kc.KubeClusterClient.GetKubernetesCluster(context.Background(), key)
		return kubeCluster
	}
}

func MeshCtl(args string) error {
	return MeshCtlCommand(strings.Fields(args)...).Run()
}

func MeshCtlCommand(arg ...string) *exec.Cmd {
	arg = append([]string{"--context", GetEnv().Management.Context}, arg...)
	return exec.Command("../../_output/meshctl", arg...)
}
