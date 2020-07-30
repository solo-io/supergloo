package main

import (
	"flag"
	"log"
	"os"

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

var (
	appName = "service-mesh-hub"

	discoveryInputSnapshotCodePath  = "pkg/api/discovery.smh.solo.io/snapshot/input/snapshot.go"
	discoveryReconcilerCodePath     = "pkg/api/discovery.smh.solo.io/snapshot/input/reconciler.go"
	discoveryOutputSnapshotCodePath = "pkg/api/discovery.smh.solo.io/snapshot/output/snapshot.go"

	networkingInputSnapshotCodePath            = "pkg/api/networking.smh.solo.io/snapshot/input/snapshot.go"
	networkingTestInputSnapshotBuilderCodePath = "pkg/api/networking.smh.solo.io/snapshot/input/test/snapshot_builder.go"
	networkingReconcilerSnapshotCodePath       = "pkg/api/networking.smh.solo.io/snapshot/input/reconciler.go"
	networkingOutputIstioSnapshotCodePath      = "pkg/api/networking.smh.solo.io/snapshot/output/istio/snapshot.go"

	smhManifestRoot = "install/helm/service-mesh-hub"
	csrManifestRoot = "install/helm/csr-agent/"

	vendoredMulticlusterCRDs = "vendor_any/github.com/solo-io/skv2/crds/multicluster.solo.io_v1alpha1_crds.yaml"
	importedMulticlusterCRDs = smhManifestRoot + "/crds/multicluster.solo.io_v1alpha1_crds.yaml"

	allApiGroups = map[string][]model.Group{
		"":                                 groups.SMHGroups,
		"github.com/solo-io/external-apis": externalapis.Groups,
		"github.com/solo-io/skv2":          {skv1alpha1.Group},
	}

	// define custom templates
	discoveryInputSnapshot = makeTopLevelTemplate(
		contrib.InputSnapshot,
		discoveryInputSnapshotCodePath,
		io.DiscoveryInputTypes,
	)

	discoveryReconciler = makeTopLevelTemplate(
		contrib.InputReconciler,
		discoveryReconcilerCodePath,
		io.DiscoveryInputTypes,
	)

	discoveryOutputSnapshot = makeTopLevelTemplate(
		contrib.OutputSnapshot,
		discoveryOutputSnapshotCodePath,
		io.DiscoveryOutputTypes,
	)

	networkingInputSnapshot = makeTopLevelTemplate(
		contrib.InputSnapshot,
		networkingInputSnapshotCodePath,
		io.NetworkingInputTypes,
	)

	networkingInputSnapshotManualBuilder = makeTopLevelTemplate(
		contrib.InputSnapshotManualBuilder,
		networkingTestInputSnapshotBuilderCodePath,
		io.NetworkingInputTypes,
	)

	networkingReconciler = makeTopLevelTemplate(
		contrib.InputReconciler,
		networkingReconcilerSnapshotCodePath,
		io.NetworkingInputTypes,
	)

	networkingOutputIstioSnapshot = makeTopLevelTemplate(
		contrib.OutputSnapshot,
		networkingOutputIstioSnapshotCodePath,
		io.NetworkingOutputIstioTypes,
	)

	topLevelTemplates = []model.CustomTemplates{
		discoveryInputSnapshot,
		discoveryReconciler,
		discoveryOutputSnapshot,
		networkingInputSnapshot,
		networkingInputSnapshotManualBuilder,
		networkingReconciler,
		networkingOutputIstioSnapshot,
	}

	anyvendorImports = sk_anyvendor.CreateDefaultMatchOptions([]string{
		"api/**/*.proto",
	})
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	log.Printf("generating service mesh hub code with version %v", version.Version)
	chartOnly := flag.Bool("chart", false, "only generate the helm chart")
	flag.Parse()

	if err := makeSmhCommand(*chartOnly).Execute(); err != nil {
		return err
	}

	if *chartOnly {
		return nil
	}

	// TODO(ilackarms): we copy skv2 CRDs out of vendor_any into our helm chart.
	// we should consider using skv2 to automate this step for us
	if err := os.Rename(vendoredMulticlusterCRDs, importedMulticlusterCRDs); err != nil {
		return err
	}

	return nil
	// TODO(ilackarms): revive this when reviving csr agent
	log.Printf("generating csr-agent")
	if err := makeCsrCommand().Execute(); err != nil {
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

func makeCsrCommand() codegen.Command {
	return codegen.Command{
		AppName:         appName,
		AnyVendorConfig: anyvendorImports,
		ManifestRoot:    csrManifestRoot,
		Groups:          groups.CSRGroups,
		RenderProtos:    true,
	}
}

func makeTopLevelTemplate(templateFunc func(params contrib.CrossGroupTemplateParameters) model.CustomTemplates, outPath string, resourceSnapshot io.Snapshot) model.CustomTemplates {
	return templateFunc(contrib.CrossGroupTemplateParameters{
		OutputFilename:    outPath,
		SelectFromGroups:  allApiGroups,
		ResourcesToSelect: resourceSnapshot,
	})
}
