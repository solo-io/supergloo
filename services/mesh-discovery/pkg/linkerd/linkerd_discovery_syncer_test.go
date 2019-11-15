package linkerd_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "github.com/solo-io/mesh-projects/pkg/api/v1"
	zeph_core "github.com/solo-io/mesh-projects/pkg/api/v1/core"
	. "github.com/solo-io/mesh-projects/services/mesh-discovery/pkg/linkerd"
	"github.com/solo-io/mesh-projects/test/inputs"
	"github.com/solo-io/solo-kit/api/external/kubernetes/deployment"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/common/kubernetes"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	appsv1 "k8s.io/api/apps/v1"
	kubev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("LinkerdDiscoverySyncer", func() {
	var (
		linkerdDiscovery  v1.DiscoverySyncer
		reconciler        *mockMeshReconciler
		ingressReconciler *mockMeshIngressReconciler
		writeNs           = "write-objects-here"
	)
	BeforeEach(func() {
		reconciler = &mockMeshReconciler{}
		ingressReconciler = &mockMeshIngressReconciler{}
		linkerdDiscovery = NewLinkerdDiscoverySyncer(
			writeNs,
			reconciler,
			ingressReconciler,
		)
	})
	Context("linkerd controller not present", func() {
		It("reconciles nil", func() {
			snap := &v1.DiscoverySnapshot{}
			err := linkerdDiscovery.Sync(context.TODO(), snap)
			Expect(err).NotTo(HaveOccurred())
			Expect(reconciler.reconcileCalledWith).To(HaveLen(1))
			Expect(reconciler.reconcileCalledWith[0]).To(HaveLen(0))
		})
	})
	Context("linkerd controller present", func() {
		expectedMesh := func(enableAutoInject bool) *v1.Mesh {
			return &v1.Mesh{
				Metadata: core.Metadata{
					Name:      "linkerd-linkerd",
					Namespace: writeNs,
					Labels:    map[string]string{"discovered_by": "linkerd-mesh-discovery"},
				},
				MeshType: &v1.Mesh_Linkerd{
					Linkerd: &v1.LinkerdMesh{
						InstallationNamespace: "linkerd",
						Version:               "1234",
					},
				},
				MtlsConfig: &v1.MtlsConfig{
					MtlsEnabled: true,
				},
				DiscoveryMetadata: &v1.DiscoveryMetadata{
					EnableAutoInject: enableAutoInject,
					MtlsConfig: &v1.MtlsConfig{
						MtlsEnabled: true,
					},
				},
				EntryPoint: &zeph_core.ClusterResourceRef{
					Resource: core.ResourceRef{
						Name:      "linkerd-linkerd",
						Namespace: writeNs,
					},
				},
			}
		}
		var snap *v1.DiscoverySnapshot
		BeforeEach(func() {
			snap = &v1.DiscoverySnapshot{
				Deployments: []*kubernetes.Deployment{linkerdDeployment("linkerd", "1234")},
			}
		})
		Context("sidecar injector deployed", func() {
			BeforeEach(func() {
				snap.Deployments = append(snap.Deployments, kubernetes.NewDeployment("linkerd", "linkerd-proxy-injector"))
			})
			It("sets enable auto inject true", func() {
				err := linkerdDiscovery.Sync(context.TODO(), snap)
				Expect(err).NotTo(HaveOccurred())
				Expect(reconciler.reconcileCalledWith).To(HaveLen(1))
				Expect(reconciler.reconcileCalledWith[0]).To(HaveLen(1))
				Expect(reconciler.reconcileCalledWith[0][0]).To(Equal(
					expectedMesh(true)))
			})
		})
		Context("with injected pods", func() {
			BeforeEach(func() {
				// if you look at the bookinfopods list, not all the pods
				// are "finished" with their init container, so they don't all get
				// recognized as injected (which is fine)
				snap.Pods = inputs.BookInfoPodsLinkerdInject("default")
				snap.Upstreams = inputs.BookInfoUpstreams("default")
			})
			It("adds upstreams for the injected pods", func() {
				expected := expectedMesh(false)
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
				err := linkerdDiscovery.Sync(context.TODO(), snap)
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

type mockMeshIngressReconciler struct {
	reconcileCalledWith []v1.MeshIngressList
}

func (r *mockMeshIngressReconciler) Reconcile(namespace string, desiredResources v1.MeshIngressList, transition v1.TransitionMeshIngressFunc, opts clients.ListOpts) error {
	r.reconcileCalledWith = append(r.reconcileCalledWith, desiredResources)
	return nil
}

func linkerdDeployment(namespace, version string) *kubernetes.Deployment {
	return &kubernetes.Deployment{
		Deployment: deployment.Deployment{
			ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: "name doesn't matter in this context"},
			Spec: appsv1.DeploymentSpec{
				Template: kubev1.PodTemplateSpec{
					Spec: kubev1.PodSpec{
						Containers: []kubev1.Container{{
							Image: "gcr.io/linkerd-io/controller:" + version,
						}},
					},
				},
			},
		},
	}
}
