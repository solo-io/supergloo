package appmesh

import (
	appmeshApi "github.com/aws/aws-sdk-go/service/appmesh"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/supergloo/test/inputs/appmesh"
	"github.com/solo-io/supergloo/test/inputs/appmesh/correct"
)

var _ = Describe("Configuration", func() {

	var (
		allResources, appMeshResources appmesh.TestResourceSet
		expectedInitVirtualNodes       map[string]*appmeshApi.VirtualNodeData
		meshName                       string
	)

	Context("kubernetes resources have been configured correctly", func() {

		initConfig := func() *awsAppMeshConfiguration {
			iConfig, err := NewAwsAppMeshConfiguration(meshName, allResources.MustGetPodList(), allResources.MustGetUpstreams())
			ExpectWithOffset(1, err).NotTo(HaveOccurred())
			ExpectWithOffset(1, iConfig).NotTo(BeNil())
			config, ok := iConfig.(*awsAppMeshConfiguration)
			ExpectWithOffset(1, ok).To(BeTrue())

			ExpectWithOffset(1, config.MeshName).To(BeEquivalentTo(meshName))
			ExpectWithOffset(1, config.podList).To(ConsistOf(appMeshResources.MustGetPodList()))
			ExpectWithOffset(1, config.upstreamList).To(ConsistOf(appMeshResources.MustGetUpstreams()))

			ExpectWithOffset(1, config.VirtualNodes).To(HaveLen(6))

			for hostname, expectedVn := range expectedInitVirtualNodes {
				vn, ok := config.VirtualNodes[hostname]
				ExpectWithOffset(1, ok).To(BeTrue())
				ExpectWithOffset(1, vn.MeshName).To(BeEquivalentTo(expectedVn.MeshName))
				ExpectWithOffset(1, vn.VirtualNodeName).To(BeEquivalentTo(expectedVn.VirtualNodeName))
				ExpectWithOffset(1, vn.Spec.Backends).To(HaveLen(0))
				ExpectWithOffset(1, vn.Spec.Listeners).To(ConsistOf(expectedVn.Spec.Listeners))
				ExpectWithOffset(1, vn.Spec.ServiceDiscovery).To(BeEquivalentTo(expectedVn.Spec.ServiceDiscovery))
			}
			return config
		}

		BeforeEach(func() {
			allResources = correct.GetAllResources()
			appMeshResources = correct.GetAppMeshRelatedResources()
			expectedInitVirtualNodes = correct.GetExpectedInitVirtualNodes()
			meshName = correct.MeshName
		})

		It("correctly initializes App Mesh configuration object", func() {
			initConfig()
		})

		FIt("allowing all traffic produces the correct configuration object", func() {
			config := initConfig()

			err := config.AllowAll()
			Expect(err).NotTo(HaveOccurred())

			expected := correct.GetExpectedVirtualNodesOnlyAllowAll()

			for hostname, expectedVn := range expected {
				vn, ok := config.VirtualNodes[hostname]
				ExpectWithOffset(1, ok).To(BeTrue())
				ExpectWithOffset(1, vn.MeshName).To(BeEquivalentTo(expectedVn.MeshName))
				ExpectWithOffset(1, vn.VirtualNodeName).To(BeEquivalentTo(expectedVn.VirtualNodeName))
				ExpectWithOffset(1, vn.Spec.Listeners).To(ConsistOf(expectedVn.Spec.Listeners))
				ExpectWithOffset(1, vn.Spec.ServiceDiscovery).To(BeEquivalentTo(expectedVn.Spec.ServiceDiscovery))
				ExpectWithOffset(1, vn.Spec.Backends).To(ConsistOf(expectedVn.Spec.Backends))
			}

		})
	})

})
