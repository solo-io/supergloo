package apply_test

import (
	"fmt"
	"time"

	"github.com/gogo/protobuf/types"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	skclients "github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/supergloo/cli/pkg/helpers/clients"
	"github.com/solo-io/supergloo/cli/test/utils"
	"github.com/solo-io/supergloo/pkg/api/external/istio/networking/v1alpha3"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
)

var _ = Describe("Retries", func() {

	BeforeEach(func() {
		clients.UseMemoryClients()
		_, err := clients.MustMeshClient().Write(&v1.Mesh{Metadata: core.Metadata{Namespace: "my", Name: "mesh"}}, skclients.WriteOpts{})
		Expect(err).NotTo(HaveOccurred())
	})

	Context("maxRetries", func() {
		maxRetries := func(name, attempts, timeout, retryOn string) error {
			return utils.Supergloo(fmt.Sprintf("apply routingrule retries max --name %v --attempts %v --per-try-timeout %v --retry-on %v --target-mesh my.mesh", name, attempts, timeout, retryOn))
		}

		Context("valid args", func() {
			It("produces the expected retry policy", func() {
				err := maxRetries("test", "3", "1m", "5xx,gateway-error")
				Expect(err).NotTo(HaveOccurred())
				rr, err := clients.MustRoutingRuleClient().Read("supergloo-system", "test", skclients.ReadOpts{})
				Expect(err).NotTo(HaveOccurred())
				rr.Metadata.ResourceVersion = ""
				Expect(rr).To(Equal(&v1.RoutingRule{
					Metadata: core.Metadata{
						Name:      "test",
						Namespace: "supergloo-system",
					},
					TargetMesh: &core.ResourceRef{Name: "mesh", Namespace: "my"},
					Spec: &v1.RoutingRuleSpec{
						RuleType: &v1.RoutingRuleSpec_Retries{
							Retries: &v1.RetryPolicy{
								MaxRetries: &v1alpha3.HTTPRetry{
									Attempts:      3,
									PerTryTimeout: types.DurationProto(time.Minute),
									RetryOn:       "5xx,gateway-error",
								},
								RetryBudget: nil,
							},
						},
					},
				}))
			})
		})
	})

	Context("retry budget", func() {
		budget := func(name, ratio, rps, ttl string) error {
			return utils.Supergloo(fmt.Sprintf("apply routingrule retries budget --name %v --ratio %v --min-retries %v --ttl %v --target-mesh my.mesh", name, ratio, rps, ttl))
		}

		Context("valid args", func() {
			It("produces the expected retry policy", func() {
				err := budget("test", "0.5", "4", "5m")
				Expect(err).NotTo(HaveOccurred())
				rr, err := clients.MustRoutingRuleClient().Read("supergloo-system", "test", skclients.ReadOpts{})
				Expect(err).NotTo(HaveOccurred())
				rr.Metadata.ResourceVersion = ""
				Expect(rr).To(Equal(&v1.RoutingRule{
					Metadata: core.Metadata{
						Name:      "test",
						Namespace: "supergloo-system",
					},
					TargetMesh: &core.ResourceRef{Name: "mesh", Namespace: "my"},
					Spec: &v1.RoutingRuleSpec{
						RuleType: &v1.RoutingRuleSpec_Retries{
							Retries: &v1.RetryPolicy{
								RetryBudget: &v1.RetryBudget{
									RetryRatio:          0.5,
									MinRetriesPerSecond: 4,
									Ttl:                 time.Minute * 5,
								},
							},
						},
					},
				}))
			})
		})
		Context("merge", func() {
			BeforeEach(func() {
				_, err := clients.MustRoutingRuleClient().Write(&v1.RoutingRule{
					Metadata: core.Metadata{
						Name:      "test",
						Namespace: "supergloo-system",
					},
					TargetMesh: &core.ResourceRef{Name: "mesh", Namespace: "my"},
					Spec: &v1.RoutingRuleSpec{
						RuleType: &v1.RoutingRuleSpec_Retries{
							Retries: &v1.RetryPolicy{
								MaxRetries: &v1alpha3.HTTPRetry{
									Attempts:      3,
									PerTryTimeout: types.DurationProto(time.Minute),
									RetryOn:       "5xx,gateway-error",
								},
								RetryBudget: nil,
							},
						},
					},
				}, skclients.WriteOpts{})
				Expect(err).NotTo(HaveOccurred())
			})
			It("merges the retry policies", func() {
				err := budget("test", "0.5", "4", "5m")
				Expect(err).NotTo(HaveOccurred())
				rr, err := clients.MustRoutingRuleClient().Read("supergloo-system", "test", skclients.ReadOpts{})
				Expect(err).NotTo(HaveOccurred())
				rr.Metadata.ResourceVersion = ""
				Expect(rr).To(Equal(&v1.RoutingRule{
					Metadata: core.Metadata{
						Name:      "test",
						Namespace: "supergloo-system",
					},
					TargetMesh: &core.ResourceRef{Name: "mesh", Namespace: "my"},
					Spec: &v1.RoutingRuleSpec{
						RuleType: &v1.RoutingRuleSpec_Retries{
							Retries: &v1.RetryPolicy{
								MaxRetries: &v1alpha3.HTTPRetry{
									Attempts:      3,
									PerTryTimeout: types.DurationProto(time.Minute),
									RetryOn:       "5xx,gateway-error",
								},
								RetryBudget: &v1.RetryBudget{
									RetryRatio:          0.5,
									MinRetriesPerSecond: 4,
									Ttl:                 time.Minute * 5,
								},
							},
						},
					},
				}))
			})
		})
	})
})
