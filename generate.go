package main

import (
	"log"

	"github.com/solo-io/service-mesh-hub/pkg/codegen"
	skv2_codegen "github.com/solo-io/skv2/codegen"
	"github.com/solo-io/skv2/codegen/model"
	"github.com/solo-io/solo-kit/pkg/code-generator/sk_anyvendor"
)

var (
	appName            = "service-mesh-hub"
	smhCrdManifestRoot = "install/helm/charts/custom-resource-definitions"
	csrCrdManifestRoot = "install/helm/charts/csr-agent/"
	protoImports       *sk_anyvendor.Imports
)

func main() {
	protoImports = sk_anyvendor.CreateDefaultMatchOptions([]string{
		"api/**/*.proto",
	})
	protoImports.External["github.com/solo-io/skv2"] = []string{
		"api/**/*.proto",
	}
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

func makeSmhCommand() skv2_codegen.Command {
	return skv2_codegen.Command{
		AppName:         appName,
		AnyVendorConfig: protoImports,
		ManifestRoot:    smhCrdManifestRoot,
		Groups:          codegen.SmhGroups,
		RenderProtos:    true,
	}
}

func makeCsrCommand() skv2_codegen.Command {
	groups := []model.Group{
		codegen.MakeGroup("security", codegen.V1alpha1Version, []codegen.ResourceToGenerate{
			{
				Kind: "VirtualMeshCertificateSigningRequest",
			},
		}),
	}

	return skv2_codegen.Command{
		AppName:         appName,
		AnyVendorConfig: protoImports,
		ManifestRoot:    csrCrdManifestRoot,
		Groups:          groups,
		RenderProtos:    true,
	}
}
