package istio_test

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/api/external/kubernetes/deployment"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/memory"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/common/kubernetes"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/supergloo/pkg/api/external/istio/authorization/v1alpha1"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	. "github.com/solo-io/supergloo/pkg/meshdiscovery/istio"
	"github.com/solo-io/supergloo/test/inputs"
	appsv1 "k8s.io/api/apps/v1"
	kubev1 "k8s.io/api/core/v1"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var _ = Describe("IstioDiscoverySyncer", func() {
	var (
		istioDiscovery   v1.DiscoverySyncer
		reconciler       *mockMeshReconciler
		meshPolicyClient v1alpha1.MeshPolicyClient
		crdGetter        *mockCrdGetter
		writeNs          = "write-objects-here"
	)
	BeforeEach(func() {
		reconciler = &mockMeshReconciler{}
		meshPolicyClient, _ = v1alpha1.NewMeshPolicyClient(
			&factory.MemoryResourceClientFactory{Cache: memory.NewInMemoryResourceCache()})
		crdGetter = &mockCrdGetter{}
		istioDiscovery = NewIstioDiscoverySyncer(
			writeNs,
			reconciler,
			meshPolicyClient,
			crdGetter,
		)
	})
	Context("pilot not present", func() {
		It("reconciles nil", func() {
			snap := &v1.DiscoverySnapshot{}
			err := istioDiscovery.Sync(context.TODO(), snap)
			Expect(err).NotTo(HaveOccurred())
			Expect(reconciler.reconcileCalledWith).To(HaveLen(1))
			Expect(reconciler.reconcileCalledWith[0]).To(HaveLen(0))
		})
	})
	Context("pilot present, istio crds not registered", func() {
		It("reconciles nil", func() {
			snap := &v1.DiscoverySnapshot{
				Deployments: []*kubernetes.Deployment{istioDeployment("istio-system", "1234")},
			}
			err := istioDiscovery.Sync(context.TODO(), snap)
			Expect(err).NotTo(HaveOccurred())
			Expect(reconciler.reconcileCalledWith).To(HaveLen(1))
			Expect(reconciler.reconcileCalledWith[0]).To(HaveLen(0))
		})
	})
	Context("pilot present, meshpolicy client failing", func() {
		BeforeEach(func() {
			// use a meshpolicy client we know will always fail
			crdGetter = &mockCrdGetter{specialErr: fmt.Errorf("they made me do it")}
			istioDiscovery = NewIstioDiscoverySyncer(
				writeNs,
				reconciler,
				meshPolicyClient,
				crdGetter,
			)
		})
		It("returns a DetectingMeshPolicy error", func() {
			snap := &v1.DiscoverySnapshot{
				Deployments: []*kubernetes.Deployment{istioDeployment("istio-system", "1234")},
			}
			err := istioDiscovery.Sync(context.TODO(), snap)
			Expect(err).To(HaveOccurred())
			Expect(IsErrorDetectingMeshPolicy(err)).To(BeTrue())
		})
	})
	Context("pilot present, istio crds registered", func() {
		expectedMesh := func(mtlsEnabled, enableAutoInject, smiEnabled bool, rootCert *core.ResourceRef) *v1.Mesh {
			return &v1.Mesh{
				Metadata: core.Metadata{
					Name:      "istio-system-istio",
					Namespace: writeNs,
					Labels:    map[string]string{"discovered_by": "istio-mesh-discovery"},
				},
				MeshType: &v1.Mesh_Istio{
					Istio: &v1.IstioMesh{
						InstallationNamespace: "istio-system",
						Version:               "1234",
					},
				},
				MtlsConfig: &v1.MtlsConfig{
					MtlsEnabled:     mtlsEnabled,
					RootCertificate: rootCert,
				},
				DiscoveryMetadata: &v1.DiscoveryMetadata{
					EnableAutoInject: enableAutoInject,
					MtlsConfig: &v1.MtlsConfig{
						MtlsEnabled:     mtlsEnabled,
						RootCertificate: rootCert,
					},
				},
				SmiEnabled: smiEnabled,
			}
		}
		var snap *v1.DiscoverySnapshot
		BeforeEach(func() {
			crdGetter.shouldSucceed = true
			snap = &v1.DiscoverySnapshot{
				Deployments: []*kubernetes.Deployment{istioDeployment("istio-system", "1234")},
			}
		})
		Context("no meshpolicy, no adapter, no smi, no root cert, no injected pods", func() {
			It("determines the correct namespace and version of the mesh", func() {
				err := istioDiscovery.Sync(context.TODO(), snap)
				Expect(err).NotTo(HaveOccurred())
				Expect(reconciler.reconcileCalledWith).To(HaveLen(1))
				Expect(reconciler.reconcileCalledWith[0]).To(HaveLen(1))
				Expect(reconciler.reconcileCalledWith[0][0]).To(Equal(
					expectedMesh(false, false, false, nil)))
			})
		})
		Context("meshpolicy with mtls enabled", func() {
			BeforeEach(func() {
				_, _ = meshPolicyClient.Write(&v1alpha1.MeshPolicy{
					Metadata: core.Metadata{
						Name: "default",
					},
					Peers: []*v1alpha1.PeerAuthenticationMethod{{
						Params: &v1alpha1.PeerAuthenticationMethod_Mtls{
							Mtls: &v1alpha1.MutualTls{},
						},
					}},
				}, clients.WriteOpts{})
			})
			Context("cacerts missing", func() {
				It("sets mtls enabled true", func() {
					err := istioDiscovery.Sync(context.TODO(), snap)
					Expect(err).NotTo(HaveOccurred())
					Expect(reconciler.reconcileCalledWith).To(HaveLen(1))
					Expect(reconciler.reconcileCalledWith[0]).To(HaveLen(1))
					Expect(reconciler.reconcileCalledWith[0][0]).To(Equal(
						expectedMesh(true, false, false, nil)))
				})
			})
			Context("cacerts present", func() {
				BeforeEach(func() {
					snap.Tlssecrets = v1.TlsSecretList{{Metadata: core.Metadata{Name: "cacerts", Namespace: "istio-system"}}}
				})
				It("sets mtls enabled true", func() {
					err := istioDiscovery.Sync(context.TODO(), snap)
					Expect(err).NotTo(HaveOccurred())
					Expect(reconciler.reconcileCalledWith).To(HaveLen(1))
					Expect(reconciler.reconcileCalledWith[0]).To(HaveLen(1))
					Expect(reconciler.reconcileCalledWith[0][0]).To(Equal(
						expectedMesh(true, false, false, &core.ResourceRef{Name: "cacerts", Namespace: "istio-system"})))
				})
			})

		})
		Context("sidecar injector deployed", func() {
			BeforeEach(func() {
				snap.Deployments = append(snap.Deployments, kubernetes.NewDeployment("istio-system", "istio-sidecar-injector"))
			})
			It("sets enable auto inject true", func() {
				err := istioDiscovery.Sync(context.TODO(), snap)
				Expect(err).NotTo(HaveOccurred())
				Expect(reconciler.reconcileCalledWith).To(HaveLen(1))
				Expect(reconciler.reconcileCalledWith[0]).To(HaveLen(1))
				Expect(reconciler.reconcileCalledWith[0][0]).To(Equal(
					expectedMesh(false, true, false, nil)))
			})
		})
		Context("smi adapter deployed", func() {
			BeforeEach(func() {
				snap.Deployments = append(snap.Deployments, kubernetes.NewDeployment("istio-system", "smi-adapter-istio"))
			})
			It("sets smienabled true", func() {
				err := istioDiscovery.Sync(context.TODO(), snap)
				Expect(err).NotTo(HaveOccurred())
				Expect(reconciler.reconcileCalledWith).To(HaveLen(1))
				Expect(reconciler.reconcileCalledWith[0]).To(HaveLen(1))
				Expect(reconciler.reconcileCalledWith[0][0]).To(Equal(
					expectedMesh(false, false, true, nil)))
			})
		})
		Context("with injected pods", func() {
			BeforeEach(func() {
				// if you look at the bookinfopods list, not all the pods
				// are "finished" with their init container, so they don't all get
				// recognized as injected (which is fine)
				snap.Pods = inputs.BookInfoPods("default")
				snap.Upstreams = inputs.BookInfoUpstreams("default")
			})
			It("adds upstreams for the injected pods", func() {
				expected := expectedMesh(false, false, false, nil)
				expected.DiscoveryMetadata.Upstreams = []*core.ResourceRef{
					{
						Name:      "default-details-9080",
						Namespace: "default",
					},
					{
						Name:      "default-details-v1-9080",
						Namespace: "default",
					},
					{
						Name:      "default-productpage-9080",
						Namespace: "default",
					},
					{
						Name:      "default-productpage-v1-9080",
						Namespace: "default",
					},
					{
						Name:      "default-ratings-9080",
						Namespace: "default",
					},
					{
						Name:      "default-ratings-v1-9080",
						Namespace: "default",
					},
					{
						Name:      "default-reviews-9080",
						Namespace: "default",
					},
					{
						Name:      "default-reviews-v1-9080",
						Namespace: "default",
					},
					{
						Name:      "default-reviews-v2-9080",
						Namespace: "default",
					},
					{
						Name:      "default-reviews-v3-9080",
						Namespace: "default",
					},
				}
				err := istioDiscovery.Sync(context.TODO(), snap)
				Expect(err).NotTo(HaveOccurred())
				Expect(reconciler.reconcileCalledWith).To(HaveLen(1))
				Expect(reconciler.reconcileCalledWith[0]).To(HaveLen(1))
				Expect(reconciler.reconcileCalledWith[0][0]).To(Equal(expected))
			})
		})
	})
})

type mockMeshReconciler struct {
	reconcileCalledWith []v1.MeshList
}

func (r *mockMeshReconciler) Reconcile(namespace string, desiredResources v1.MeshList, transition v1.TransitionMeshFunc, opts clients.ListOpts) error {
	r.reconcileCalledWith = append(r.reconcileCalledWith, desiredResources)
	return nil
}

type mockCrdGetter struct {
	shouldSucceed bool
	specialErr    error
}

func (e *mockCrdGetter) Get(_ string, _ metav1.GetOptions) (*v1beta1.CustomResourceDefinition, error) {
	if e.specialErr != nil {
		return nil, e.specialErr
	}
	if !e.shouldSucceed {
		return nil, errors.NewNotFound(schema.GroupResource{}, "")
	}
	return nil, nil
}

func istioDeployment(namespace, version string) *kubernetes.Deployment {
	return &kubernetes.Deployment{
		Deployment: deployment.Deployment{
			ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: "name doesn't matter in this context"},
			Spec: appsv1.DeploymentSpec{
				Template: kubev1.PodTemplateSpec{
					Spec: kubev1.PodSpec{
						Containers: []kubev1.Container{{
							Image: "docker.io/istio/pilot:" + version,
						}},
					},
				},
			},
		},
	}
}
