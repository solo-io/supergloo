package e2e

import (
	"context"
	"os/exec"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	smh_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1/types"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	smh_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func StatusOf(something interface{}, kc KubeContext) func() smh_core_types.Status_State {
	switch obj := something.(type) {
	case smh_networking.TrafficPolicy:
		return func() smh_core_types.Status_State { return StatusOfTrafficPolicy(obj, kc) }
	case *smh_networking.TrafficPolicy:
		return func() smh_core_types.Status_State { return StatusOfTrafficPolicy(*obj, kc) }
	case smh_networking.FailoverService:
		return func() smh_core_types.Status_State { return StatusOfFailoverService(obj, kc) }
	case *smh_networking.FailoverService:
		return func() smh_core_types.Status_State { return StatusOfFailoverService(*obj, kc) }
	default:
		Fail("unknown object")
	}
	panic("never happens")
}

func StatusOfTrafficPolicy(tp smh_networking.TrafficPolicy, kc KubeContext) smh_core_types.Status_State {
	key := client.ObjectKey{
		Name:      tp.Name,
		Namespace: tp.Namespace,
	}
	newtp, err := kc.TrafficPolicyClient.GetTrafficPolicy(context.Background(), key)
	Expect(err).NotTo(HaveOccurred())
	return newtp.Status.GetTranslationStatus().GetState()
}

func StatusOfFailoverService(failoverService smh_networking.FailoverService, kc KubeContext) smh_core_types.Status_State {
	key := client.ObjectKey{
		Name:      failoverService.Name,
		Namespace: failoverService.Namespace,
	}
	fetchedFailoverService, err := kc.FailoverServiceClient.GetFailoverService(context.Background(), key)
	Expect(err).NotTo(HaveOccurred())
	return fetchedFailoverService.Status.GetTranslationStatus().GetState()
}

func KubeCluster(key client.ObjectKey, kc KubeContext) func() *smh_discovery.KubernetesCluster {
	return func() *smh_discovery.KubernetesCluster {
		kubeCluster, _ := kc.KubeClusterClient.GetKubernetesCluster(context.Background(), key)
		return kubeCluster
	}
}

func Mesh(key client.ObjectKey, kc KubeContext) func() *smh_discovery.Mesh {
	return func() *smh_discovery.Mesh {
		mesh, _ := kc.MeshClient.GetMesh(context.Background(), key)
		return mesh
	}
}

func MeshCtl(args string) error {
	return MeshCtlCommand(strings.Fields(args)...).Run()
}

func MeshCtlCommand(arg ...string) *exec.Cmd {
	arg = append([]string{"--context", GetEnv().Management.Context}, arg...)
	return exec.Command("../../_output/meshctl", arg...)
}

var ApplyTrafficPolicy = func(env Env, tpYaml string) {
	var tp smh_networking.TrafficPolicy
	ParseYaml(tpYaml, &tp)
	err := env.Management.TrafficPolicyClient.CreateTrafficPolicy(context.Background(), &tp)
	Expect(err).NotTo(HaveOccurred())
	// see that it was accepted
	Eventually(StatusOf(tp, env.Management), "1m", "1s").Should(Equal(smh_core_types.Status_ACCEPTED))
}

var CurlReviews = func(env Env) func() string {
	return func() string {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute/3)
		defer cancel()
		out := env.Management.GetPod("default", "productpage").Curl(ctx, "http://reviews:9080/reviews/1", "-v")
		GinkgoWriter.Write([]byte(out))
		return out
	}
}
