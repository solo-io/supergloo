package main

import (
	externalapis "github.com/solo-io/external-apis/codegen"
	"github.com/solo-io/skv2/contrib"
	istionetworkingv1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"log"

	"github.com/solo-io/service-mesh-hub/codegen/groups"
	"github.com/solo-io/service-mesh-hub/codegen/templates"
	"github.com/solo-io/service-mesh-hub/pkg/common/constants"
	"github.com/solo-io/skv2/codegen"
	"github.com/solo-io/skv2/codegen/model"
	"github.com/solo-io/solo-kit/pkg/code-generator/sk_anyvendor"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

var (
	appName = "service-mesh-hub"

	discoveryInputSnapshotCodePath      = "pkg/api/discovery.smh.solo.io/snapshot/input/snapshot.go"
	discoveryReconcilerSnapshotCodePath = "pkg/api/discovery.smh.solo.io/snapshot/input/reconciler.go"
	discoveryOutputSnapshotCodePath     = "pkg/api/discovery.smh.solo.io/snapshot/output/snapshot.go"

	networkingInputSnapshotCodePath      = "pkg/api/networking.smh.solo.io/snapshot/input/snapshot.go"
	networkingReconcilerSnapshotCodePath = "pkg/api/networking.smh.solo.io/snapshot/input/reconciler.go"
	networkingOutputSnapshotCodePath     = "pkg/api/networking.smh.solo.io/snapshot/output/snapshot.go"

	smhCrdManifestRoot = "install/helm/charts/custom-resource-definitions"
	csrCrdManifestRoot = "install/helm/charts/csr-agent/"

	inputDiscoverySnapshot = map[schema.GroupVersion][]string{
		corev1.SchemeGroupVersion: {
			"Pod",
			"Service",
			"ConfigMap",
		},
		appsv1.SchemeGroupVersion: {
			"Deployment",
			"ReplicaSet",
			"DaemonSet",
			"StatefulSet",
		},
	}

	outputDiscoverySnapshot = map[schema.GroupVersion][]string{
		schema.GroupVersion{
			Group:   "discovery." + constants.ServiceMeshHubApiGroupSuffix,
			Version: "v1alpha1",
		}: {
			"Mesh",
			"MeshWorkload",
			"MeshService",
		},
	}

	inputNetworkingSnapshot = map[schema.GroupVersion][]string{
		schema.GroupVersion{
			Group:   "discovery." + constants.ServiceMeshHubApiGroupSuffix,
			Version: "v1alpha1",
		}: {
			"Mesh",
			"MeshWorkload",
			"MeshService",
		},
		schema.GroupVersion{
			Group:   "networking." + constants.ServiceMeshHubApiGroupSuffix,
			Version: "v1alpha1",
		}: {
			"TrafficPolicy",
			"AccessControlPolicy",
			"VirtualMesh",
		},
	}

	outputNetworkingIstioSnapshot = map[schema.GroupVersion][]string{
		istionetworkingv1alpha3.SchemeGroupVersion: {
			"DestinationRule",
			"VirtualService",
			"EnvoyFilter",
		},
	}

	protoImports = sk_anyvendor.CreateDefaultMatchOptions([]string{
		"api/**/*.proto",
	})
)

func run() error {
	log.Printf("generating smh")
	if err := makeSmhCommand().Execute(); err != nil {
		return err
	}
	log.Printf("generating csr-agent")
	if err := makeCsrCommand().Execute(); err != nil {
		return err
	}
	return nil
}

func makeSmhCommand() codegen.Command {

	protoImports.External["github.com/solo-io/skv2"] = []string{
		"api/**/*.proto",
	}

	topLevelTemplates := []model.CustomTemplates{
		makeDiscoveryInputSnapshotTemplate(),
		makeDiscoveryReconcilerTemplate(),
		makeDiscoveryOutputSnapshotTemplate(),
		makeNetworkingInputSnapshotTemplate(),
		makeNetworkingReconcilerTemplate(),
		makeNetworkingOutputSnapshotTemplate(),
	}

	return codegen.Command{
		AppName:           appName,
		AnyVendorConfig:   protoImports,
		ManifestRoot:      smhCrdManifestRoot,
		TopLevelTemplates: topLevelTemplates,
		Groups:            groups.SMHGroups,
	}
}

func makeCsrCommand() codegen.Command {

	return codegen.Command{
		AppName:         appName,
		AnyVendorConfig: protoImports,
		ManifestRoot:    csrCrdManifestRoot,
		Groups:          groups.CSRGroups,
	}
}

func makeDiscoveryInputSnapshotTemplate() model.CustomTemplates {
	return contrib.InputSnapshot(
		discoveryInputSnapshotCodePath,
		"github.com/solo-io/external-apis",
		externalapis.Groups,
		inputDiscoverySnapshot,
	)
}

func makeDiscoveryReconcilerTemplate() model.CustomTemplates {
	inputGroups := templates.SelectResources(
		"github.com/solo-io/external-apis",
		externalapis.Groups,
		inputDiscoverySnapshot,
	)

	return model.CustomTemplates{
		Templates: map[string]string{
			discoveryReconcilerSnapshotCodePath: templates.ReconcilerTemplateContents,
		},
		Funcs: templates.MakeSnapshotFuncs(inputGroups),
	}
}

func makeDiscoveryOutputSnapshotTemplate() model.CustomTemplates {
	outputGroups := templates.SelectResources(
		"",
		groups.SMHGroups,
		outputDiscoverySnapshot,
	)

	return model.CustomTemplates{
		Templates: map[string]string{
			discoveryOutputSnapshotCodePath: templates.OutputSnapshotTemplateContents,
		},
		Funcs: templates.MakeSnapshotFuncs(outputGroups),
	}
}

func makeNetworkingInputSnapshotTemplate() model.CustomTemplates {
	return contrib.InputSnapshot(
		networkingInputSnapshotCodePath,
		"",
		groups.SMHGroups,
		inputNetworkingSnapshot,
	)
}

func makeNetworkingReconcilerTemplate() model.CustomTemplates {
	inputGroups := templates.SelectResources(
		"",
		groups.SMHGroups,
		inputNetworkingSnapshot,
	)

	return model.CustomTemplates{
		Templates: map[string]string{
			networkingReconcilerSnapshotCodePath: templates.ReconcilerTemplateContents,
		},
		Funcs: templates.MakeSnapshotFuncs(inputGroups),
	}
}

func makeNetworkingOutputSnapshotTemplate() model.CustomTemplates {
	outputGroups := templates.SelectResources("github.com/solo-io/external-apis", externalapis.Groups, outputNetworkingIstioSnapshot)

	return model.CustomTemplates{
		Templates: map[string]string{
			networkingOutputSnapshotCodePath: templates.OutputSnapshotTemplateContents,
		},
		Funcs: templates.MakeSnapshotFuncs(outputGroups),
	}
}
