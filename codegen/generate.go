package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	"github.com/solo-io/service-mesh-hub/pkg/common/version"

	externalapis "github.com/solo-io/external-apis/codegen"
	"github.com/solo-io/service-mesh-hub/codegen/groups"
	"github.com/solo-io/service-mesh-hub/codegen/helm"
	"github.com/solo-io/service-mesh-hub/codegen/io"
	skv1alpha1 "github.com/solo-io/skv2/api/multicluster/v1alpha1"
	"github.com/solo-io/skv2/codegen"
	"github.com/solo-io/skv2/codegen/model"
	"github.com/solo-io/skv2/contrib"
	"github.com/solo-io/solo-kit/pkg/code-generator/sk_anyvendor"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

// generates an input snapshot, input reconciler, and output snapshot for each
// top-level component. top-level components are defined
// by mapping a given set of inputs to outputs.
type topLevelComponent struct {
	// path where the generated top-level component's code will be placed.
	// input snapshots will live in <generatedCodeRoot>/input/snapshot.go
	// input reconcilers will live in <generatedCodeRoot>/input/reconciler.go
	// output snapshots will live in <generatedCodeRoot>/output/snapshot.go
	generatedCodeRoot string

	// the set of input resources for which to generate a snapshot and reconciler
	inputResources io.Snapshot

	// the set of output resources for which to generate a snapshot
	outputResources []io.OutputSnapshot
}

func (t topLevelComponent) makeCodegenTemplates() []model.CustomTemplates {
	var topLevelTemplates []model.CustomTemplates

	if len(t.inputResources) > 0 {
		topLevelTemplates = append(topLevelTemplates, makeTopLevelTemplate(
			contrib.InputSnapshot,
			t.generatedCodeRoot+"/input/snapshot.go",
			t.inputResources,
		))

		topLevelTemplates = append(topLevelTemplates, makeTopLevelTemplate(
			contrib.InputReconciler,
			t.generatedCodeRoot+"/input/reconciler.go",
			t.inputResources,
		))

		topLevelTemplates = append(topLevelTemplates, makeTopLevelTemplate(
			contrib.InputSnapshotManualBuilder,
			t.generatedCodeRoot+"/input/snapshot_manual_builder.go",
			t.inputResources,
		))

	}

	for _, outputResources := range t.outputResources {
		filePath := filepath.Join(t.generatedCodeRoot, "output", outputResources.Name, "snapshot.go")
		topLevelTemplates = append(topLevelTemplates, makeTopLevelTemplate(
			contrib.OutputSnapshot,
			filePath,
			outputResources.Snapshot,
		))
	}

	return topLevelTemplates
}

var (
	appName = "service-mesh-hub"

	topLevelComponents = []topLevelComponent{
		// discovery component
		{
			generatedCodeRoot: "pkg/api/discovery.smh.solo.io",
			inputResources:    io.DiscoveryInputTypes,
			outputResources:   []io.OutputSnapshot{io.DiscoveryOutputTypes},
		},
		// networking snapshot
		{
			generatedCodeRoot: "pkg/api/networking.smh.solo.io",
			inputResources:    io.NetworkingInputTypes,
			outputResources: []io.OutputSnapshot{
				io.IstioNetworkingOutputTypes,
				io.SmiNetworkingOutputTypes,
				io.LocalNetworkingOutputTypes,
			},
		},
		// certificate issuer component
		{
			generatedCodeRoot: "pkg/api/certificates.smh.solo.io/issuer",
			inputResources:    io.CertificateIssuerInputTypes,
		},
		// certificate agent component
		{
			generatedCodeRoot: "pkg/api/certificates.smh.solo.io/agent",
			inputResources:    io.CertificateAgentInputTypes,
			outputResources:   []io.OutputSnapshot{io.CertificateAgentOutputTypes},
		},
	}

	smhManifestRoot       = "install/helm/service-mesh-hub"
	certAgentManifestRoot = "install/helm/cert-agent/"

	vendoredMultiClusterCRDs = "vendor_any/github.com/solo-io/skv2/crds/multicluster.solo.io_v1alpha1_crds.yaml"
	importedMultiClusterCRDs = smhManifestRoot + "/crds/multicluster.solo.io_v1alpha1_crds.yaml"

	allApiGroups = map[string][]model.Group{
		"":                                 append(groups.SMHGroups, groups.CertAgentGroups...),
		"github.com/solo-io/external-apis": externalapis.Groups,
		"github.com/solo-io/skv2":          {skv1alpha1.Group},
	}

	topLevelTemplates = func() []model.CustomTemplates {
		var allTemplates []model.CustomTemplates
		for _, component := range topLevelComponents {
			allTemplates = append(allTemplates, component.makeCodegenTemplates()...)
		}
		return allTemplates
	}()

	anyvendorImports = sk_anyvendor.CreateDefaultMatchOptions([]string{
		"api/**/*.proto",
	})
)

func run() error {
	log.Printf("generating service mesh hub code with version %v", version.Version)
	chartOnly := flag.Bool("chart", false, "only generate the helm chart")
	flag.Parse()

	if err := makeSmhCommand(*chartOnly).Execute(); err != nil {
		return err
	}

	if err := makeCertAgentCommand(*chartOnly).Execute(); err != nil {
		return err
	}

	if *chartOnly {
		return nil
	}

	// TODO(ilackarms): we copy skv2 CRDs out of vendor_any into our helm chart.
	// we should consider using skv2 to automate this step for us
	if err := os.Rename(vendoredMultiClusterCRDs, importedMultiClusterCRDs); err != nil {
		return err
	}

	return nil
}

func makeSmhCommand(chartOnly bool) codegen.Command {

	anyvendorImports.External["github.com/solo-io/skv2"] = []string{
		"api/**/*.proto",
		"crds/multicluster.solo.io_v1alpha1_crds.yaml",
	}

	if chartOnly {
		return codegen.Command{
			AppName:      appName,
			ManifestRoot: smhManifestRoot,
			Chart:        helm.Chart,
		}
	}

	return codegen.Command{
		AppName:           appName,
		AnyVendorConfig:   anyvendorImports,
		ManifestRoot:      smhManifestRoot,
		TopLevelTemplates: topLevelTemplates,
		Groups:            groups.SMHGroups,
		RenderProtos:      true,
		Chart:             helm.Chart,
	}
}

func makeCertAgentCommand(chartOnly bool) codegen.Command {
	if chartOnly {
		return codegen.Command{
			AppName:      appName,
			ManifestRoot: certAgentManifestRoot,
			Chart:        helm.CertAgentChart,
		}
	}

	return codegen.Command{
		AppName:           appName,
		AnyVendorConfig:   anyvendorImports,
		ManifestRoot:      certAgentManifestRoot,
		TopLevelTemplates: topLevelTemplates,
		Groups:            groups.CertAgentGroups,
		RenderProtos:      true,
		Chart:             helm.CertAgentChart,
	}
}

func makeTopLevelTemplate(templateFunc func(params contrib.CrossGroupTemplateParameters) model.CustomTemplates, outPath string, resourceSnapshot io.Snapshot) model.CustomTemplates {
	return templateFunc(contrib.CrossGroupTemplateParameters{
		OutputFilename:    outPath,
		SelectFromGroups:  allApiGroups,
		ResourcesToSelect: resourceSnapshot,
	})
}
