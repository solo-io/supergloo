package main

import (
	"log"

	"github.com/solo-io/autopilot/codegen"
	"github.com/solo-io/autopilot/codegen/model"
	"github.com/solo-io/solo-kit/pkg/code-generator/cmd"
	"github.com/solo-io/solo-kit/pkg/code-generator/docgen/options"
	"github.com/solo-io/solo-kit/pkg/code-generator/sk_anyvendor"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

//go:generate go run generate.go
//go:generate bash ./api/external/smi/generate.sh
//go:generate mockgen -package mock_manager -destination ./test/mocks/manager/mock_manager.go sigs.k8s.io/controller-runtime/pkg/manager Manager
//go:generate mockgen -package mock_manager -destination ./test/mocks/manager/mock_cache.go sigs.k8s.io/controller-runtime/pkg/cache Cache

const (
	GlooPkg = "github.com/solo-io/gloo"
)

func main() {

	imports := sk_anyvendor.CreateDefaultMatchOptions(
		[]string{
			"api/**/*.proto",
			sk_anyvendor.SoloKitMatchPattern,
		},
	)
	imports.External[GlooPkg] = []string{"projects/**/*.proto"}
	imports.External["istio.io/api"] = []string{
		"networking/v1alpha3/*.proto", "security/v1beta1/*.proto", "type/v1beta1/*.proto",
	}

	log.Printf("starting generate")
	docsOpts := &cmd.DocsOptions{
		Output: options.Hugo,
	}
	if err := cmd.Generate(cmd.GenerateOptions{
		RelativeRoot:       ".",
		CompileProtos:      true,
		GenDocs:            docsOpts,
		SkipGeneratedTests: true,
		SkipGenMocks:       true,
		ExternalImports:    imports,
	}); err != nil {
		log.Fatalf("generate failed!: %v", err)
	}

	apImports := sk_anyvendor.CreateDefaultMatchOptions([]string{
		"api/config/v1alpha1/*.proto",
		"api/core/v1alpha1/*.proto",
	})
	apImports.External["istio.io/api"] = []string{
		"networking/v1alpha3/*.proto", "security/v1beta1/*.proto", "type/v1beta1/*.proto",
	}
	autopilotCmd := codegen.Command{
		AppName: "service-mesh-hub",
		Groups: []model.Group{
			{
				GroupVersion: schema.GroupVersion{
					Group:   "config.zephyr.solo.io",
					Version: "v1alpha1",
				},
				Module: "github.com/solo-io/mesh-projects",
				Resources: []model.Resource{
					{
						Kind:   "RoutingRule",
						Spec:   model.Field{Type: "RoutingRuleSpec"},
						Status: &model.Field{Type: "RoutingRuleStatus"},
					},
					{
						Kind:   "SecurityRule",
						Spec:   model.Field{Type: "SecurityRuleSpec"},
						Status: &model.Field{Type: "SecurityRuleStatus"},
					},
				},
				RenderManifests:  true,
				RenderTypes:      true,
				RenderController: true,
				RenderProtos:     true,
				ApiRoot:          "pkg/api",
			},
			{
				GroupVersion: schema.GroupVersion{
					Group:   "core.zephyr.solo.io",
					Version: "v1alpha1",
				},
				Module: "github.com/solo-io/mesh-projects",
				Resources: []model.Resource{
					{
						Kind: "KubernetesCluster",
						Spec: model.Field{Type: "KubernetesClusterSpec"},
					},
					{
						Kind: "Service",
						Spec: model.Field{Type: "ServiceSpec"},
					},
					{
						Kind: "Mesh",
						Spec: model.Field{Type: "MeshSpec"},
					},
					{
						Kind: "MeshGroup",
						Spec: model.Field{Type: "MeshGroupSpec"},
					},
				},
				RenderManifests:  true,
				RenderTypes:      true,
				RenderController: true,
				RenderProtos:     true,
				ApiRoot:          "pkg/api",
			},
			{
				GroupVersion: schema.GroupVersion{
					Group:   "core",
					Version: "v1",
				},
				Module: "k8s.io/api",
				Resources: []model.Resource{
					{
						Kind: "Secret",
					},
					{
						Kind: "Service",
					},
					{
							Kind: "Pod",
					},
				},
				RenderController:      true,
				CustomTypesImportPath: "k8s.io/api/core/v1",
				ApiRoot:               "services/common/cluster",
			},
		},
		AnyVendorConfig: apImports,
		ManifestRoot:    "hack",
	}
	if err := autopilotCmd.Execute(); err != nil {
		log.Fatal(err)
	}

	log.Printf("Finished generating code")
}
