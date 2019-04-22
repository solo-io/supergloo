package linkerd

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/supergloo/cli/pkg/helpers/clients"
	"github.com/solo-io/supergloo/pkg/api/custom/clients/kubernetes"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/meshdiscovery/clientset"
	kubev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("istio discovery config", func() {

	var (
		cs  *clientset.Clientset
		ctx context.Context
	)

	BeforeEach(func() {
		var err error
		ctx = context.TODO()
		cs, err = clientset.ClientsetFromContext(ctx)
		Expect(err).NotTo(HaveOccurred())
		clients.UseMemoryClients()
	})

	Context("plugin creation", func() {
		It("can be initialized without an error", func() {
			_, err := NewLinkerdConfigDiscoveryRunner(ctx, cs)
			Expect(err).NotTo(HaveOccurred())
		})
	})
	Context("full mesh", func() {

		var (
			mesh    *v1.Mesh
			install *v1.Install
		)
		BeforeEach(func() {
			mesh = &v1.Mesh{
				MeshType: &v1.Mesh_LinkerdMesh{
					LinkerdMesh: &v1.LinkerdMesh{
						InstallationNamespace: "hello",
					},
				},
				MtlsConfig: &v1.MtlsConfig{},
				DiscoveryMetadata: &v1.DiscoveryMetadata{
					InstallationNamespace: "hello",
				},
			}
			install = &v1.Install{
				InstallationNamespace: "world",
				InstallType: &v1.Install_Mesh{
					Mesh: &v1.MeshInstall{
						MeshInstallType: &v1.MeshInstall_LinkerdMesh{
							LinkerdMesh: &v1.LinkerdInstall{
								LinkerdVersion:   "2.2.1",
								EnableMtls:       true,
								EnableAutoInject: true,
							},
						},
					},
				},
			}
		})

		It("Can merge properly with no install or mesh policy", func() {
			fm := &meshResources{
				Mesh: mesh,
			}
			Expect(fm.merge()).To(BeEquivalentTo(fm.Mesh))
		})
		It("can merge properly with install", func() {
			fm := &meshResources{
				Mesh:    mesh,
				Install: install,
			}
			merge := fm.merge()
			Expect(merge.DiscoveryMetadata.MtlsConfig).To(BeEquivalentTo(&v1.MtlsConfig{
				MtlsEnabled: true,
			}))
			Expect(merge.DiscoveryMetadata).To(BeEquivalentTo(&v1.DiscoveryMetadata{
				MtlsConfig: &v1.MtlsConfig{
					MtlsEnabled: true,
				},
				InstallationNamespace:  "world",
				MeshVersion:            "2.2.1",
				InjectedNamespaceLabel: injectionAnnotation,
				EnableAutoInject:       true,
			}))
		})

	})

	Context("filtering annotated pods", func() {
		var (
			injectionEnabled = map[string]string{
				injectionAnnotation: enabled,
			}
			injectionDisabled = map[string]string{
				injectionAnnotation: disabled,
			}
		)

		It("Only use annotated pods", func() {
			pod := &kubev1.Pod{
				Status: kubev1.PodStatus{
					Phase: kubev1.PodRunning,
				},
				ObjectMeta: metav1.ObjectMeta{
					Annotations: injectionEnabled,
				},
				Spec: kubev1.PodSpec{
					Containers: []kubev1.Container{
						{
							Name: proxyContainer,
						},
					},
				},
			}
			namespace := &kubev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "one",
				},
			}
			useNamespace := kubernetes.FromKubeNamespace(namespace)
			usePod := kubernetes.FromKubePod(pod)
			Expect(injectedPodsWithNamespaceAnnotation(usePod, useNamespace)).To(BeTrue())
		})
		It("uses all pods in injected namespaces", func() {
			pod := &kubev1.Pod{
				Status: kubev1.PodStatus{
					Phase: kubev1.PodRunning,
				},
				Spec: kubev1.PodSpec{
					Containers: []kubev1.Container{
						{
							Name: proxyContainer,
						},
					},
				},
			}
			namespace := &kubev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "one",
					Annotations: injectionEnabled,
				},
			}
			useNamespace := kubernetes.FromKubeNamespace(namespace)
			usePod := kubernetes.FromKubePod(pod)
			Expect(injectedPodsWithNamespaceAnnotation(usePod, useNamespace)).To(BeTrue())

		})
		It("skips disabled pods in injected namespaces", func() {
			pod := &kubev1.Pod{
				Status: kubev1.PodStatus{
					Phase: kubev1.PodRunning,
				},
				ObjectMeta: metav1.ObjectMeta{
					Annotations: injectionDisabled,
				},
				Spec: kubev1.PodSpec{
					Containers: []kubev1.Container{
						{
							Name: proxyContainer,
						},
					},
				},
			}
			namespace := &kubev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "one",
					Annotations: injectionEnabled,
				},
			}
			useNamespace := kubernetes.FromKubeNamespace(namespace)
			usePod := kubernetes.FromKubePod(pod)
			Expect(injectedPodsWithNamespaceAnnotation(usePod, useNamespace)).To(BeFalse())
		})
	})
})
