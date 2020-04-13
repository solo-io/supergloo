package main

import (
	"context"
	"io/ioutil"
	"log"
	"os"

	"github.com/solo-io/service-mesh-hub/cli/pkg/cliconstants"
	"github.com/solo-io/service-mesh-hub/cli/pkg/wire"
	docgen "github.com/solo-io/service-mesh-hub/docs"
	"github.com/solo-io/skv2/codegen"
	"github.com/solo-io/skv2/codegen/model"
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

	// load custom client template
	customClientTemplateBytes, err := ioutil.ReadFile("custom_client.gotmpl")
	customClientTemplate := string(customClientTemplateBytes)
	if err != nil {
		log.Fatal(err)
	}
	// load custom client providers template
	customClientProvidersBytes, err := ioutil.ReadFile("custom_client_providers.gotmpl")
	customClientProviders := string(customClientProvidersBytes)
	if err != nil {
		log.Fatal(err)
	}

	apImports := sk_anyvendor.CreateDefaultMatchOptions([]string{
		"api/**/*.proto",
	})
	skv2Cmd := codegen.Command{
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
				CustomTemplates: map[string]string{
					"clients.go":          customClientTemplate,
					"client_providers.go": customClientProviders,
				},
				ApiRoot: "pkg/api",
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
				CustomTemplates: map[string]string{
					"clients.go":          customClientTemplate,
					"client_providers.go": customClientProviders,
				},
				ApiRoot: "pkg/api",
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
				CustomTemplates: map[string]string{
					"clients.go":          customClientTemplate,
					"client_providers.go": customClientProviders,
				},
				ApiRoot: "pkg/api",
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
				ApiRoot:               "pkg/api/kubernetes",
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
				ApiRoot:               "pkg/api/kubernetes",
			},
			{
				GroupVersion: schema.GroupVersion{
					Group:   "networking",
					Version: "v1alpha3",
				},
				Module: "istio.io/client-go/pkg/apis",
				Resources: []model.Resource{
					{
						Kind: "DestinationRule",
					},
					{
						Kind: "EnvoyFilter",
					},
					{
						Kind: "Gateway",
					},
					{
						Kind: "ServiceEntry",
					},
					{
						Kind: "VirtualService",
					},
				},
				CustomTypesImportPath: "istio.io/client-go/pkg/apis/networking/v1alpha3",
				CustomTemplates: map[string]string{
					"clients.go":          customClientTemplate,
					"client_providers.go": customClientProviders,
				},
				ApiRoot: "pkg/api/istio",
			},
			{
				GroupVersion: schema.GroupVersion{
					Group:   "security",
					Version: "v1beta1",
				},
				Module: "istio.io/client-go/pkg/apis",
				Resources: []model.Resource{
					{
						Kind: "AuthorizationPolicy",
					},
				},
				CustomTypesImportPath: "istio.io/client-go/pkg/apis/security/v1beta1",
				CustomTemplates: map[string]string{
					"clients.go":          customClientTemplate,
					"client_providers.go": customClientProviders,
				},
				ApiRoot: "pkg/api/istio",
			},
		},
		AnyVendorConfig: apImports,
		ManifestRoot:    "install/helm/charts/custom-resource-definitions",
	}
	if err := skv2Cmd.Execute(); err != nil {
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
