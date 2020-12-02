package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	"github.com/solo-io/skv2/codegen/render"

	externalapis "github.com/solo-io/external-apis/codegen"
	"github.com/solo-io/gloo-mesh/codegen/anyvendor"
	"github.com/solo-io/gloo-mesh/codegen/groups"
	"github.com/solo-io/gloo-mesh/codegen/helm"
	"github.com/solo-io/gloo-mesh/codegen/io"
	"github.com/solo-io/gloo-mesh/pkg/common/version"
	skv1alpha1 "github.com/solo-io/skv2/api/multicluster/v1alpha1"
	"github.com/solo-io/skv2/codegen"
	"github.com/solo-io/skv2/codegen/model"
	"github.com/solo-io/skv2/contrib"
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
	inputResources []io.Snapshot

	// the set of output resources for which to generate a snapshot
	outputResources []io.Snapshot
}

func (t topLevelComponent) makeCodegenTemplates() []model.CustomTemplates {
	var topLevelTemplates []model.CustomTemplates

	for _, inputResources := range t.inputResources {
		topLevelTemplates = append(topLevelTemplates, makeTopLevelTemplate(
			contrib.InputSnapshot,
			filepath.Join(t.generatedCodeRoot, "input", inputResources.Name, "snapshot.go"),
			inputResources.Resources,
		))

		topLevelTemplates = append(topLevelTemplates, makeTopLevelTemplate(
			contrib.InputReconciler,
			filepath.Join(t.generatedCodeRoot, "input", inputResources.Name, "reconciler.go"),
			inputResources.Resources,
		))

		topLevelTemplates = append(topLevelTemplates, makeTopLevelTemplate(
			contrib.InputSnapshotManualBuilder,
			filepath.Join(t.generatedCodeRoot, "input", inputResources.Name, "snapshot_manual_builder.go"),
			inputResources.Resources,
		))

	}

	for _, outputResources := range t.outputResources {
		filePath := filepath.Join(t.generatedCodeRoot, "output", outputResources.Name, "snapshot.go")
		topLevelTemplates = append(topLevelTemplates, makeTopLevelTemplate(
			contrib.OutputSnapshot,
			filePath,
			outputResources.Resources,
		))
	}

	return topLevelTemplates
}

var (
	appName = "gloo-mesh"

	topLevelComponents = []topLevelComponent{
		// discovery component
		{
			generatedCodeRoot: "pkg/api/discovery.mesh.gloo.solo.io",
			inputResources:    []io.Snapshot{io.DiscoveryInputTypes},
			outputResources:   []io.Snapshot{io.DiscoveryOutputTypes},
		},
		// networking snapshot
		{
			generatedCodeRoot: "pkg/api/networking.mesh.gloo.solo.io",
			inputResources: []io.Snapshot{
				io.NetworkingInputTypes,
				io.IstioNetworkingOutputTypes, // needed to initiate watches on istio output types
			},
			outputResources: []io.Snapshot{
				io.IstioNetworkingOutputTypes,
				io.SmiNetworkingOutputTypes,
				io.LocalNetworkingOutputTypes,
				io.AppMeshNetworkingOutputTypes,
			},
		},
		// certificate issuer component
		{
			generatedCodeRoot: "pkg/api/certificates.mesh.gloo.solo.io/issuer",
			inputResources:    []io.Snapshot{io.CertificateIssuerInputTypes},
		},
		// certificate agent component
		{
			generatedCodeRoot: "pkg/api/certificates.mesh.gloo.solo.io/agent",
			inputResources:    []io.Snapshot{io.CertificateAgentInputTypes},
			outputResources:   []io.Snapshot{io.CertificateAgentOutputTypes},
		},
	}

	glooMeshManifestRoot  = "install/helm/gloo-mesh"
	certAgentManifestRoot = "install/helm/cert-agent/"
	xdsAgentManifestRoot  = "install/helm/xds-agent/"

	vendoredMultiClusterCRDs = "vendor_any/github.com/solo-io/skv2/crds/multicluster.solo.io_v1alpha1_crds.yaml"
	importedMultiClusterCRDs = glooMeshManifestRoot + "/crds/multicluster.solo.io_v1alpha1_crds.yaml"

	allApiGroups = map[string][]model.Group{
		"":                                 append(append(groups.GlooMeshGroups, groups.CertAgentGroups...), groups.XdsAgentGroup),
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

	anyvendorImports = anyvendor.AnyVendorImports()
)

func run() error {
	log.Printf("generating gloo mesh code with version %v", version.Version)
	chartOnly := flag.Bool("chart", false, "only generate the helm chart")
	flag.Parse()

	if err := makeGlooMeshCommand(*chartOnly).Execute(); err != nil {
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

	if err := makeXdsAgentCommand().Execute(); err != nil {
		return err
	}

	return nil
}

func makeGlooMeshCommand(chartOnly bool) codegen.Command {

	if chartOnly {
		return codegen.Command{
			AppName:      appName,
			ManifestRoot: glooMeshManifestRoot,
			Chart:        helm.Chart,
		}
	}

	return codegen.Command{
		AppName:           appName,
		AnyVendorConfig:   anyvendorImports,
		ManifestRoot:      glooMeshManifestRoot,
		TopLevelTemplates: topLevelTemplates,
		Groups:            groups.GlooMeshGroups,
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

func makeXdsAgentCommand() codegen.Command {
	return codegen.Command{
		AppName:      appName,
		ManifestRoot: xdsAgentManifestRoot,
		Groups:       []render.Group{groups.XdsAgentGroup},
		RenderProtos: true,
	}
}

func makeTopLevelTemplate(templateFunc func(params contrib.CrossGroupTemplateParameters) model.CustomTemplates, outPath string, resourceSnapshot io.SnapshotResources) model.CustomTemplates {
	return templateFunc(contrib.CrossGroupTemplateParameters{
		OutputFilename:    outPath,
		SelectFromGroups:  allApiGroups,
		ResourcesToSelect: resourceSnapshot,
	})
}
