package version_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/kubeutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/supergloo/cli/pkg/common"
	v1 "github.com/solo-io/supergloo/pkg2api/v1"
	"github.com/solo-io/supergloo/test/utils"
)

var _ = Describe("Version", func() {

	It("are you outta ya mind!?", func() {
		namespace := "supergloo-system"

		cfg, err := kubeutils.GetConfig("", "")
		Expect(err).NotTo(HaveOccurred())
		// deploy bookinfo
		if false {
			err = utils.DeployBookinfo(cfg, "istio-system")
			Expect(err).NotTo(HaveOccurred())
		}

		err = utils.DeployTestRunner(cfg, "istio-system")
		Expect(err).NotTo(HaveOccurred())

		// get preinstalled meshes
		meshC, err := common.GetMeshClient()
		Expect(err).NotTo(HaveOccurred())

		meshes, err := (*meshC).List(namespace, clients.ListOpts{})
		Expect(err).NotTo(HaveOccurred())
		// add a route rule for each one
		for _, m := range meshes {
			ref := m.Metadata.Ref()
			rrc, err := common.GetRoutingRuleClient()
			Expect(err).NotTo(HaveOccurred())
			err = (*rrc).Delete(namespace, "hi", clients.DeleteOpts{})
			Expect(err).NotTo(HaveOccurred())
			_, err = (*rrc).Write(&v1.RoutingRule{
				Metadata:   core.Metadata{Namespace: namespace, Name: "hi"},
				TargetMesh: &ref,
				Destinations: []*core.ResourceRef{{
					Name:      "istio-system-reviews-9080",
					Namespace: namespace,
				}},
				TrafficShifting: &v1.TrafficShifting{
					Destinations: []*v1.WeightedDestination{
						{
							Upstream: &core.ResourceRef{
								Name:      "istio-system-reviews-v1-9080",
								Namespace: "supergloo-system",
							},
							Weight: 100,
						},
					},
				},
			}, clients.WriteOpts{})
			Expect(err).NotTo(HaveOccurred())
		}
	})
})
