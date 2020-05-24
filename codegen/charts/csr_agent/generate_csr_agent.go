package main

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/solo-io/service-mesh-hub/pkg/constants"
	"github.com/solo-io/skv2/codegen"
	"github.com/solo-io/skv2/codegen/model"
	"github.com/solo-io/solo-kit/pkg/code-generator/sk_anyvendor"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

//go:generate go run generate_csr_agent.go

func main() {
	log.Println("starting generate CSR agent")

	var renderTypes bool
	if os.Getenv("REGENERATE_TYPES") == "" {
		log.Println("REGENERATE_TYPES is not set, skipping autopilot client gen")
	} else {
		renderTypes = true
	}

	// load custom client template
	customClientTemplateBytes, err := ioutil.ReadFile("../custom_client.gotmpl")
	customClientTemplate := string(customClientTemplateBytes)
	if err != nil {
		log.Fatal(err)
	}
	// load custom client providers template
	customClientProvidersBytes, err := ioutil.ReadFile("../custom_client_providers.gotmpl")
	customClientProviders := string(customClientProvidersBytes)
	if err != nil {
		log.Fatal(err)
	}

	apImports := sk_anyvendor.CreateDefaultMatchOptions([]string{
		"api/**/*.proto",
	})
	skv2Cmd := codegen.Command{
		AppName: "service-mesh-hub",
		Groups: []model.Group{
			{
				GroupVersion: schema.GroupVersion{
					Group:   "security." + constants.ServiceMeshHubApiGroupSuffix,
					Version: "v1alpha1",
				},
				Module: "github.com/solo-io/service-mesh-hub",
				Resources: []model.Resource{
					{
						Kind: "VirtualMeshCertificateSigningRequest",
						Spec: model.Field{
							Type: model.Type{
								Name:      "VirtualMeshCertificateSigningRequestSpec",
								GoPackage: "github.com/solo-io/service-mesh-hub/pkg/api/security.zephyr.solo.io/v1alpha1/types",
							},
						},
						Status: &model.Field{Type: model.Type{
							Name:      "VirtualMeshCertificateSigningRequestStatus",
							GoPackage: "github.com/solo-io/service-mesh-hub/pkg/api/security.zephyr.solo.io/v1alpha1/types",
						}},
					},
				},
				RenderManifests:  true,
				RenderTypes:      renderTypes,
				RenderController: true,
				RenderProtos:     true,
				CustomTemplates: model.CustomTemplates{
					Templates: map[string]string{
						"clients.go":          customClientTemplate,
						"client_providers.go": customClientProviders,
					},
				},
				ApiRoot: "pkg/api",
			},
		},
		AnyVendorConfig: apImports,
		ManifestRoot:    "install/helm/charts/csr-agent/",
	}
	if err := skv2Cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
