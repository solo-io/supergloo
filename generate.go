package main

import (
	"context"
	"log"
	"os"

	"github.com/solo-io/autopilot/codegen"
	"github.com/solo-io/autopilot/codegen/model"
	"github.com/solo-io/service-mesh-hub/cli/pkg/cliconstants"
	"github.com/solo-io/service-mesh-hub/cli/pkg/wire"
	docgen "github.com/solo-io/service-mesh-hub/docs"
	"github.com/solo-io/solo-kit/pkg/code-generator/sk_anyvendor"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

//go:generate go run generate.go
//go:generate mockgen -package mock_controller_runtime -destination ./test/mocks/controller-runtime/mock_manager.go sigs.k8s.io/controller-runtime/pkg/manager Manager
//go:generate mockgen -package mock_controller_runtime -destination ./test/mocks/controller-runtime/mock_predicate.go sigs.k8s.io/controller-runtime/pkg/predicate Predicate
//go:generate mockgen -package mock_controller_runtime -destination ./test/mocks/controller-runtime/mock_cache.go sigs.k8s.io/controller-runtime/pkg/cache Cache
//go:generate mockgen -package mock_controller_runtime -destination ./test/mocks/controller-runtime/mock_dynamic_client.go  sigs.k8s.io/controller-runtime/pkg/client Client,StatusWriter
//go:generate mockgen -package mock_cli_runtime -destination ./test/mocks/cli_runtime/mock_rest_client_getter.go k8s.io/cli-runtime/pkg/resource RESTClientGetter
//go:generate mockgen -package mock_corev1 -destination ./test/mocks/corev1/mock_service_controller.go github.com/solo-io/service-mesh-hub/services/common/cluster/core/v1/controller ServiceController
//go:generate mockgen -package mock_zephyr_discovery -destination ./test/mocks/zephyr/discovery/mock_mesh_workload_controller.go github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/controller MeshWorkloadController,MeshServiceController
//go:generate mockgen -package mock_zephyr_networking -destination ./test/mocks/zephyr/networking/mock_virtual_mesh_controller.go github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1/controller VirtualMeshController,TrafficPolicyController,AccessControlPolicyController

func main() {
	log.Println("starting generate")

	apImports := sk_anyvendor.CreateDefaultMatchOptions([]string{
		"api/**/*.proto",
	})
	autopilotCmd := codegen.Command{
		AppName: "service-mesh-hub",
		Groups: []model.Group{
			{
				GroupVersion: schema.GroupVersion{
					Group:   "security." + cliconstants.ServiceMeshHubApiGroupSuffix,
					Version: "v1alpha1",
				},
				Module: "github.com/solo-io/service-mesh-hub",
				Resources: []model.Resource{
					{
						Kind: "VirtualMeshCertificateSigningRequest",
						Spec: model.Field{
							Type: model.Type{
								Name:      "VirtualMeshCertificateSigningRequestSpec",
								GoPackage: "github.com/solo-io/service-mesh-hub/pkg/api/security.zephyr.solo.io/v1alpha1/types",
							},
						},
						Status: &model.Field{Type: model.Type{
							Name:      "VirtualMeshCertificateSigningRequestStatus",
							GoPackage: "github.com/solo-io/service-mesh-hub/pkg/api/security.zephyr.solo.io/v1alpha1/types",
						}},
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
					Group:   "networking." + cliconstants.ServiceMeshHubApiGroupSuffix,
					Version: "v1alpha1",
				},
				Module: "github.com/solo-io/service-mesh-hub",
				Resources: []model.Resource{
					{
						Kind: "TrafficPolicy",
						Spec: model.Field{
							Type: model.Type{
								Name:      "TrafficPolicySpec",
								GoPackage: "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1/types",
							},
						},
						Status: &model.Field{Type: model.Type{
							Name:      "TrafficPolicyStatus",
							GoPackage: "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1/types",
						}},
					},
					{
						Kind: "AccessControlPolicy",
						Spec: model.Field{
							Type: model.Type{
								Name:      "AccessControlPolicySpec",
								GoPackage: "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1/types",
							},
						},
						Status: &model.Field{Type: model.Type{
							Name:      "AccessControlPolicyStatus",
							GoPackage: "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1/types",
						}},
					},
					{
						Kind: "VirtualMesh",
						Spec: model.Field{
							Type: model.Type{
								Name:      "VirtualMeshSpec",
								GoPackage: "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1/types",
							},
						},
						Status: &model.Field{Type: model.Type{
							Name:      "VirtualMeshStatus",
							GoPackage: "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1/types",
						}},
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
					Group:   "discovery." + cliconstants.ServiceMeshHubApiGroupSuffix,
					Version: "v1alpha1",
				},
				Module: "github.com/solo-io/service-mesh-hub",
				Resources: []model.Resource{
					{
						Kind: "KubernetesCluster",
						Spec: model.Field{
							Type: model.Type{
								Name:      "KubernetesClusterSpec",
								GoPackage: "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/types",
							},
						},
					},
					{
						Kind: "MeshService",
						Spec: model.Field{
							Type: model.Type{
								Name:      "MeshServiceSpec",
								GoPackage: "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/types",
							},
						},
						Status: &model.Field{Type: model.Type{
							Name:      "MeshServiceStatus",
							GoPackage: "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/types",
						}},
					},
					{
						Kind: "MeshWorkload",
						Spec: model.Field{
							Type: model.Type{
								Name:      "MeshWorkloadSpec",
								GoPackage: "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/types",
							},
						},
						Status: &model.Field{Type: model.Type{
							Name:      "MeshWorkloadStatus",
							GoPackage: "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/types",
						}},
					},
					{
						Kind: "Mesh",
						Spec: model.Field{
							Type: model.Type{
								Name:      "MeshSpec",
								GoPackage: "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/types",
							},
						},
						Status: &model.Field{Type: model.Type{
							Name:      "MeshStatus",
							GoPackage: "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/types",
						}},
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
			{
				GroupVersion: schema.GroupVersion{
					Group:   "apps",
					Version: "v1",
				},
				Module: "k8s.io/api",
				Resources: []model.Resource{
					{
						Kind: "Deployment",
					},
				},
				RenderController:      true,
				CustomTypesImportPath: "k8s.io/api/apps/v1",
				ApiRoot:               "services/common/cluster",
			},
		},
		AnyVendorConfig: apImports,
		ManifestRoot:    "install/helm/charts/custom-resource-definitions",
	}
	if err := autopilotCmd.Execute(); err != nil {
		log.Fatal(err)
	}

	log.Printf("Finished generating code\n")
	log.Printf("Started docs generation\n")

	// generate docs
	rootCmd := wire.InitializeCLI(context.TODO(), os.Stdin, os.Stdout)
	docsGen := docgen.Options{
		Proto: docgen.ProtoOptions{
			OutputDir: "content/reference/api",
		},
		Cli: docgen.CliOptions{
			RootCmd:   rootCmd,
			OutputDir: "content/reference/cli",
		},
		DocsRoot: "docs",
	}

	if err := docgen.Execute(docsGen); err != nil {
		log.Fatal(err)
	}

	log.Printf("Finished generating docs\n")
}
