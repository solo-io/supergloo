package appmesh

import (
	"context"
	"log"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	kubernetes "github.com/solo-io/solo-kit/pkg/api/external/kubernetes/configmap"
	"github.com/solo-io/solo-kit/pkg/api/external/kubernetes/pod"
	skclients "github.com/solo-io/solo-kit/pkg/api/v1/clients"
	skkube "github.com/solo-io/solo-kit/pkg/api/v1/resources/common/kubernetes"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/solo-kit/test/helpers"
	"github.com/solo-io/supergloo/cli/pkg/helpers/clients"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/config/appmesh"
	kubev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("appmesh integration", func() {
	if os.Getenv("AWS_ACCESS_KEY_ID") == "" || os.Getenv("AWS_SECRET_ACCESS_KEY") == "" {
		log.Printf("Both AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY must be set to run these tests")
		return
	}

	const (
		defaultRegion = "us-east-1"
	)
	var (
		ctx           context.Context
		clientBuilder appmesh.ClientBuilder
		awsClient     appmesh.Client
		meshDiscovery *appmeshDiscoverySyncer

		secret       *gloov1.Secret
		secretClient gloov1.SecretClient
		meshName     string
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

	var createSecret = func() *gloov1.Secret {
		accessKeyId, secretAccessKey := os.Getenv("AWS_ACCESS_KEY_ID"), os.Getenv("AWS_SECRET_ACCESS_KEY")
		secret := &gloov1.Secret{
			Metadata: core.Metadata{
				Name: awsConfigMap,
			},
			Kind: &gloov1.Secret_Aws{
				Aws: &gloov1.AwsSecret{
					AccessKey: accessKeyId,
					SecretKey: secretAccessKey,
				},
			},
		}
		createdSecret, err := secretClient.Write(secret, skclients.WriteOpts{})
		Expect(err).NotTo(HaveOccurred())
		return createdSecret
	}

	BeforeEach(func() {
		var err error
		clients.UseMemoryClients()
		secretClient = clients.MustSecretClient()
		secret = createSecret()
		ctx = context.TODO()
		clientBuilder = appmesh.NewAppMeshClientBuilder(secretClient)
		meshDiscovery = NewAppmeshDiscoverySyncer(clientBuilder, secretClient)
		secretRef := secret.Metadata.Ref()
		awsClient, err = clientBuilder.GetClientInstance(&secretRef, defaultRegion)
		Expect(err).NotTo(HaveOccurred())
		meshName = helpers.RandString(8)
		_, err = awsClient.CreateMesh(ctx, meshName)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		err := awsClient.DeleteMesh(ctx, meshName)
		Expect(err).NotTo(HaveOccurred())
	})

	Context("discover meshes", func() {
		It("returns one new mesh when all inputs are correct, aws is mocked", func() {
			secretRef := secret.Metadata.Ref()
			snap := &v1.DiscoverySnapshot{
				Pods:       pods,
				Configmaps: configMaps,
			}
			meshes, err := meshDiscovery.DiscoverMeshes(ctx, snap)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(meshes)).To(BeNumerically(">", 0))
			var targetMesh *v1.Mesh
			for _, mesh := range meshes {
				if mesh.Metadata.Name == meshName {
					targetMesh = mesh
					break
				}
			}
			Expect(targetMesh).NotTo(BeNil())
			Expect(targetMesh.Metadata).To(BeEquivalentTo(core.Metadata{
				Name:      meshName,
				Labels:    DiscoverySelector,
				Namespace: "supergloo-system",
			}))
			Expect(targetMesh.DiscoveryMetadata).To(Equal(&v1.DiscoveryMetadata{}))
			appMesh := targetMesh.GetAwsAppMesh()
			Expect(appMesh).NotTo(BeNil())
			Expect(appMesh).To(BeEquivalentTo(&v1.AwsAppMesh{
				VirtualNodeLabel: DefaultVirtualNodeLabel,
				Region:           defaultRegion,
				AwsSecret:        &secretRef,
			}))
		})

	})

})
