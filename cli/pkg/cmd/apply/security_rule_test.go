package apply_test

import (
	"fmt"
	"strings"

	"github.com/solo-io/supergloo/cli/pkg/helpers/clients"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	skclients "github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/supergloo/cli/test/utils"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
)

var _ = Describe("SecurityRule", func() {
	srArgs := func(name string, allowedMethods, allowedPaths []string) string {
		args := fmt.Sprintf("apply securityrule --name=%v --target-mesh=my.mesh ", name)
		if len(allowedMethods) > 0 {
			args += fmt.Sprintf("--allowed-methods=%v ", strings.Join(allowedMethods, ","))
		}
		if len(allowedPaths) > 0 {
			args += fmt.Sprintf("--allowed-paths=%v ", strings.Join(allowedPaths, ","))
		}
		return args
	}

	BeforeEach(func() {
		clients.UseMemoryClients()
		_, _ = clients.MustMeshClient().Write(&v1.Mesh{Metadata: core.Metadata{Namespace: "my", Name: "mesh"}}, skclients.WriteOpts{})
	})

	getSecurityRule := func(name string) *v1.SecurityRule {
		sr, err := clients.MustSecurityRuleClient().Read("supergloo-system", name, skclients.ReadOpts{})
		ExpectWithOffset(1, err).NotTo(HaveOccurred())
		return sr
	}

	Context("no target mesh", func() {
		It("returns an error", func() {
			err := utils.Supergloo("apply securityrule --name foo ")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("target mesh must be specified, provide with --target-mesh flag"))
		})
	})
	Context("nonexistant target mesh", func() {
		It("returns an error", func() {
			err := utils.Supergloo("apply securityrule --name foo --target-mesh notmy.mesh")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("notmy.mesh does not exist"))
		})
	})
	Context("valid inputs", func() {
		selectorTest := func(methods, paths []string, extraArgs ...string) (*v1.SecurityRule, error) {
			name := "sr"

			args := srArgs(name, methods, paths) + strings.Join(extraArgs, " ")

			err := utils.Supergloo(args)
			if err != nil {
				return nil, err
			}

			securityRule := getSecurityRule(name)

			return securityRule, nil
		}
		methods := []string{"GET", "POST"}
		paths := []string{"/v1", "/v2"}
		Context("no selector", func() {
			It("creates a rule with no selectors", func() {
				securityRule, err := selectorTest(nil, nil)
				Expect(err).NotTo(HaveOccurred())

				Expect(securityRule.SourceSelector).To(BeNil())
				Expect(securityRule.DestinationSelector).To(BeNil())
			})
		})
		Context("labels selector", func() {
			It("creates a rule with label selectors", func() {
				securityRule, err := selectorTest(methods, paths, "--source-labels KEY1=VAL1",
					"--source-labels KEY2=VAL2",
					"--dest-labels KEY1=VAL1",
					"--dest-labels KEY2=VAL2")
				Expect(err).NotTo(HaveOccurred())

				expectedMap := map[string]string{"KEY1": "VAL1", "KEY2": "VAL2"}

				Expect(securityRule.SourceSelector).NotTo(BeNil())
				Expect(securityRule.SourceSelector.SelectorType).To(BeAssignableToTypeOf(&v1.PodSelector_LabelSelector_{}))
				ss := securityRule.SourceSelector.SelectorType.(*v1.PodSelector_LabelSelector_).LabelSelector
				Expect(ss.LabelsToMatch).To(Equal(expectedMap))

				Expect(securityRule.DestinationSelector).NotTo(BeNil())
				Expect(securityRule.DestinationSelector.SelectorType).To(BeAssignableToTypeOf(&v1.PodSelector_LabelSelector_{}))
				ds := securityRule.DestinationSelector.SelectorType.(*v1.PodSelector_LabelSelector_).LabelSelector
				Expect(ds.LabelsToMatch).To(Equal(expectedMap))

				Expect(securityRule.AllowedMethods).To(Equal(methods))
				Expect(securityRule.AllowedPaths).To(Equal(paths))
			})
		})
		Context("overwrite previous rule", func() {
			It("updates an existing rule with the same name", func() {
				name := "sr"

				args := srArgs(name, methods, nil)
				err := utils.Supergloo(args)
				Expect(err).NotTo(HaveOccurred())
				securityRule := getSecurityRule(name)

				Expect(securityRule.AllowedMethods).To(Equal(methods))
				Expect(securityRule.AllowedPaths).To(BeEmpty())

				args = srArgs(name, nil, paths)
				err = utils.Supergloo(args)
				Expect(err).NotTo(HaveOccurred())
				securityRule = getSecurityRule(name)

				Expect(securityRule.AllowedMethods).To(BeEmpty())
				Expect(securityRule.AllowedPaths).To(Equal(paths))
			})
		})
		Context("--dryrun flag", func() {
			It("prints the kubernetes yaml", func() {
				name := "sr"

				args := srArgs(name, methods, paths)
				args += "--source-labels KEY1=VAL1 "
				args += "--source-labels KEY2=VAL2 "
				args += "--dest-labels KEY1=VAL1 "
				args += "--dest-labels KEY2=VAL2 "
				args += " --dryrun"

				out, err := utils.SuperglooOut(args)
				Expect(err).NotTo(HaveOccurred())
				Expect(out).To(Equal(`apiVersion: supergloo.solo.io/v1
kind: SecurityRule
metadata:
  creationTimestamp: null
  name: sr
  namespace: supergloo-system
spec:
  allowedMethods:
  - GET
  - POST
  allowedPaths:
  - /v1
  - /v2
  destinationSelector:
    labelSelector:
      labelsToMatch:
        KEY1: VAL1
        KEY2: VAL2
  sourceSelector:
    labelSelector:
      labelsToMatch:
        KEY1: VAL1
        KEY2: VAL2
  targetMesh:
    name: mesh
    namespace: my
status: {}
`))
			})
		})
	})
})
