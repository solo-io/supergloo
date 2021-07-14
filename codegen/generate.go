package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	"github.com/rotisserie/eris"
	externalapis "github.com/solo-io/external-apis/codegen"
	"github.com/solo-io/gloo-mesh/codegen/anyvendor"
	"github.com/solo-io/gloo-mesh/codegen/groups"
	"github.com/solo-io/gloo-mesh/codegen/helm"
	"github.com/solo-io/gloo-mesh/codegen/io"
	gloomeshmodel "github.com/solo-io/gloo-mesh/codegen/model"
	"github.com/solo-io/gloo-mesh/pkg/common/version"
	skv1alpha1 "github.com/solo-io/skv2/api/multicluster/v1alpha1"
	"github.com/solo-io/skv2/codegen"
	"github.com/solo-io/skv2/codegen/model"
	"github.com/solo-io/skv2/codegen/util"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

var (
	appName = "gloo-mesh"

	topLevelComponents = []gloomeshmodel.TopLevelComponent{
		// discovery component
		{
			GeneratedCodeRoot:    "pkg/api/discovery.mesh.gloo.solo.io",
			RemoteInputResources: io.DiscoveryInputTypes,
			RemoteSnapshotName:   "DiscoveryInput",
			LocalInputResources:  io.DiscoveryLocalInputTypes,
			LocalSnapshotName:    "Settings",
			OutputResources:      []io.OutputSnapshot{io.DiscoveryOutputTypes},
			AgentMode:            true,
		},
		// networking component
		{
			GeneratedCodeRoot:    "pkg/api/networking.mesh.gloo.solo.io",
			LocalInputResources:  io.NetworkingInputTypes,
			RemoteInputResources: io.IstioNetworkingOutputTypes.Snapshot,
			OutputResources: []io.OutputSnapshot{
				io.IstioNetworkingOutputTypes,
				io.SmiNetworkingOutputTypes,
				io.LocalNetworkingOutputTypes,
				io.AppMeshNetworkingOutputTypes,
			},
		},
		// certificate issuer component
		{
			GeneratedCodeRoot:    "pkg/api/certificates.mesh.gloo.solo.io/issuer",
			RemoteInputResources: io.CertificateIssuerInputTypes,
		},
		// certificate agent component
		{
			GeneratedCodeRoot:   "pkg/api/certificates.mesh.gloo.solo.io/agent",
			LocalInputResources: io.CertificateAgentInputTypes,
			OutputResources:     []io.OutputSnapshot{io.CertificateAgentOutputTypes},
		},
	}

	glooMeshManifestRoot     = "install/helm/gloo-mesh/"
	glooMeshCrdsManifestRoot = "install/helm/gloo-mesh-crds/"
	certAgentManifestRoot    = "install/helm/cert-agent/"
	agentCrdsManifestRoot    = "install/helm/agent-crds/"

	vendoredMultiClusterCRDs = "vendor_any/github.com/solo-io/skv2/crds/multicluster.solo.io_v1alpha1_crds.yaml"
	importedMultiClusterCRDs = glooMeshCrdsManifestRoot + "/crds/multicluster.solo.io_v1alpha1_crds.yaml"
	vendoredRatelimitCRDs    = "vendor_any/github.com/solo-io/solo-apis/crds/ratelimit.solo.io_v1alpha1_crds.yaml"
	importedRatelimitCRDs    = glooMeshCrdsManifestRoot + "/crds/ratelimit.solo.io_v1alpha1_crds.yaml"

	snapshotApiGroups = map[string][]model.Group{
		"":                                 groups.AllGeneratedGroups,
		"github.com/solo-io/external-apis": externalapis.Groups,
		"github.com/solo-io/skv2":          {skv1alpha1.Group},
		//"github.com/solo-io/solo-apis":     soloapi_codegen.RateLimiterGroups(),
	}

	project = gloomeshmodel.Project{
		SnapshotApiGroups:  snapshotApiGroups,
		TopLevelComponents: topLevelComponents,
	}

	anyvendorImports = anyvendor.AnyVendorImports()
)

func run() error {
	log.Printf("generating gloo mesh code with version %v", version.Version)
	chartOnly := flag.Bool("chart", false, "only generate the helm chart")
	flag.Parse()

	if err := makeGlooMeshCrdsCommand().Execute(); err != nil {
		return err
	}

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

	if err := generateHelmValuesDoc(helm.Chart, "Gloo Mesh", "gloo_mesh_helm_values_reference.md"); err != nil {
		return err
	}

	if err := generateHelmValuesDoc(helm.CertAgentChart, "Cert Agent", "cert_agent_helm_values_reference.md"); err != nil {
		return err
	}

	// TODO(ilackarms): we copy skv2 CRDs out of vendor_any into our helm chart.
	// we should consider using skv2 to automate this step for us
	if err := os.Rename(vendoredMultiClusterCRDs, importedMultiClusterCRDs); err != nil {
		return err
	}

	// copy solo-apis ratelimit crds into helm chart
	if err := os.Rename(vendoredRatelimitCRDs, importedRatelimitCRDs); err != nil {
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
		TopLevelTemplates: project.TopLevelTemplates(),
		Chart:             helm.Chart,
	}
}

func makeGlooMeshCrdsCommand() codegen.Command {
	return codegen.Command{
		AppName:         appName,
		AnyVendorConfig: anyvendorImports,
		ManifestRoot:    glooMeshCrdsManifestRoot,
		Groups:          groups.GlooMeshGroups,
		RenderProtos:    true,
		Chart:           helm.CrdsChart,
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
		AppName:         appName,
		AnyVendorConfig: anyvendorImports,
		ManifestRoot:    certAgentManifestRoot,
		RenderProtos:    true,
		Chart:           helm.CertAgentChart,
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

func generateHelmValuesDoc(chart *model.Chart, title string, filename string) error {
	originalDir, err := os.Getwd()
	if err != nil {
		return err
	}

	// change wd to codegen dir
	codegenDir := util.MustGetThisDir()
	if err := os.Chdir(codegenDir); err != nil {
		log.Fatal(err)
	}

	// change back to original directory
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			log.Fatal(err)
		}
	}()

	helmValuesDoc := chart.GenerateHelmDoc(title)

	filename = filepath.Join("helm", filename)

	f, err := os.OpenFile(filename, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0777)
	if err != nil {
		return eris.Errorf("error generating Helm values reference doc: %v", err)
	}

	if _, err := f.Write([]byte(helmValuesDoc)); err != nil {
		return eris.Errorf("error generating Helm values reference doc: %v", err)
	}

	return nil
}
