package appmesh_test

import (
	"context"
	"fmt"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/solo-kit/pkg/api/external/kubernetes/pod"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/common/kubernetes"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	clienthelpers "github.com/solo-io/supergloo/cli/pkg/helpers/clients"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	mocks "github.com/solo-io/supergloo/pkg/config/appmesh/mocks"
	. "github.com/solo-io/supergloo/pkg/meshdiscovery/appmesh"
	"github.com/solo-io/supergloo/test/inputs/appmesh/scenarios"
	kubev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Appmesh", func() {

	var (
		appmeshDiscovery v1.DiscoverySyncer
		reconciler       *mockMeshReconciler
		writeNs          = "write-objects-here"
		ctrl             *gomock.Controller

		ctx           context.Context
		clientBuilder *mocks.MockClientBuilder
		awsClient     *mocks.MockClient
		secretClient  gloov1.SecretClient
		region        = "us-east-1"
		secretName    = "my-aws-secret"
	)
	BeforeEach(func() {
		clienthelpers.UseMemoryClients()
		secretClient = clienthelpers.MustSecretClient()
		ctx = context.TODO()
		ctrl = gomock.NewController(T)
		clientBuilder = mocks.NewMockClientBuilder(ctrl)
		awsClient = mocks.NewMockClient(ctrl)

		reconciler = &mockMeshReconciler{}
		appmeshDiscovery = NewAppmeshDiscoverySyncer(
			writeNs,
			reconciler,
			clientBuilder,
			secretClient,
		)
	})

	var createSecret = func(name string, aws bool) *gloov1.Secret {
		secret := &gloov1.Secret{
			Metadata: core.Metadata{
				Name:      name,
				Namespace: name,
			},
			Kind: &gloov1.Secret_Aws{
				Aws: &gloov1.AwsSecret{},
			},
		}
		if !aws {
			secret.Kind = &gloov1.Secret_Extension{
				Extension: &gloov1.Extension{},
			}
		}
		secret, err := secretClient.Write(secret, clients.WriteOpts{})
		Expect(err).NotTo(HaveOccurred())
		return secret
	}

	var expectedMesh = func(meshName, secretName string, upstreams []*core.ResourceRef) *v1.Mesh {
		return &v1.Mesh{
			Metadata: core.Metadata{
				Name:      "appmesh-" + meshName,
				Namespace: writeNs,
				Labels: map[string]string{
					"discovered_by": "appmesh-mesh-discovery",
				},
			},
			MeshType: &v1.Mesh_AwsAppMesh{
				AwsAppMesh: &v1.AwsAppMesh{
					AwsSecret: &core.ResourceRef{
						Name:      secretName,
						Namespace: secretName,
					},
					Region:           region,
					VirtualNodeLabel: "virtual-node",
				},
			},
			DiscoveryMetadata: &v1.DiscoveryMetadata{
				Upstreams: upstreams,
			},
		}
	}

	Context("eks pods not present", func() {
		It("reconciles nil", func() {
			snap := &v1.DiscoverySnapshot{}
			err := appmeshDiscovery.Sync(ctx, snap)
			Expect(err).NotTo(HaveOccurred())
			Expect(reconciler.reconcileCalledWith).To(HaveLen(1))
			Expect(reconciler.reconcileCalledWith[0]).To(HaveLen(0))
		})
	})
	Context("eks pods present, no aws secrets", func() {
		It("reconciles nil", func() {
			createSecret(secretName, false)
			snap := &v1.DiscoverySnapshot{
				Pods: kubernetes.PodList{eksPod()},
			}
			err := appmeshDiscovery.Sync(ctx, snap)
			Expect(err).NotTo(HaveOccurred())
			Expect(reconciler.reconcileCalledWith).To(HaveLen(1))
			Expect(reconciler.reconcileCalledWith[0]).To(HaveLen(0))
		})
	})
	Context("eks pods present, 1 aws secret, 0 meshes present", func() {
		It("reconciles nil", func() {
			secret := createSecret(secretName, true)
			secretRef := secret.Metadata.Ref()
			snap := &v1.DiscoverySnapshot{
				Pods: kubernetes.PodList{eksPod()},
			}

			wrappedCtx := contextutils.WithLogger(ctx, fmt.Sprintf("appmesh-mesh-discovery-%v", snap.Hash()))

			clientBuilder.EXPECT().GetClientInstance(&secretRef, region).Times(1).Return(awsClient, nil)
			awsClient.EXPECT().ListMeshes(wrappedCtx).Times(1).Return([]string{}, nil)

			err := appmeshDiscovery.Sync(ctx, snap)
			Expect(err).NotTo(HaveOccurred())
			Expect(reconciler.reconcileCalledWith).To(HaveLen(1))
			Expect(reconciler.reconcileCalledWith[0]).To(HaveLen(0))
		})
	})
	Context("eks pods present, 1 aws secret, 1 mesh present", func() {
		It("reconciles a mesh ", func() {
			secret := createSecret(secretName, true)
			secretRef := secret.Metadata.Ref()
			snap := &v1.DiscoverySnapshot{
				Pods: kubernetes.PodList{eksPod()},
			}

			clientBuilder.EXPECT().GetClientInstance(&secretRef, region).Times(1).Return(awsClient, nil)
			meshName := "test-mesh"
			awsClient.EXPECT().ListMeshes(gomock.Any()).Times(1).Return([]string{meshName}, nil)

			err := appmeshDiscovery.Sync(ctx, snap)
			Expect(err).NotTo(HaveOccurred())
			Expect(reconciler.reconcileCalledWith).To(HaveLen(1))
			Expect(reconciler.reconcileCalledWith[0]).To(HaveLen(1))
			Expect(reconciler.reconcileCalledWith[0][0]).To(Equal(expectedMesh(meshName, secret.Metadata.Name, nil)))
		})
	})
	Context("eks pods present, 2 aws secrets that return the same mesh", func() {
		It("reconciles only one mesh", func() {
			secret1 := createSecret("secret1", true)
			secret2 := createSecret("secret2", true)
			secretRef1 := secret1.Metadata.Ref()
			secretRef2 := secret2.Metadata.Ref()
			snap := &v1.DiscoverySnapshot{
				Pods: kubernetes.PodList{eksPod()},
			}

			clientBuilder.EXPECT().GetClientInstance(&secretRef1, region).Times(1).Return(awsClient, nil)
			clientBuilder.EXPECT().GetClientInstance(&secretRef2, region).Times(1).Return(awsClient, nil)
			meshName := "test-mesh"
			awsClient.EXPECT().ListMeshes(gomock.Any()).Times(2).Return([]string{meshName}, nil)

			err := appmeshDiscovery.Sync(ctx, snap)
			Expect(err).NotTo(HaveOccurred())
			Expect(reconciler.reconcileCalledWith).To(HaveLen(1))
			Expect(reconciler.reconcileCalledWith[0]).To(HaveLen(1))
			Expect(reconciler.reconcileCalledWith[0][0]).To(Equal(expectedMesh(meshName, secret2.Metadata.Name, nil)))
		})
	})
	Context("eks pods present, 2 aws secrets that return different meshes", func() {
		It("reconciles a mesh for each secret", func() {
			secret1 := createSecret("secret1", true)
			secret2 := createSecret("secret2", true)
			secretRef1 := secret1.Metadata.Ref()
			secretRef2 := secret2.Metadata.Ref()
			snap := &v1.DiscoverySnapshot{
				Pods: kubernetes.PodList{eksPod()},
			}

			awsClient1 := mocks.NewMockClient(ctrl)
			awsClient2 := mocks.NewMockClient(ctrl)

			clientBuilder.EXPECT().GetClientInstance(&secretRef1, region).Times(1).Return(awsClient1, nil)
			clientBuilder.EXPECT().GetClientInstance(&secretRef2, region).Times(1).Return(awsClient2, nil)

			meshName1 := "test-mesh"
			meshName2 := "another-mesh"

			awsClient1.EXPECT().ListMeshes(gomock.Any()).Times(2).Return([]string{meshName1}, nil)
			awsClient2.EXPECT().ListMeshes(gomock.Any()).Times(2).Return([]string{meshName2}, nil)

			err := appmeshDiscovery.Sync(ctx, snap)
			Expect(err).NotTo(HaveOccurred())
			Expect(reconciler.reconcileCalledWith).To(HaveLen(1))
			Expect(reconciler.reconcileCalledWith[0]).To(HaveLen(2))
			Expect(reconciler.reconcileCalledWith[0][0]).To(Equal(expectedMesh(meshName1, secret1.Metadata.Name, nil)))
			Expect(reconciler.reconcileCalledWith[0][1]).To(Equal(expectedMesh(meshName2, secret2.Metadata.Name, nil)))
		})
	})
	Context("api mesh present, with injected pods", func() {
		It("reconciles a mesh with the correct upstream list", func() {
			secret := createSecret(secretName, true)
			secretRef := secret.Metadata.Ref()

			scenario := scenarios.InitializeOnly()
			injectedPods := scenario.GetResources().MustGetPodList()
			injectedUpstreams := scenario.GetResources().MustGetUpstreams()

			snap := &v1.DiscoverySnapshot{
				Pods:      append(injectedPods, eksPod()),
				Upstreams: injectedUpstreams,
			}

			clientBuilder.EXPECT().GetClientInstance(&secretRef, region).Times(1).Return(awsClient, nil)
			meshName := "test-mesh"
			awsClient.EXPECT().ListMeshes(gomock.Any()).Times(1).Return([]string{meshName}, nil)

			err := appmeshDiscovery.Sync(ctx, snap)
			Expect(err).NotTo(HaveOccurred())
			Expect(reconciler.reconcileCalledWith).To(HaveLen(1))
			Expect(reconciler.reconcileCalledWith[0]).To(HaveLen(1))
			Expect(reconciler.reconcileCalledWith[0][0]).To(Equal(expectedMesh(meshName, secret.Metadata.Name, []*core.ResourceRef{
				{
					Name:      "default-details-9080",
					Namespace: "gloo-system",
				},
				{
					Name:      "default-details-v1-9080",
					Namespace: "gloo-system",
				},
				{
					Name:      "default-productpage-9080",
					Namespace: "gloo-system",
				},
				{
					Name:      "default-productpage-v1-9080",
					Namespace: "gloo-system",
				},
				{
					Name:      "default-ratings-9080",
					Namespace: "gloo-system",
				},
				{
					Name:      "default-ratings-v1-9080",
					Namespace: "gloo-system",
				},
				{
					Name:      "default-reviews-9080",
					Namespace: "gloo-system",
				},
				{
					Name:      "default-reviews-v2-9080",
					Namespace: "gloo-system",
				},
				{
					Name:      "default-reviews-v3-9080",
					Namespace: "gloo-system",
				},
			})))
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

func eksPod() *kubernetes.Pod {
	return pod.FromKubePod(
		&kubev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "aws-node-1234",
				Namespace: "kube-system",
			},
			Spec: kubev1.PodSpec{
				Containers: []kubev1.Container{
					{
						Name:  "aws-node",
						Image: "602401143452.dkr.ecr.us-east-1.amazonaws.com/amazon-k8s-cni:v1.3.2",
					},
				},
			},
		})
}
