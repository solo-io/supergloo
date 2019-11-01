package istio_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/mesh-discovery/pkg/istio"
	"github.com/solo-io/solo-kit/api/external/kubernetes/deployment"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/common/kubernetes"
	appsv1 "k8s.io/api/apps/v1"
	kubev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	detector      istio.IstioResourceDetector
	testNamespace = "test-ns"
)

var _ = Describe("Istio Resource Detector", func() {

	istioDeployment := func(namespace, registryRepo, version string) *kubernetes.Deployment {
		return &kubernetes.Deployment{
			Deployment: deployment.Deployment{
				ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: "name doesn't matter in this context"},
				Spec: appsv1.DeploymentSpec{
					Template: kubev1.PodTemplateSpec{
						Spec: kubev1.PodSpec{
							Containers: []kubev1.Container{{
								Image: registryRepo + ":" + version,
							}},
						},
					},
				},
			},
		}
	}

	BeforeEach(func() {
		detector = istio.NewIstioResourceDetector()
	})

	Describe("DetectPilotDeployments", func() {
		It("works", func() {
			testCases := []struct {
				description string
				deployments kubernetes.DeploymentList
				expected    []istio.PilotDeployment
			}{
				{
					description: "with an istio pilot deployment",
					deployments: kubernetes.DeploymentList{istioDeployment(testNamespace, "docker.io/istio/pilot", "1.3.0")},
					expected:    []istio.PilotDeployment{{Namespace: testNamespace, Version: "1.3.0"}},
				},
				{
					description: "when no pilot containers are found",
					deployments: kubernetes.DeploymentList{
						istioDeployment(testNamespace, "docker.io/foo/pilot", "1.3.0"),
						istioDeployment(testNamespace, "registry.redhat.io/foo-service-mesh/pilot-rhel8", "1.0.0"),
						istioDeployment(testNamespace, "registry.redhat.io/foo-service-mesh/foo-rhel8", "2.0.0"),
					},
					expected: nil,
				},
			}

			for _, tc := range testCases {
				Expect(detector.DetectPilotDeployments(context.Background(), tc.deployments)).To(Equal(tc.expected), tc.description)
			}
		})
	})

})
