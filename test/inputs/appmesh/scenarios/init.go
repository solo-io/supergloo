package scenarios

import (
	appmeshApi "github.com/aws/aws-sdk-go/service/appmesh"
	. "github.com/onsi/gomega"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/translator/appmesh"
	appmeshInputs "github.com/solo-io/supergloo/test/inputs/appmesh"
)

type initOnlyScenario struct {
	meshName string
	allResources,
	appMeshResources appmeshInputs.TestResourceSet
	expectedVirtualNodes map[string]*appmeshApi.VirtualNodeData
}

func InitializeOnly() AppMeshTestScenario {
	return &initOnlyScenario{
		meshName:             MeshName,
		allResources:         GetAllResources(),
		appMeshResources:     GetAppMeshRelatedResources(),
		expectedVirtualNodes: initVirtualNodes(),
	}
}

func (s *initOnlyScenario) GetMeshName() string {
	return s.meshName
}

func (s *initOnlyScenario) GetResources() appmeshInputs.TestResourceSet {
	return s.allResources
}

func (s *initOnlyScenario) GetRoutingRules() v1.RoutingRuleList {
	return nil
}

func (s *initOnlyScenario) VerifyExpectations(iConfig appmesh.AwsAppMeshConfiguration) {
	config, ok := iConfig.(*appmesh.AwsAppMeshConfigurationImpl)
	ExpectWithOffset(1, ok).To(BeTrue())

	ExpectWithOffset(1, config.MeshName).To(BeEquivalentTo(MeshName))
	ExpectWithOffset(1, config.PodList).To(ConsistOf(s.appMeshResources.MustGetPodList()))
	ExpectWithOffset(1, config.UpstreamList).To(ConsistOf(s.appMeshResources.MustGetUpstreams()))

	ExpectWithOffset(1, config.VirtualNodes).To(HaveLen(6))

	for hostname, expectedVn := range s.expectedVirtualNodes {
		vn, ok := config.VirtualNodes[hostname]
		ExpectWithOffset(1, ok).To(BeTrue())
		ExpectWithOffset(1, vn.MeshName).To(BeEquivalentTo(expectedVn.MeshName))
		ExpectWithOffset(1, vn.VirtualNodeName).To(BeEquivalentTo(expectedVn.VirtualNodeName))
		ExpectWithOffset(1, vn.Spec.Backends).To(HaveLen(0))
		ExpectWithOffset(1, vn.Spec.Listeners).To(ConsistOf(expectedVn.Spec.Listeners))
		ExpectWithOffset(1, vn.Spec.ServiceDiscovery).To(BeEquivalentTo(expectedVn.Spec.ServiceDiscovery))
	}
}

// Returns the virtual node set as it is before any Routing Rules have been processed or all traffic has been allowed
func initVirtualNodes() map[string]*appmeshApi.VirtualNodeData {
	return map[string]*appmeshApi.VirtualNodeData{
		productPageHostname: createVirtualNode(productPageVnName, productPageHostname, MeshName, "http", 9080, nil),
		detailsHostname:     createVirtualNode(detailsVnName, detailsHostname, MeshName, "http", 9080, nil),
		reviewsV1Hostname:   createVirtualNode(reviewsV1VnName, reviewsV1Hostname, MeshName, "http", 9080, nil),
		reviewsV2Hostname:   createVirtualNode(reviewsV2VnName, reviewsV2Hostname, MeshName, "http", 9080, nil),
		reviewsV3Hostname:   createVirtualNode(reviewsV3VnName, reviewsV3Hostname, MeshName, "http", 9080, nil),
		ratingsHostname:     createVirtualNode(ratingsVnName, ratingsHostname, MeshName, "http", 9080, nil),
	}
}
