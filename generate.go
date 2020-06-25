package main

import (
	"log"

	"github.com/solo-io/service-mesh-hub/pkg/common/constants"
	"github.com/solo-io/skv2/codegen"
	"github.com/solo-io/skv2/codegen/model"
	"github.com/solo-io/skv2/contrib"
	"github.com/solo-io/solo-kit/pkg/code-generator/sk_anyvendor"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	appName            = "service-mesh-hub"
	smhModule          = "github.com/solo-io/service-mesh-hub"
	v1alpha1Version    = "v1alpha1"
	apiRoot            = "pkg/api"
	smhCrdManifestRoot = "install/helm/charts/custom-resource-definitions"
	csrCrdManifestRoot = "install/helm/charts/csr-agent/"

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
	groups := []model.Group{
		makeGroup("core", v1alpha1Version, []resourceToGenerate{
			{kind: "Settings", noStatus: true},
		}),
		makeGroup("discovery", v1alpha1Version, []resourceToGenerate{
			{kind: "KubernetesCluster", noStatus: true}, // TODO(ilackarms): remove this kubernetes cluster and use skv2 multicluster
			{kind: "MeshService"},
			{kind: "MeshWorkload"},
			{kind: "Mesh"},
		}),
		makeGroup("networking", v1alpha1Version, []resourceToGenerate{
			{kind: "TrafficPolicy"},
			{kind: "AccessControlPolicy"},
			{kind: "VirtualMesh"},
			{kind: "FailoverService"},
		}),
	}

	return codegen.Command{
		AppName:         appName,
		AnyVendorConfig: protoImports,
		ManifestRoot:    smhCrdManifestRoot,
		Groups:          groups,
	}
}

func makeCsrCommand() codegen.Command {
	groups := []model.Group{
		makeGroup("security", v1alpha1Version, []resourceToGenerate{
			{kind: "VirtualMeshCertificateSigningRequest"},
		}),
	}

	return codegen.Command{
		AppName:         appName,
		AnyVendorConfig: protoImports,
		ManifestRoot:    csrCrdManifestRoot,
		Groups:          groups,
	}
}

type resourceToGenerate struct {
	kind     string
	noStatus bool // don't put a status on this resource
}

func makeGroup(groupPrefix, version string, resourcesToGenerate []resourceToGenerate) model.Group {
	var resources []model.Resource
	for _, resource := range resourcesToGenerate {
		res := model.Resource{
			Kind: resource.kind,
			Spec: model.Field{
				Type: model.Type{
					Name: resource.kind + "Spec",
				},
			},
		}
		if !resource.noStatus {
			res.Status = &model.Field{Type: model.Type{
				Name: resource.kind + "Status",
			}}
		}
		resources = append(resources, res)
	}

	return model.Group{
		GroupVersion: schema.GroupVersion{
			Group:   groupPrefix + "." + constants.ServiceMeshHubApiGroupSuffix,
			Version: version,
		},
		Module:           smhModule,
		Resources:        resources,
		RenderManifests:  true,
		RenderTypes:      true,
		RenderClients:    true,
		RenderController: true,
		RenderProtos:     true,
		MockgenDirective: true,
		CustomTemplates:  contrib.AllCustomTemplates,
		ApiRoot:          apiRoot,
	}
}
