package status_test

import (
	"bytes"
	"fmt"
	"io/ioutil"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/solo-io/service-mesh-hub/cli/pkg/tree/check/healthcheck"
	healthcheck_types "github.com/solo-io/service-mesh-hub/cli/pkg/tree/check/healthcheck/types"
	"github.com/solo-io/service-mesh-hub/cli/pkg/tree/check/status"
	"github.com/solo-io/service-mesh-hub/pkg/common/container-runtime/version"
)

var _ = Describe("Status pretty printer", func() {
	var (
		runTestCase = func(goldenFileName string, statusReport *status.StatusReport) {
			fileContents, err := ioutil.ReadFile("./test_golden/" + goldenFileName)
			Expect(err).NotTo(HaveOccurred(), "Could not read golden "+goldenFileName)

			output := &bytes.Buffer{}
			status.NewPrettyPrinter().Print(output, statusReport)

			dmp := diffmatchpatch.New()
			diff := dmp.DiffMain(string(fileContents), output.String(), false)
			if len(diff) > 1 || diff[0].Type != diffmatchpatch.DiffEqual {
				//fmt.Printf("%+v", diff)
				fmt.Println(dmp.DiffPrettyText(diff))
				fmt.Println("========== DETAILED DIFF ==========")
				fmt.Printf("%+v\n", diff)
				Fail("Found diffs, see diff output above")
			}
		}
	)

	DescribeTable("printer", runTestCase,
		Entry("reports success summaries correctly", "success-summary.txt", &status.StatusReport{
			Success: true,
		}),
		Entry("reports failure summaries correctly", "failed-summary.txt", &status.StatusReport{
			Success: false,
		}),
		Entry("can report a mix of success and failure, including hints and docs", "federation-failure.txt", &status.StatusReport{
			Success: false,
			Results: map[healthcheck_types.Category][]*status.HealthCheckResult{
				healthcheck.KubernetesAPI: {
					{
						Success:     true,
						Description: "Kubernetes API server is reachable",
					},
					{
						Success:     true,
						Description: fmt.Sprintf("running the minimum supported Kubernetes version (required: >=1.%d)", version.MinimumSupportedKubernetesMinorVersion),
					},
				},
				healthcheck.ManagementPlane: {
					{
						Success:     true,
						Description: "installation namespace exists",
					},
					{
						Success:     true,
						Description: "components are running",
					},
				},
				healthcheck.ServiceFederation: {
					{
						Success:     false,
						Description: "federation decisions have been written to MeshServices",
						Message:     "failed to write federation metadata to mesh service 'test-mesh-service.service-mesh-hub'; status is 'INVALID'",
						Hint:        "get details from the failing MeshService: `kubectl -n service-mesh-hub get meshservice test-mesh-service -oyaml`",
						DocsLink:    "https://docs.solo.io/test-link",
					},
				},
			},
		}),
	)
})
