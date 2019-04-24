package appmesh_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	translator "github.com/solo-io/supergloo/pkg/translator/appmesh"
	"github.com/solo-io/supergloo/test/inputs/appmesh/scenarios"
)

var _ = Describe("Configuration", func() {

	var (
		scenario scenarios.AppMeshTestScenario
		config   translator.AwsAppMeshConfiguration
		err      error
	)

	Context("kubernetes resources have been configured correctly", func() {

		JustBeforeEach(func() {
			config, err = translator.NewAwsAppMeshConfiguration(
				scenario.GetMeshName(),
				scenario.GetResources().MustGetPodList(),
				scenario.GetResources().MustGetUpstreams())
			Expect(err).NotTo(HaveOccurred())
			Expect(config).NotTo(BeNil())
		})

		When("the configuration is initialized", func() {
			BeforeEach(func() {
				scenario = scenarios.InitializeOnly()
			})

			It("produces the correct configuration object", func() {
				scenario.VerifyExpectations(config)
			})
		})

		When("when all traffic is allowed in the mesh", func() {
			BeforeEach(func() {
				scenario = scenarios.AllowAllOnly()
			})

			It("produces the correct configuration object", func() {
				err = config.AllowAll()
				Expect(err).NotTo(HaveOccurred())

				scenario.VerifyExpectations(config)
			})
		})

		When("when a single traffic shifting rule is applied", func() {
			BeforeEach(func() {
				scenario = scenarios.RoutingRule1()
			})

			It("produces the correct configuration object", func() {
				err = config.ProcessRoutingRules(scenario.GetRoutingRules())
				Expect(err).NotTo(HaveOccurred())

				scenario.VerifyExpectations(config)
			})
		})

		When("when a single traffic shifting rule is applied and subsequently all traffic is allowed", func() {
			BeforeEach(func() {
				scenario = scenarios.RoutingRule3()
			})

			It("produces the correct configuration object", func() {
				err = config.ProcessRoutingRules(scenario.GetRoutingRules())
				Expect(err).NotTo(HaveOccurred())
				err = config.AllowAll()
				Expect(err).NotTo(HaveOccurred())

				scenario.VerifyExpectations(config)
			})
		})

		When("two traffic shifting rules that differ only on the matched path prefix are applied", func() {
			BeforeEach(func() {
				scenario = scenarios.RoutingRule2()
			})

			It("produces the correct configuration object", func() {
				err = config.ProcessRoutingRules(scenario.GetRoutingRules())
				Expect(err).NotTo(HaveOccurred())

				scenario.VerifyExpectations(config)
			})
		})

		When("when a single traffic shifting rule matching multiple destinations is applied", func() {
			BeforeEach(func() {
				scenario = scenarios.RoutingRule4()
			})

			It("produces the correct configuration object", func() {
				err = config.ProcessRoutingRules(scenario.GetRoutingRules())
				Expect(err).NotTo(HaveOccurred())
				err = config.AllowAll()
				Expect(err).NotTo(HaveOccurred())

				scenario.VerifyExpectations(config)
			})
		})

	})
})
