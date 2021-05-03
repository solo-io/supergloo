package common

import (
	"fmt"
	"github.com/solo-io/gloo-mesh/pkg/test/apps/context"
	"github.com/solo-io/gloo-mesh/pkg/test/manifests"
	"istio.io/istio/pkg/test/framework/components/istio"
	"istio.io/istio/pkg/test/framework/resource"
	"istio.io/istio/pkg/test/scopes"
	"istio.io/pkg/log"
	"testing"
	"time"
)

func IstioSetupFunc(operatorFile string) func(ctx resource.Context, cfg *istio.Config) {
	return func(ctx resource.Context, cfg *istio.Config) {
		// load custom Istio control plane value
		istioGatewayConfig, err := manifests.RenderOperator(operatorFile, nil)
		if err != nil {
			scopes.Framework.Errorf("error rendering istio operator configuration file %s %w", operatorFile, err)
		}
		cfg.ControlPlaneValues = istioGatewayConfig
		cfg.DeployEastWestGW = false
	}
}

// TestGroup is a group of tests all using the same deployment. e.g. multiple tests all using the same httpbin k8s instance
type TestGroup struct {
	// Name of the tests group
	Name string
	// Cases are a list of tests cases to run with the deployed application
	Cases []TestCase
}

// TestCase is a single tests, generally used inside a TestGroup
type TestCase struct {
	// Name of the tests
	Name string
	// Namespace to apply configuration
	Namespace string
	// Description of the Test
	Description string
	// Test is the tests function. The output from the TestGroup Deploy call is passed to the tests
	Test func(ctx resource.Context, t *testing.T, deploymentCtx *context.DeploymentContext)
	// Skip configures the tests to be skipped for the configured reason
	Skip string
	// Test file name to deploy to be fetched from box
	FileName string
}

// Run runs the tests group by first by deploying the group application, then running through each tests case.
func (g TestGroup) Run(ctx resource.Context, t *testing.T, deploymentContext *context.DeploymentContext) {
	var skipNextTests error

	for _, test := range g.Cases {
		name := g.Name + "_" + test.Name
		t.Run(name, func(t *testing.T) {
			// Do not even attempt to create the scenario if we're gonna skip the tests
			if test.Skip != "" {
				t.Skip(test.Skip)
			}

			// If failures from previous tests prevent this one from running properly, skip it
			if skipNextTests != nil {
				t.Skipf("skipping due to previous errors: %v", skipNextTests)
			}

			configYAMLStr, err := manifests.RenderTestFile(test.FileName, deploymentContext)
			if err != nil {
				t.Fatalf("failed to render tests file: %v", err)
			}

			t.Cleanup(func() {
				if err = ctx.Config().DeleteYAML("", configYAMLStr); err != nil {
					// Don't try to continue in this case as leftover config could invalidate future tests
					skipNextTests = fmt.Errorf("%s: failed to cleanup objects: %v", name, err)
					t.Fatal(skipNextTests.Error())
				}
				// Wait for config deletion to be propagated
				time.Sleep(time.Second * 10)
			})

			if err = ctx.Config(ctx.Clusters().Default()).ApplyYAML("", configYAMLStr); err != nil {
				t.Fatalf("failed to apply config: %v", err)
			}

			// Wait for config to be applied / previous config (if any) to be deleted
			log.Info("Waiting 10 seconds for config to be applied")
			time.Sleep(10 * time.Second)

			test.Test(ctx, t, deploymentContext)
		})

	}
}
