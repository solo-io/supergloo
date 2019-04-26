package istio

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	skpod "github.com/solo-io/solo-kit/pkg/api/external/kubernetes/pod"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/common/kubernetes"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/supergloo/cli/pkg/helpers/clients"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	kubev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("istio mesh discovery unit tests", func() {
	var (
		istioNamespace     = "istio-system"
		superglooNamespace = "supergloo-system"
	)

	var constructPod = func(container kubev1.Container, namespace string) *kubernetes.Pod {

		pod := &kubev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: namespace,
				Name:      "istio-pilot",
			},
			Spec: kubev1.PodSpec{
				Containers: []kubev1.Container{
					container,
				},
			},
		}
		return skpod.FromKubePod(pod)
	}

	BeforeEach(func() {
		clients.UseMemoryClients()
	})
	Context("discovery data", func() {
		It("can properly construct the discovery data", func() {
			container := kubev1.Container{
				Image: "istio-pilot:1.0.6",
			}
			pod := constructPod(container, istioNamespace)
			mesh, err := constructDiscoveredMesh(context.TODO(), pod, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(mesh.Metadata).To(BeEquivalentTo(core.Metadata{
				Labels:    DiscoverySelector,
				Namespace: superglooNamespace,
				Name:      fmt.Sprintf("istio-%s", istioNamespace),
			}))
		})
		It("overwrites discovery data with install info", func() {
			container := kubev1.Container{
				Image: "istio-pilot:1.0.6",
			}
			pod := constructPod(container, istioNamespace)
			helloWorldCert := &core.ResourceRef{
				Namespace: "hello",
				Name:      "world",
			}
			installMeta := core.Metadata{
				Namespace: superglooNamespace,
				Name:      "my-istio",
			}
			installs := v1.InstallList{
				{
					Metadata:              installMeta,
					InstallationNamespace: istioNamespace,
					InstallType: &v1.Install_Mesh{
						Mesh: &v1.MeshInstall{
							MeshInstallType: &v1.MeshInstall_Istio{
								Istio: &v1.IstioInstall{
									CustomRootCert: helloWorldCert,
									EnableMtls:     true,
								},
							},
						},
					},
				},
			}
			mesh, err := constructDiscoveredMesh(context.TODO(), pod, installs)
			Expect(err).NotTo(HaveOccurred())
			Expect(mesh.MtlsConfig).To(BeEquivalentTo(&v1.MtlsConfig{
				RootCertificate: helloWorldCert,
				MtlsEnabled:     true,
			}))
			Expect(mesh.Metadata).To(BeEquivalentTo(core.Metadata{
				Labels:    DiscoverySelector,
				Namespace: installMeta.Namespace,
				Name:      installMeta.Name,
			}))
		})
	})
})
