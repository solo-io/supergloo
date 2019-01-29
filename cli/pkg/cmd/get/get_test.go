package get_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/kubeutils"
	"github.com/solo-io/go-utils/testutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/solo-kit/test/setup"
	"github.com/solo-io/supergloo/cli/test/utils"
	"github.com/solo-io/supergloo/pkg/api/v1"
)

var _ = Describe("Get", func() {
	var (
		namespace         string
	)
	BeforeEach(func() {
		namespace = "a" + testutils.RandString(6)
		cfg, err := kubeutils.GetConfig("", "")
		Expect(err).NotTo(HaveOccurred())
		err = setup.SetupKubeForTest(namespace)
		Expect(err).NotTo(HaveOccurred())
		routingRuleClient, err := v1.NewRoutingRuleClient(&factory.KubeResourceClientFactory{
			Crd:         v1.RoutingRuleCrd,
			Cfg:         cfg,
			SharedCache: kube.NewKubeCache(),
		})
		Expect(err).NotTo(HaveOccurred())
		writeRouteRule(routingRuleClient, namespace)
	})
	AfterEach(func() {
		setup.TeardownKube(namespace)
	})
	It("does not panic on table output", func() {
		err := utils.Supergloo("get routingrule")
		Expect(err).NotTo(HaveOccurred())
	})
	It("does not panic on yaml output", func() {
		err := utils.Supergloo("get routingrule  -o yaml")
		Expect(err).NotTo(HaveOccurred())
	})
})

func writeRouteRule(routingRules v1.RoutingRuleClient, namespace string) {
	rrMeta := core.Metadata{Name: "rrrr", Namespace: namespace}
	rr1, err := routingRules.Write(&v1.RoutingRule{
		Metadata:   rrMeta,
		Destinations: []*core.ResourceRef{{
			Name:      namespace + "-reviews-9080",
			Namespace: namespace,
		}},
		TrafficShifting: &v1.TrafficShifting{
			Destinations: []*v1.WeightedDestination{
				{
					Upstream: &core.ResourceRef{
						Name:      namespace + "-reviews-v1-9080",
						Namespace: namespace,
					},
					Weight: 100,
				},
			},
		},
	}, clients.WriteOpts{})
	Expect(err).NotTo(HaveOccurred())
	Expect(rr1).NotTo(BeNil())
}
