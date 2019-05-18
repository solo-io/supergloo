package mesh_test

import (
	"context"

	"github.com/solo-io/solo-kit/pkg/api/v1/resources/common/kubernetes"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/supergloo/cli/pkg/helpers/clients"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/meshdiscovery/mesh"
	"github.com/solo-io/supergloo/pkg/meshdiscovery/mesh/istio"
)

var _ = Describe("discovery syncer", func() {
	var (
		meshClient v1.MeshClient
	)
	BeforeEach(func() {
		clients.UseMemoryClients()
		meshClient = clients.MustMeshClient()
	})
	It("can run without erroring", func() {
		ctx := context.TODO()
		snap := &v1.DiscoverySnapshot{}
		mockPlugin := istio.NewMockIstioMeshDiscovery()
		syncer := mesh.NewMeshDiscoverySyncer(meshClient, mockPlugin)
		err := syncer.Sync(ctx, snap)
		Expect(err).NotTo(HaveOccurred())
	})

	It("filters out namespaces correctly", func() {
		install1 := &v1.Install{
			Metadata: core.Metadata{
				Namespace: "doesn't matter",
			},
			InstallationNamespace: "foo",
			Disabled:              true,
		}
		install2 := &v1.Install{
			Metadata: core.Metadata{
				Namespace: "doesn't matter",
			},
			InstallationNamespace: "bar",
		}
		pod1 := kubernetes.NewPod("foo", "test1")
		pod2 := kubernetes.NewPod("bar", "test2")
		configMap1 := kubernetes.NewConfigMap("foo", "test1")
		configMap2 := kubernetes.NewConfigMap("bar", "test2")
		input := &v1.DiscoverySnapshot{
			Installs:   v1.InstallList{install1, install2},
			Pods:       kubernetes.PodList{pod1, pod2},
			Configmaps: kubernetes.ConfigMapList{configMap1, configMap2},
		}
		expected := &v1.DiscoverySnapshot{
			Installs:   v1.InstallList{install1, install2},
			Pods:       kubernetes.PodList{pod2},
			Configmaps: kubernetes.ConfigMapList{configMap2},
		}
		actual := mesh.FilterOutNamespacesWithInstalls(input)
		Expect(actual).To(BeEquivalentTo(expected))
	})
})
