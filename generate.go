package main

import (
	"github.com/solo-io/service-mesh-hub/codegen/groups"
	"log"

	externalcodegen "github.com/solo-io/external-apis/codegen"
	"github.com/solo-io/service-mesh-hub/codegen/templates"
	"github.com/solo-io/service-mesh-hub/pkg/common/constants"
	"github.com/solo-io/skv2/codegen"
	"github.com/solo-io/skv2/codegen/model"
	"github.com/solo-io/solo-kit/pkg/code-generator/sk_anyvendor"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	appName                         = "service-mesh-hub"
	v1alpha1Version                 = "v1alpha1"
	discoveryOutputSnapshotCodePath = "pkg/mesh-discovery/snapshots/output"
	smhCrdManifestRoot              = "install/helm/charts/custom-resource-definitions"
	csrCrdManifestRoot              = "install/helm/charts/csr-agent/"

	outputDiscoverySnapshot = map[schema.GroupVersion][]string{
		schema.GroupVersion{
			Group:   "discovery." + constants.ServiceMeshHubApiGroupSuffix,
			Version: "v1alpha1",
		}: {"Mesh", "MeshWorkload", "MeshService"},
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
		makeDiscoverySnapshotTemplate(groups.SMHGroups),
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

func makeDiscoverySnapshotTemplate(groups []model.Group) model.CustomTemplates {
	groups = templates.SelectResources(externalcodegen.K8sGroups(), outputDiscoverySnapshot)

	return model.CustomTemplates{
		Templates: map[string]string{
			discoveryOutputSnapshotCodePath: templates.OutputSnapshotTemplateContents,
		},
		Funcs: templates.MakeSnapshotFuncs(groups),
	}
}
