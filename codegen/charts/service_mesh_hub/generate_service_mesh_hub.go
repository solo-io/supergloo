package main

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/solo-io/service-mesh-hub/cli/pkg/cliconstants"
	"github.com/solo-io/skv2/codegen"
	"github.com/solo-io/skv2/codegen/model"
	"github.com/solo-io/skv2/contrib"
	"github.com/solo-io/solo-kit/pkg/code-generator/sk_anyvendor"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

//go:generate go run generate_service_mesh_hub.go

func main() {
	log.Println("starting generate SMH")

	var renderTypes bool
	if os.Getenv("REGENERATE_TYPES") == "" {
		log.Println("REGENERATE_TYPES is not set, skipping autopilot client gen")
	} else {
		renderTypes = true
	}

	// load custom client template
	customClientTemplateBytes, err := ioutil.ReadFile("../custom_client.gotmpl")
	customClientTemplate := string(customClientTemplateBytes)
	if err != nil {
		log.Fatal(err)
	}
	// load custom client providers template
	customClientProvidersBytes, err := ioutil.ReadFile("../custom_client_providers.gotmpl")
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
					Group:   "core." + cliconstants.ServiceMeshHubApiGroupSuffix,
					Version: "v1alpha1",
				},
				Module: "github.com/solo-io/service-mesh-hub",
				Resources: []model.Resource{
					{
						Kind: "Settings",
						Spec: model.Field{
							Type: model.Type{
								Name:      "SettingsSpec",
								GoPackage: "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types",
							},
						},
						Status: &model.Field{Type: model.Type{
							Name:      "SettingsStatus",
							GoPackage: "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types",
						}},
					},
				},
				ApiRoot:          "pkg/api/",
				RenderManifests:  true,
				RenderTypes:      renderTypes,
				RenderController: true,
				RenderProtos:     true,
				RenderClients:    true,
				CustomTemplates: []model.CustomTemplates{
					{
						Templates: map[string]string{
							"client_providers.go": customClientProviders,
						},
					},
				},
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
				RenderTypes:      renderTypes,
				RenderController: true,
				RenderProtos:     true,
				CustomTemplates: []model.CustomTemplates{
					{
						Templates: map[string]string{
							"clients.go":          customClientTemplate,
							"client_providers.go": customClientProviders,
						},
					},
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
				RenderTypes:      renderTypes,
				RenderController: true,
				RenderProtos:     true,
				CustomTemplates: []model.CustomTemplates{
					{
						Templates: map[string]string{
							"clients.go":          customClientTemplate,
							"client_providers.go": customClientProviders,
						},
					},
					contrib.Sets,
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
						Kind: "ServiceAccount",
					},
					{
						Kind: "ConfigMap",
					},
					{
						Kind: "Service",
					},
					{
						Kind: "Pod",
					},
					{
						Kind: "Namespace",
					},
					{
						Kind: "Node",
					},
				},
				RenderController: true,
				RenderClients:    true,
				CustomTemplates: []model.CustomTemplates{
					{
						Templates: map[string]string{
							"client_providers.go": customClientProviders,
						},
					},
				},
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
					{
						Kind: "ReplicaSet",
					},
				},
				RenderController: true,
				RenderClients:    true,
				CustomTemplates: []model.CustomTemplates{
					{
						Templates: map[string]string{
							"client_providers.go": customClientProviders,
						},
					},
				},
				CustomTypesImportPath: "k8s.io/api/apps/v1",
				ApiRoot:               "pkg/api/kubernetes",
			},
			{
				GroupVersion: schema.GroupVersion{
					Group:   "apiextensions.k8s.io",
					Version: "v1beta1",
				},
				Module: "k8s.io/apiextensions-apiserver",
				Resources: []model.Resource{
					{
						Kind: "CustomResourceDefinition",
					},
				},
				RenderClients: true,
				CustomTemplates: []model.CustomTemplates{
					{
						Templates: map[string]string{
							"client_providers.go": customClientProviders,
						},
					},
				},
				CustomTypesImportPath: "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1",
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
				CustomTemplates: []model.CustomTemplates{
					{
						Templates: map[string]string{
							"clients.go":          customClientTemplate,
							"client_providers.go": customClientProviders,
						},
					},
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
				CustomTemplates: []model.CustomTemplates{
					{
						Templates: map[string]string{
							"clients.go":          customClientTemplate,
							"client_providers.go": customClientProviders,
						},
					},
				},
				ApiRoot: "pkg/api/istio",
			},
			{
				GroupVersion: schema.GroupVersion{
					Group:   "",
					Version: "v1alpha2",
				},
				Module: "github.com/linkerd/linkerd2",
				Resources: []model.Resource{
					{
						Kind: "ServiceProfile",
					},
				},
				CustomTypesImportPath: "github.com/linkerd/linkerd2/controller/gen/apis/serviceprofile/v1alpha2",
				CustomTemplates: []model.CustomTemplates{
					{
						Templates: map[string]string{
							"clients.go":          customClientTemplate,
							"client_providers.go": customClientProviders,
						},
					},
				},
				ApiRoot: "pkg/api/linkerd",
			},
			{
				GroupVersion: schema.GroupVersion{
					Group:   "split",
					Version: "v1alpha1",
				},
				Module: "github.com/servicemeshinterface/smi-sdk-go",
				Resources: []model.Resource{
					{
						Kind: "TrafficSplit",
					},
				},
				CustomTypesImportPath: "github.com/servicemeshinterface/smi-sdk-go/pkg/apis/split/v1alpha1",
				CustomTemplates: []model.CustomTemplates{
					{
						Templates: map[string]string{
							"clients.go":          customClientTemplate,
							"client_providers.go": customClientProviders,
						},
					},
				},
				ApiRoot: "pkg/api/smi",
			},
		},
		AnyVendorConfig: apImports,
		ManifestRoot:    "install/helm/charts/custom-resource-definitions",
	}
	if err := skv2Cmd.Execute(); err != nil {
		log.Fatal(err)
	}

	log.Printf("Finished generating Service Mesh Hub parent chart code\n")
}
