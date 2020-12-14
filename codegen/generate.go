package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"

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
	// local inptus are read from the local cluster where the controller runs
	localInputResources io.Snapshot
	// remote inptus are read from managed cluster registered to the controller cluster
	remoteInputResources io.Snapshot

	// the set of output resources for which to generate a snapshot
	outputResources []io.OutputSnapshot
}

func (t topLevelComponent) makeCodegenTemplates() []model.CustomTemplates {
	var topLevelTemplates []model.CustomTemplates

	switch {
	case t.localInputResources != nil && t.remoteInputResources != nil:
		topLevelTemplates = append(topLevelTemplates, makeTopLevelTemplate(
			contrib.InputSnapshot,
			"Local",
			t.generatedCodeRoot+"/input/local_snapshot.go",
			contrib.HomogenousSnapshotResources{ResourcesToSelect: t.localInputResources, MultiCluster: false},
		))
		topLevelTemplates = append(topLevelTemplates, makeTopLevelTemplate(
			contrib.InputSnapshot,
			"Remote",
			t.generatedCodeRoot+"/input/remote_snapshot.go",
			contrib.HomogenousSnapshotResources{ResourcesToSelect: t.remoteInputResources, MultiCluster: true},
		))
		topLevelTemplates = append(topLevelTemplates, makeTopLevelTemplate(
			contrib.InputSnapshotManualBuilder,
			"Local",
			t.generatedCodeRoot+"/input/local_snapshot_manual_builder.go",
			contrib.HomogenousSnapshotResources{ResourcesToSelect: t.localInputResources, MultiCluster: false},
		))
		topLevelTemplates = append(topLevelTemplates, makeTopLevelTemplate(
			contrib.InputSnapshotManualBuilder,
			"Remote",
			t.generatedCodeRoot+"/input/remote_snapshot_manual_builder.go",
			contrib.HomogenousSnapshotResources{ResourcesToSelect: t.remoteInputResources, MultiCluster: true},
		))

		topLevelTemplates = append(topLevelTemplates, makeTopLevelTemplate(
			contrib.InputReconciler,
			"",
			t.generatedCodeRoot+"/input/reconciler.go",
			contrib.HybridSnapshotResources{
				LocalResourcesToSelect:  t.localInputResources,
				RemoteResourcesToSelect: t.remoteInputResources,
			},
		))
	case t.localInputResources != nil:
		topLevelTemplates = append(topLevelTemplates, makeTopLevelTemplate(
			contrib.InputSnapshot,
			"",
			t.generatedCodeRoot+"/input/snapshot.go",
			contrib.HomogenousSnapshotResources{ResourcesToSelect: t.localInputResources, MultiCluster: false},
		))
		topLevelTemplates = append(topLevelTemplates, makeTopLevelTemplate(
			contrib.InputSnapshotManualBuilder,
			"",
			t.generatedCodeRoot+"/input/snapshot_manual_builder.go",
			contrib.HomogenousSnapshotResources{ResourcesToSelect: t.localInputResources, MultiCluster: false},
		))
		topLevelTemplates = append(topLevelTemplates, makeTopLevelTemplate(
			contrib.InputReconciler,
			"",
			t.generatedCodeRoot+"/input/reconciler.go",
			contrib.HomogenousSnapshotResources{ResourcesToSelect: t.localInputResources, MultiCluster: false},
		))
	case t.remoteInputResources != nil:
		topLevelTemplates = append(topLevelTemplates, makeTopLevelTemplate(
			contrib.InputSnapshot,
			"",
			t.generatedCodeRoot+"/input/snapshot.go",
			contrib.HomogenousSnapshotResources{ResourcesToSelect: t.remoteInputResources, MultiCluster: true},
		))
		topLevelTemplates = append(topLevelTemplates, makeTopLevelTemplate(
			contrib.InputSnapshotManualBuilder,
			"",
			t.generatedCodeRoot+"/input/snapshot_manual_builder.go",
			contrib.HomogenousSnapshotResources{ResourcesToSelect: t.remoteInputResources, MultiCluster: true},
		))
		topLevelTemplates = append(topLevelTemplates, makeTopLevelTemplate(
			contrib.InputReconciler,
			"",
			t.generatedCodeRoot+"/input/reconciler.go",
			contrib.HomogenousSnapshotResources{ResourcesToSelect: t.remoteInputResources, MultiCluster: true},
		))
	}

	for _, outputResources := range t.outputResources {
		filePath := filepath.Join(t.generatedCodeRoot, "output", outputResources.Name, "snapshot.go")
		topLevelTemplates = append(topLevelTemplates, makeTopLevelTemplate(
			contrib.OutputSnapshot,
			"",
			filePath,
			contrib.HomogenousSnapshotResources{ResourcesToSelect: outputResources.Snapshot},
		))
	}

	return topLevelTemplates
}

var (
	appName = "gloo-mesh"

	topLevelComponents = []topLevelComponent{
		// discovery component
		{
			generatedCodeRoot:    "pkg/api/discovery.mesh.gloo.solo.io",
			remoteInputResources: io.DiscoveryRemoteInputTypes,
			localInputResources:  io.DiscoveryLocalInputTypes,
			outputResources:      []io.OutputSnapshot{io.DiscoveryOutputTypes},
		},
		// networking snapshot
		{
			generatedCodeRoot:   "pkg/api/networking.mesh.gloo.solo.io",
			localInputResources: io.NetworkingInputTypes,
			outputResources: []io.OutputSnapshot{
				io.IstioNetworkingOutputTypes,
				io.SmiNetworkingOutputTypes,
				io.LocalNetworkingOutputTypes,
				io.AppMeshNetworkingOutputTypes,
			},
		},
		// certificate issuer component
		{
			generatedCodeRoot:    "pkg/api/certificates.mesh.gloo.solo.io/issuer",
			remoteInputResources: io.CertificateIssuerInputTypes,
		},
		// certificate agent component
		{
			generatedCodeRoot:   "pkg/api/certificates.mesh.gloo.solo.io/agent",
			localInputResources: io.CertificateAgentInputTypes,
			outputResources:     []io.OutputSnapshot{io.CertificateAgentOutputTypes},
		},
	}

	glooMeshManifestRoot  = "install/helm/gloo-mesh"
	certAgentManifestRoot = "install/helm/cert-agent/"
	agentCrdsManifestRoot = "install/helm/agent-crds/"

	vendoredMultiClusterCRDs = "vendor_any/github.com/solo-io/skv2/crds/multicluster.solo.io_v1alpha1_crds.yaml"
	importedMultiClusterCRDs = glooMeshManifestRoot + "/crds/multicluster.solo.io_v1alpha1_crds.yaml"

	snapshotApiGroups = map[string][]model.Group{
		"":                                 groups.AllGeneratedGroups,
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

	if err := makeAgentCrdsCommand().Execute(); err != nil {
		return err
	}

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
		RenderProtos:      true,
		Chart:             helm.CertAgentChart,
	}
}

func makeAgentCrdsCommand() codegen.Command {
	return codegen.Command{
		AppName:         appName,
		AnyVendorConfig: anyvendorImports,
		ManifestRoot:    agentCrdsManifestRoot,
		Groups:          append(groups.CertAgentGroups, groups.XdsAgentGroup),
		RenderProtos:    true,
		Chart:           helm.AgentCrdsChart,
	}
}

func makeTopLevelTemplate(templateFunc func(params contrib.SnapshotTemplateParameters) model.CustomTemplates, snapshotName, outPath string, snapshotResources contrib.SnapshotResources) model.CustomTemplates {
	return templateFunc(contrib.SnapshotTemplateParameters{
		SnapshotName:      snapshotName,
		OutputFilename:    outPath,
		SelectFromGroups:  snapshotApiGroups,
		SnapshotResources: snapshotResources,
	})
}
