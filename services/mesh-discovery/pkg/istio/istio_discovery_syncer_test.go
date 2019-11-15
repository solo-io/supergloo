package istio_test

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/mesh-projects/pkg/api/external/istio/authorization/v1alpha1"
	v1 "github.com/solo-io/mesh-projects/pkg/api/v1"
	zeph_core "github.com/solo-io/mesh-projects/pkg/api/v1/core"
	. "github.com/solo-io/mesh-projects/services/mesh-discovery/pkg/istio"
	"github.com/solo-io/mesh-projects/test/inputs"
	"github.com/solo-io/solo-kit/api/external/kubernetes/deployment"
	"github.com/solo-io/solo-kit/api/external/kubernetes/job"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/memory"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/common/kubernetes"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	appsv1 "k8s.io/api/apps/v1"
	v13 "k8s.io/api/batch/v1"
	kubev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var _ = Describe("IstioDiscoverySyncer", func() {
	var (
		istioDiscovery    v1.DiscoverySyncer
		reconciler        *mockMeshReconciler
		ingressReconciler *mockMeshIngressReconciler
		meshPolicyClient  v1alpha1.MeshPolicyClient
		crdGetter         *mockCrdGetter
		jobGetter         *mockJobClient
		writeNs           = "write-objects-here"
	)
	BeforeEach(func() {
		reconciler = &mockMeshReconciler{}
		ingressReconciler = &mockMeshIngressReconciler{}
		meshPolicyClient, _ = v1alpha1.NewMeshPolicyClient(
			&factory.MemoryResourceClientFactory{Cache: memory.NewInMemoryResourceCache()})
		crdGetter = &mockCrdGetter{}
		jobGetter = newCompletedJobGetter()
		istioDiscovery = NewIstioDiscoverySyncer(
			writeNs,
			reconciler,
			meshPolicyClient,
			crdGetter,
			jobGetter,
			ingressReconciler,
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
	Context("pilot present, istio crds registered, job not complete", func() {
		var snap *v1.DiscoverySnapshot
		BeforeEach(func() {
			crdGetter.shouldSucceed = true
			snap = &v1.DiscoverySnapshot{
				Deployments: []*kubernetes.Deployment{istioDeployment("istio-system", "1234")},
			}
			istioDiscovery = NewIstioDiscoverySyncer(
				writeNs,
				reconciler,
				meshPolicyClient,
				crdGetter,
				newIncompleteJobGetter(),
				ingressReconciler,
			)
		})
		It("reconciles nil", func() {
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
				jobGetter,
				ingressReconciler,
			)
		})
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
	Context("pilot present, istio crds registered", func() {
		expectedMesh := func(mtlsEnabled, enableAutoInject, smiEnabled bool, rootCert *core.ResourceRef) *v1.Mesh {
			var mtlsConfig *v1.MtlsConfig
			if mtlsEnabled {
				mtlsConfig = &v1.MtlsConfig{
					MtlsEnabled:     mtlsEnabled,
					RootCertificate: rootCert,
				}
			}
			return &v1.Mesh{
				Metadata: core.Metadata{
					Name:      "istio-istio-system",
					Namespace: writeNs,
					Labels:    map[string]string{"discovered_by": "istio-mesh-discovery"},
				},
				MeshType: &v1.Mesh_Istio{
					Istio: &v1.IstioMesh{
						InstallationNamespace: "istio-system",
						Version:               "1234",
					},
				},
				MtlsConfig: mtlsConfig,
				DiscoveryMetadata: &v1.DiscoveryMetadata{
					EnableAutoInject: enableAutoInject,
					MtlsConfig:       mtlsConfig,
				},
				SmiEnabled: smiEnabled,
				EntryPoint: &zeph_core.ClusterResourceRef{
					Resource: core.ResourceRef{
						Name:      "istio-istio-system",
						Namespace: writeNs,
					},
				},
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
				snap.Pods = inputs.BookInfoPodsIstioInject("default")
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

type mockMeshIngressReconciler struct {
	reconcileCalledWith []v1.MeshIngressList
}

func (r *mockMeshIngressReconciler) Reconcile(namespace string, desiredResources v1.MeshIngressList, transition v1.TransitionMeshIngressFunc, opts clients.ListOpts) error {
	r.reconcileCalledWith = append(r.reconcileCalledWith, desiredResources)
	return nil
}

type mockCrdGetter struct {
	shouldSucceed bool
	specialErr    error
}

func (e *mockCrdGetter) Read(_ string, _ clients.ReadOpts) (*kubernetes.CustomResourceDefinition, error) {
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

type mockJobClient struct {
	jobListToReturn kubernetes.JobList
	errorToReturn   error
}

func (m *mockJobClient) BaseClient() clients.ResourceClient {
	panic("implement me")
}

func (m *mockJobClient) Register() error {
	panic("implement me")
}

func (m *mockJobClient) Read(namespace, name string, opts clients.ReadOpts) (*kubernetes.Job, error) {
	panic("implement me")
}

func (m *mockJobClient) Write(resource *kubernetes.Job, opts clients.WriteOpts) (*kubernetes.Job, error) {
	panic("implement me")
}

func (m *mockJobClient) Delete(namespace, name string, opts clients.DeleteOpts) error {
	panic("implement me")
}

func (m *mockJobClient) List(namespace string, opts clients.ListOpts) (kubernetes.JobList, error) {
	return m.jobListToReturn, m.errorToReturn
}

func (m *mockJobClient) Watch(namespace string, opts clients.WatchOpts) (<-chan kubernetes.JobList, <-chan error, error) {
	panic("implement me")
}

func newCompletedJobGetter() *mockJobClient {
	var jobList kubernetes.JobList
	for _, jobName := range PostInstallJobs {
		job := &kubernetes.Job{
			Job: job.Job{
				ObjectMeta: metav1.ObjectMeta{Name: jobName},
				Status: v13.JobStatus{
					Conditions: []v13.JobCondition{{
						Type:   v13.JobComplete,
						Status: kubev1.ConditionTrue,
					}},
				},
			},
		}
		jobList = append(jobList, job)
	}

	return &mockJobClient{
		jobListToReturn: jobList,
	}
}

func newIncompleteJobGetter() *mockJobClient {
	var jobList kubernetes.JobList
	for _, jobName := range PostInstallJobs {
		j := &kubernetes.Job{
			Job: job.Job{
				ObjectMeta: metav1.ObjectMeta{Name: jobName},
				Status:     v13.JobStatus{},
			},
		}
		jobList = append(jobList, j)
	}

	return &mockJobClient{
		jobListToReturn: jobList,
	}
}
