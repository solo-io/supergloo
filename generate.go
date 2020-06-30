package main

import (
	"log"

	externalapis "github.com/solo-io/external-apis/codegen"
	"github.com/solo-io/skv2/contrib"
	istionetworkingv1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	"github.com/solo-io/service-mesh-hub/codegen/groups"
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

	discoveryInputSnapshotCodePath  = "pkg/api/discovery.smh.solo.io/snapshot/input/snapshot.go"
	discoveryReconcilerCodePath     = "pkg/api/discovery.smh.solo.io/snapshot/input/reconciler.go"
	discoveryOutputSnapshotCodePath = "pkg/api/discovery.smh.solo.io/snapshot/output/snapshot.go"

	networkingInputSnapshotCodePath      = "pkg/api/networking.smh.solo.io/snapshot/input/snapshot.go"
	networkingReconcilerSnapshotCodePath = "pkg/api/networking.smh.solo.io/snapshot/input/reconciler.go"
	networkingOutputSnapshotCodePath     = "pkg/api/networking.smh.solo.io/snapshot/output/snapshot.go"

	smhCrdManifestRoot = "install/helm/charts/custom-resource-definitions"
	csrCrdManifestRoot = "install/helm/charts/csr-agent/"

	discoveryInputTypes = map[schema.GroupVersion][]string{
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

	discoveryOutputTypes = map[schema.GroupVersion][]string{
		schema.GroupVersion{
			Group:   "discovery." + constants.ServiceMeshHubApiGroupSuffix,
			Version: "v1alpha1",
		}: {
			"Mesh",
			"MeshWorkload",
			"MeshService",
		},
	}

	networkingInputTypes = map[schema.GroupVersion][]string{
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
			"AccessPolicy",
			"VirtualMesh",
		},
	}

	networkingOutputTypes = map[schema.GroupVersion][]string{
		istionetworkingv1alpha3.SchemeGroupVersion: {
			"DestinationRule",
			"VirtualService",
			"EnvoyFilter",
		},
	}

	allApiGroups = map[string][]model.Group{
		"":                                 groups.SMHGroups,
		"github.com/solo-io/external-apis": externalapis.Groups,
	}

	// define custom templates
	discoveryInputSnapshot = makeTopLevelTemplate(
		contrib.InputSnapshot,
		discoveryInputSnapshotCodePath,
		discoveryInputTypes,
	)

	discoveryReconciler = makeTopLevelTemplate(
		contrib.InputReconciler,
		discoveryReconcilerCodePath,
		discoveryInputTypes,
	)

	discoveryOutputSnapshot = makeTopLevelTemplate(
		contrib.OutputSnapshot,
		discoveryOutputSnapshotCodePath,
		discoveryOutputTypes,
	)

	networkingInputSnapshot = makeTopLevelTemplate(
		contrib.InputSnapshot,
		networkingInputSnapshotCodePath,
		networkingInputTypes,
	)

	networkingReconciler = makeTopLevelTemplate(
		contrib.InputReconciler,
		networkingReconcilerSnapshotCodePath,
		networkingInputTypes,
	)

	networkingOutputSnapshot = makeTopLevelTemplate(
		contrib.OutputSnapshot,
		networkingOutputSnapshotCodePath,
		networkingOutputTypes,
	)

	topLevelTemplates = []model.CustomTemplates{
		discoveryInputSnapshot,
		discoveryReconciler,
		discoveryOutputSnapshot,
		networkingInputSnapshot,
		networkingReconciler,
		networkingOutputSnapshot,
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

	return codegen.Command{
		AppName:           appName,
		AnyVendorConfig:   protoImports,
		ManifestRoot:      smhCrdManifestRoot,
		TopLevelTemplates: topLevelTemplates,
		Groups:            groups.SMHGroups,
		RenderProtos:      true,
	}
}

func makeCsrCommand() codegen.Command {
	return codegen.Command{
		AppName:         appName,
		AnyVendorConfig: protoImports,
		ManifestRoot:    csrCrdManifestRoot,
		Groups:          groups.CSRGroups,
		RenderProtos:    true,
	}
}

func makeTopLevelTemplate(templateFunc func(params contrib.CrossGroupTemplateParameters) model.CustomTemplates, outPath string, resourceSnapshot map[schema.GroupVersion][]string) model.CustomTemplates {
	return templateFunc(contrib.CrossGroupTemplateParameters{
		OutputFilename:    outPath,
		SelectFromGroups:  allApiGroups,
		ResourcesToSelect: resourceSnapshot,
	})
}
