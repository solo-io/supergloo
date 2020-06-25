package main

import (
	externalapis "github.com/solo-io/external-apis/codegen"
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

var (
	appName                         = "service-mesh-hub"
	discoveryInputSnapshotCodePath = "pkg/api/discovery.smh.solo.io/snapshot/input/snapshot.go"
	discoveryOutputSnapshotCodePath = "pkg/api/discovery.smh.solo.io/snapshot/output/snapshot.go"
	smhCrdManifestRoot              = "install/helm/charts/custom-resource-definitions"
	csrCrdManifestRoot              = "install/helm/charts/csr-agent/"

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

	protoImports = sk_anyvendor.CreateDefaultMatchOptions([]string{
		"api/**/*.proto",
	})
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

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

	topLevelTemplates := []model.CustomTemplates{
		makeDiscoveryInputSnapshotTemplate(),
		makeDiscoveryOutputSnapshotTemplate(),
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
	inputGroups := templates.SelectResources("github.com/solo-io/external-apis", externalapis.Groups, inputDiscoverySnapshot)

	return model.CustomTemplates{
		Templates: map[string]string{
			discoveryInputSnapshotCodePath: templates.InputSnapshotTemplateContents,
		},
		Funcs: templates.MakeSnapshotFuncs(inputGroups),
	}
}

func makeDiscoveryOutputSnapshotTemplate() model.CustomTemplates {
	outputGroups := templates.SelectResources("", groups.SMHGroups, outputDiscoverySnapshot)

	return model.CustomTemplates{
		Templates: map[string]string{
			discoveryOutputSnapshotCodePath: templates.OutputSnapshotTemplateContents,
		},
		Funcs: templates.MakeSnapshotFuncs(outputGroups),
	}
}
