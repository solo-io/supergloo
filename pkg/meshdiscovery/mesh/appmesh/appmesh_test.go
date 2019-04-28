package appmesh

import (
	"context"
	"fmt"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/errors"
	kubernetes "github.com/solo-io/solo-kit/pkg/api/external/kubernetes/configmap"
	"github.com/solo-io/solo-kit/pkg/api/external/kubernetes/pod"
	skclients "github.com/solo-io/solo-kit/pkg/api/v1/clients"
	skkube "github.com/solo-io/solo-kit/pkg/api/v1/resources/common/kubernetes"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/supergloo/cli/pkg/helpers/clients"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	mocks "github.com/solo-io/supergloo/pkg/config/appmesh/mocks"
	kubev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("appmesh", func() {
	const (
		defaultRegion = "us-east-1"
	)
	var (
		ctrl *gomock.Controller

		ctx           context.Context
		clientBuilder *mocks.MockClientBuilder
		awsClient     *mocks.MockClient
		meshDiscovery *appmeshDiscoverySyncer

		secretClient gloov1.SecretClient
	)

	var (
		pods = skkube.PodsByNamespace{
			"": skkube.PodList{
				pod.FromKubePod(&kubev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      awsNode,
						Namespace: kubeSysyem,
					},
					Spec: kubev1.PodSpec{
						Containers: []kubev1.Container{
							{
								Name:  awsNode,
								Image: "602401143452.dkr.ecr.us-east-1.amazonaws.com/amazon-k8s-cni:v1.3.2",
							},
						},
					},
				}),
			},
		}
		configMaps = skkube.ConfigmapsByNamespace{
			"": skkube.ConfigMapList{
				kubernetes.FromKubeConfigMap(&kubev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      awsConfigMap,
						Namespace: kubeSysyem,
					},
				}),
			},
		}
	)

	BeforeEach(func() {
		clients.UseMemoryClients()
		secretClient = clients.MustSecretClient()
		ctx = context.TODO()
		ctrl = gomock.NewController(T)
		clientBuilder = mocks.NewMockClientBuilder(ctrl)
		awsClient = mocks.NewMockClient(ctrl)
		meshDiscovery = NewAppmeshDiscoverySyncer(clientBuilder, secretClient)
	})

	var createSecret = func(aws bool) *gloov1.Secret {
		secret := &gloov1.Secret{
			Metadata: core.Metadata{
				Name: awsConfigMap,
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
		_, err := secretClient.Write(secret, skclients.WriteOpts{})
		Expect(err).NotTo(HaveOccurred())
		return secret
	}

	Context("discover meshes", func() {
		It("returns nothing when neither pods nor config maps exist", func() {
			meshes, err := meshDiscovery.DiscoverMeshes(ctx, &v1.DiscoverySnapshot{})
			Expect(err).NotTo(HaveOccurred())
			Expect(meshes).To(BeNil())
		})
		It("returns nothing when config map exists, but no pods", func() {
			meshes, err := meshDiscovery.DiscoverMeshes(ctx, &v1.DiscoverySnapshot{
				Pods: pods,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(meshes).To(BeNil())
		})
		It("returns nothing when pod exists but no config map", func() {
			meshes, err := meshDiscovery.DiscoverMeshes(ctx, &v1.DiscoverySnapshot{
				Configmaps: configMaps,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(meshes).To(BeNil())
		})

		It("returns one new mesh when all inputs are correct, aws is mocked", func() {
			secret := createSecret(true)
			secretRef := secret.Metadata.Ref()
			snap := &v1.DiscoverySnapshot{
				Pods:       pods,
				Configmaps: configMaps,
			}
			newCtx := contextutils.WithLogger(ctx, fmt.Sprintf("appmesh-mesh-discovery-sync-%v", snap.Hash()))
			clientBuilder.EXPECT().GetClientInstance(&secretRef, defaultRegion).Times(1).Return(awsClient, nil)
			awsClient.EXPECT().ListMeshes(newCtx).Times(1).Return([]string{awsConfigMap}, nil)
			meshes, err := meshDiscovery.DiscoverMeshes(ctx, snap)
			Expect(err).NotTo(HaveOccurred())
			Expect(meshes).To(HaveLen(1))
			mesh := meshes[0]
			Expect(mesh.Metadata).To(BeEquivalentTo(core.Metadata{
				Name:      awsConfigMap,
				Labels:    DiscoverySelector,
				Namespace: "supergloo-system",
			}))
			Expect(mesh.DiscoveryMetadata).To(Equal(&v1.DiscoveryMetadata{}))
			appMesh := mesh.GetAwsAppMesh()
			Expect(appMesh).NotTo(BeNil())
			Expect(appMesh).To(BeEquivalentTo(&v1.AwsAppMesh{
				EnableAutoInject: true,
				Region:           defaultRegion,
				AwsSecret:        &secretRef,
			}))
		})

	})

	Context("finding a secret", func() {
		It("will error if no secrets exist", func() {
			ads := &appmeshDiscoverySyncer{
				secrets: secretClient,
			}
			_, err := ads.getAwsSecret()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("could not find an AWS secret to use for discovery"))
		})
		It("will error if no aws secrets exist", func() {
			createSecret(false)
			ads := &appmeshDiscoverySyncer{
				secrets: secretClient,
			}
			_, err := ads.getAwsSecret()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("could not find an AWS secret to use for discovery"))

		})
		It("will correctly find an aws secret", func() {
			createSecret(true)
			ads := &appmeshDiscoverySyncer{
				secrets: secretClient,
			}
			secret, err := ads.getAwsSecret()
			Expect(err).NotTo(HaveOccurred())
			Expect(secret.Metadata.Name).To(Equal(awsConfigMap))
		})
	})

	Context("appmesh discovery data", func() {
		It("returns nil when nothing exists", func() {
			amd := newAppmeshDiscoveryData(ctx, skkube.ConfigMapList{}, skkube.PodList{})
			Expect(amd).To(BeEquivalentTo(&appmeshDiscoveryData{}))
		})
		It("all correct objects when valid", func() {
			amd := newAppmeshDiscoveryData(ctx, configMaps.List(), pods.List())
			Expect(amd.configMaps).To(HaveLen(1))
			Expect(amd.pods).To(HaveLen(1))
			Expect(amd.region).To(Equal(defaultRegion))
		})
	})

	Context("region finder", func() {
		It("can find the region properly when it exists", func() {
			region, err := awsRegionFromImage("602401143452.dkr.ecr.us-east-1.amazonaws.com/amazon-k8s-cni:v1.3.2")
			Expect(err).NotTo(HaveOccurred())
			Expect(region).To(Equal(defaultRegion))
		})
		It("returns an error if no region is found", func() {
			_, err := awsRegionFromImage("602401143452.dkr.ecr.us-east1.amazonaws.com/amazon-k8s-cni:v1.3.2")
			Expect(err).To(HaveOccurred())
		})
	})

	Context("construct aws meshes", func() {
		It("returns empty list if no meshes can be found", func() {
			awsClient.EXPECT().ListMeshes(ctx).Times(1).Return([]string{}, nil)
			_, err := constructAwsMeshes(ctx, awsClient, defaultRegion, &core.ResourceRef{})
			Expect(err).To(BeNil())
		})
		It("retuns an error if the list mesh call fails", func() {
			awsClient.EXPECT().ListMeshes(ctx).Times(1).Return(nil, errors.Errorf("this is an error"))
			_, err := constructAwsMeshes(ctx, awsClient, defaultRegion, &core.ResourceRef{})
			Expect(err).To(HaveOccurred())
		})
		It("retuns one Mesh object per app mesh returned", func() {
			awsClient.EXPECT().ListMeshes(ctx).Times(1).Return([]string{"1", "2", "3", "4", "5"}, nil)
			meshes, err := constructAwsMeshes(ctx, awsClient, defaultRegion, &core.ResourceRef{})
			Expect(err).NotTo(HaveOccurred())
			Expect(meshes).To(HaveLen(5))
		})
		It("mesh is populated properly", func() {
			awsClient.EXPECT().ListMeshes(ctx).Times(1).Return([]string{"default"}, nil)
			meshes, err := constructAwsMeshes(ctx, awsClient, defaultRegion, &core.ResourceRef{})
			Expect(err).NotTo(HaveOccurred())
			Expect(meshes).To(HaveLen(1))
			mesh := meshes[0]
			Expect(mesh.Metadata).To(BeEquivalentTo(core.Metadata{
				Name:      "default",
				Labels:    DiscoverySelector,
				Namespace: "supergloo-system",
			}))
			Expect(mesh.DiscoveryMetadata).To(Equal(&v1.DiscoveryMetadata{}))
			appMesh := mesh.GetAwsAppMesh()
			Expect(appMesh).NotTo(BeNil())
			Expect(appMesh).To(BeEquivalentTo(&v1.AwsAppMesh{
				EnableAutoInject: true,
				Region:           defaultRegion,
				AwsSecret:        &core.ResourceRef{},
			}))
		})
	})
})
