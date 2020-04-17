package linkerd_test

import (
	"context"

	linkerdconfig "github.com/linkerd/linkerd2/controller/gen/config"
	"github.com/linkerd/linkerd2/pkg/config"
	"github.com/solo-io/service-mesh-hub/test/fakes"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/testutils"
	zephyr_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/common/docker"
	mock_docker "github.com/solo-io/service-mesh-hub/pkg/common/docker/mocks"
	"github.com/solo-io/service-mesh-hub/pkg/env"
	"github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/mesh/linkerd"
	k8s_apps_types "k8s.io/api/apps/v1"
	k8s_core_types "k8s.io/api/core/v1"
	k8s_meta_types "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Linkerd Mesh Scanner", func() {
	var (
		ctrl        *gomock.Controller
		ctx         context.Context
		linkerdNs   = "linkerd"
		client      client.Client
		clusterName = "test-cluster-name"
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()

		client = fakes.InMemoryClient()
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("does not discover linkerd when it is not there", func() {
		imageParser := mock_docker.NewMockImageNameParser(ctrl)

		scanner := linkerd.NewLinkerdMeshScanner(imageParser)

		deployment := &k8s_apps_types.Deployment{
			ObjectMeta: k8s_meta_types.ObjectMeta{Namespace: linkerdNs, Name: "name doesn't matter in this context"},
			Spec: k8s_apps_types.DeploymentSpec{
				Template: k8s_core_types.PodTemplateSpec{
					Spec: k8s_core_types.PodSpec{
						Containers: []k8s_core_types.Container{
							{
								Image: "test-image",
							},
						},
					},
				},
			},
		}

		mesh, err := scanner.ScanDeployment(ctx, clusterName, deployment, client)
		Expect(err).NotTo(HaveOccurred())
		Expect(mesh).To(BeNil())
	})

	It("discovers linkerd", func() {

		linkerdConfigMap := func() *k8s_core_types.ConfigMap {
			cfg := &linkerdconfig.All{
				Global: &linkerdconfig.Global{
					ClusterDomain: "applebees.com",
				},
				Proxy:   &linkerdconfig.Proxy{},
				Install: &linkerdconfig.Install{},
			}
			global, proxy, install, err := config.ToJSON(cfg)
			Expect(err).NotTo(HaveOccurred())
			return &k8s_core_types.ConfigMap{
				ObjectMeta: k8s_meta_types.ObjectMeta{Name: linkerd.LinkerdConfigMapName, Namespace: linkerdNs},
				Data: map[string]string{
					"global":  global,
					"proxy":   proxy,
					"install": install,
				},
			}
		}()

		err := client.Create(ctx, linkerdConfigMap)
		Expect(err).NotTo(HaveOccurred())

		imageParser := mock_docker.NewMockImageNameParser(ctrl)

		scanner := linkerd.NewLinkerdMeshScanner(imageParser)

		deployment := &k8s_apps_types.Deployment{
			ObjectMeta: k8s_meta_types.ObjectMeta{Namespace: linkerdNs, Name: "name doesn't matter in this context"},
			Spec: k8s_apps_types.DeploymentSpec{
				Template: k8s_core_types.PodTemplateSpec{
					Spec: k8s_core_types.PodSpec{
						Containers: []k8s_core_types.Container{
							{
								Image: "linkerd-io/controller:0.6.9",
							},
						},
					},
				},
			},
		}

		imageParser.
			EXPECT().
			Parse("linkerd-io/controller:0.6.9").
			Return(&docker.Image{
				Domain: "docker.io",
				Path:   "linkerd-io/controller",
				Tag:    "0.6.9",
			}, nil)

		expectedMesh := &zephyr_discovery.Mesh{
			ObjectMeta: k8s_meta_types.ObjectMeta{
				Name:      "linkerd-linkerd-" + clusterName,
				Namespace: env.GetWriteNamespace(),
				Labels:    linkerd.DiscoveryLabels,
			},
			Spec: zephyr_discovery_types.MeshSpec{
				MeshType: &zephyr_discovery_types.MeshSpec_Linkerd{
					Linkerd: &zephyr_discovery_types.MeshSpec_LinkerdMesh{
						Installation: &zephyr_discovery_types.MeshSpec_MeshInstallation{
							InstallationNamespace: deployment.GetNamespace(),
							Version:               "0.6.9",
						},
						ClusterDomain: "applebees.com",
					},
				},
				Cluster: &zephyr_core_types.ResourceRef{
					Name:      clusterName,
					Namespace: env.GetWriteNamespace(),
				},
			},
		}

		mesh, err := scanner.ScanDeployment(ctx, clusterName, deployment, client)
		Expect(err).NotTo(HaveOccurred())
		Expect(mesh).To(Equal(expectedMesh))
	})

	It("reports an error when the image name is un-parseable", func() {
		imageParser := mock_docker.NewMockImageNameParser(ctrl)

		scanner := linkerd.NewLinkerdMeshScanner(imageParser)

		deployment := &k8s_apps_types.Deployment{
			ObjectMeta: k8s_meta_types.ObjectMeta{Namespace: linkerdNs, Name: "name doesn't matter in this context"},
			Spec: k8s_apps_types.DeploymentSpec{
				Template: k8s_core_types.PodTemplateSpec{
					Spec: k8s_core_types.PodSpec{
						Containers: []k8s_core_types.Container{
							{
								Image: "linkerd-io/controller:0.6.9",
							},
						},
					},
				},
			},
		}

		testErr := eris.New("test-err")

		imageParser.
			EXPECT().
			Parse("linkerd-io/controller:0.6.9").
			Return(nil, testErr)

		mesh, err := scanner.ScanDeployment(ctx, clusterName, deployment, client)
		Expect(mesh).To(BeNil())
		Expect(err).To(testutils.HaveInErrorChain(testErr))
	})
})
