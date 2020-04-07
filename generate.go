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

	var renderClients bool
	if os.Getenv("REGENERATE_CLIENTS") == "" {
		log.Println("REGENERATE_CLIENTS is not set, skipping autopilot client gen")
	} else {
		renderClients = true
	}

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
						Kind:                 "VirtualMeshCertificateSigningRequest",
						RelativePathFromRoot: "pkg/api/security.zephyr.solo.io/v1alpha1/types",
						Spec:                 model.Field{Type: "VirtualMeshCertificateSigningRequestSpec"},
						Status:               &model.Field{Type: "VirtualMeshCertificateSigningRequestStatus"},
					},
				},
				RenderManifests:  true,
				RenderClients:    renderClients,
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
						Kind:                 "TrafficPolicy",
						RelativePathFromRoot: "pkg/api/networking.zephyr.solo.io/v1alpha1/types",
						Spec:                 model.Field{Type: "TrafficPolicySpec"},
						Status:               &model.Field{Type: "TrafficPolicyStatus"},
					},
					{
						Kind:                 "AccessControlPolicy",
						RelativePathFromRoot: "pkg/api/networking.zephyr.solo.io/v1alpha1/types",
						Spec:                 model.Field{Type: "AccessControlPolicySpec"},
						Status:               &model.Field{Type: "AccessControlPolicyStatus"},
					},
					{
						Kind:                 "VirtualMesh",
						RelativePathFromRoot: "pkg/api/networking.zephyr.solo.io/v1alpha1/types",
						Spec:                 model.Field{Type: "VirtualMeshSpec"},
						Status:               &model.Field{Type: "VirtualMeshStatus"},
					},
				},
				RenderManifests:  true,
				RenderClients:    renderClients,
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
						Kind:                 "KubernetesCluster",
						RelativePathFromRoot: "pkg/api/discovery.zephyr.solo.io/v1alpha1/types",
						Spec:                 model.Field{Type: "KubernetesClusterSpec"},
					},
					{
						Kind:                 "MeshService",
						RelativePathFromRoot: "pkg/api/discovery.zephyr.solo.io/v1alpha1/types",
						Spec:                 model.Field{Type: "MeshServiceSpec"},
						Status:               &model.Field{Type: "MeshServiceStatus"},
					},
					{
						Kind:                 "MeshWorkload",
						RelativePathFromRoot: "pkg/api/discovery.zephyr.solo.io/v1alpha1/types",
						Spec:                 model.Field{Type: "MeshWorkloadSpec"},
						Status:               &model.Field{Type: "MeshWorkloadStatus"},
					},
					{
						Kind:                 "Mesh",
						RelativePathFromRoot: "pkg/api/discovery.zephyr.solo.io/v1alpha1/types",
						Spec:                 model.Field{Type: "MeshSpec"},
						Status:               &model.Field{Type: "MeshStatus"},
					},
				},
				RenderManifests:  true,
				RenderTypes:      true,
				RenderController: true,
				RenderClients:    renderClients,
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
