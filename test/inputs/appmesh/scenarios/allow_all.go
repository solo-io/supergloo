package scenarios

import (
	appmeshApi "github.com/aws/aws-sdk-go/service/appmesh"
	. "github.com/onsi/gomega"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/translator/appmesh"
	appmeshInputs "github.com/solo-io/supergloo/test/inputs/appmesh"
)

type allowAllOnlyScenario struct {
	meshName     string
	allResources appmeshInputs.TestResourceSet
}

func AllowAllOnly() AppMeshTestScenario {
	return &allowAllOnlyScenario{
		meshName:     MeshName,
		allResources: GetAllResources(),
	}
}

func (s *allowAllOnlyScenario) GetResources() appmeshInputs.TestResourceSet {
	return s.allResources
}

func (s *allowAllOnlyScenario) GetMeshName() string {
	return s.meshName
}

func (s *allowAllOnlyScenario) GetRoutingRules() v1.RoutingRuleList {
	return nil
}

func (s *allowAllOnlyScenario) VerifyExpectations(configuration appmesh.AwsAppMeshConfiguration) {
	config, ok := configuration.(*appmesh.AwsAppMeshConfigurationImpl)
	ExpectWithOffset(1, ok).To(BeTrue())

	// Verify virtual nodes
	ExpectWithOffset(1, config.VirtualNodes).To(HaveLen(6))
	for hostname, expectedVn := range s.expectedVirtualNodes() {

		vn, ok := config.VirtualNodes[hostname]
		ExpectWithOffset(1, ok).To(BeTrue())
		ExpectWithOffset(1, vn.MeshName).To(BeEquivalentTo(expectedVn.MeshName))
		ExpectWithOffset(1, vn.VirtualNodeName).To(BeEquivalentTo(expectedVn.VirtualNodeName))
		ExpectWithOffset(1, vn.Spec.Listeners).To(ConsistOf(expectedVn.Spec.Listeners))
		ExpectWithOffset(1, vn.Spec.ServiceDiscovery).To(BeEquivalentTo(expectedVn.Spec.ServiceDiscovery))
		ExpectWithOffset(1, vn.Spec.Backends).To(ConsistOf(expectedVn.Spec.Backends))
	}

	// Verify virtual services
	ExpectWithOffset(1, config.VirtualServices).To(HaveLen(6))
	for hostname, expectedVs := range s.expectedVirtualServices() {
		vs, ok := config.VirtualServices[hostname]
		ExpectWithOffset(1, ok).To(BeTrue())
		ExpectWithOffset(1, vs.VirtualServiceName).To(BeEquivalentTo(expectedVs.VirtualServiceName))
		ExpectWithOffset(1, vs.MeshName).To(BeEquivalentTo(expectedVs.MeshName))
		ExpectWithOffset(1, vs.Spec.Provider.VirtualNode).To(BeEquivalentTo(expectedVs.Spec.Provider.VirtualNode))
		ExpectWithOffset(1, vs.Spec.Provider.VirtualRouter).To(BeNil())
	}
}

// Returns the virtual node set as it is expected to be after allowing all traffic (no Routing Rules)
func (s *allowAllOnlyScenario) expectedVirtualNodes() map[string]*appmeshApi.VirtualNodeData {
	return map[string]*appmeshApi.VirtualNodeData{
		productPageHostname: createVirtualNode(productPageVnName, productPageHostname, MeshName, "http", 9080, allHostsMinus(productPageHostname)),
		detailsHostname:     createVirtualNode(detailsVnName, detailsHostname, MeshName, "http", 9080, allHostsMinus(detailsHostname)),
		reviewsV1Hostname:   createVirtualNode(reviewsV1VnName, reviewsV1Hostname, MeshName, "http", 9080, allHostsMinus(reviewsV1Hostname)),
		reviewsV2Hostname:   createVirtualNode(reviewsV2VnName, reviewsV2Hostname, MeshName, "http", 9080, allHostsMinus(reviewsV2Hostname)),
		reviewsV3Hostname:   createVirtualNode(reviewsV3VnName, reviewsV3Hostname, MeshName, "http", 9080, allHostsMinus(reviewsV3Hostname)),
		ratingsHostname:     createVirtualNode(ratingsVnName, ratingsHostname, MeshName, "http", 9080, allHostsMinus(ratingsHostname)),
	}
}

// Returns the virtual node set as it is expected to be after allowing all traffic (no Routing Rules)
func (s *allowAllOnlyScenario) expectedVirtualServices() map[string]*appmeshApi.VirtualServiceData {
	return map[string]*appmeshApi.VirtualServiceData{
		productPageHostname: createVirtualServiceWithVn(productPageHostname, MeshName, productPageVnName),
		detailsHostname:     createVirtualServiceWithVn(detailsHostname, MeshName, detailsVnName),
		reviewsV1Hostname:   createVirtualServiceWithVn(reviewsV1Hostname, MeshName, reviewsV1VnName),
		reviewsV2Hostname:   createVirtualServiceWithVn(reviewsV2Hostname, MeshName, reviewsV2VnName),
		reviewsV3Hostname:   createVirtualServiceWithVn(reviewsV3Hostname, MeshName, reviewsV3VnName),
		ratingsHostname:     createVirtualServiceWithVn(ratingsHostname, MeshName, ratingsVnName),
	}
}
