package appmesh

import (
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
)

func AppmeshMesh(name string) *v1.Mesh {
	if name == "" {
		name = "appmesh"
	}
	return AppmeshMeshWithConfig("supergloo-system", name, "vn")
}

func AppmeshMeshWithConfig(namespace, name, vnLabel string) *v1.Mesh {
	return &v1.Mesh{
		Metadata: core.Metadata{
			Name:      name,
			Namespace: namespace,
		},
		MeshType: &v1.Mesh_AwsAppMesh{
			AwsAppMesh: &v1.AwsAppMesh{
				Region:           "us-east-1",
				VirtualNodeLabel: vnLabel,
				EnableAutoInject: true,
				SidecarPatchConfigMap: &core.ResourceRef{
					Name:      "sidecar-injector-webhook-configmap",
					Namespace: "supergloo-system",
				},
				InjectionSelector: &v1.PodSelector{
					SelectorType: &v1.PodSelector_NamespaceSelector_{
						NamespaceSelector: &v1.PodSelector_NamespaceSelector{
							Namespaces: []string{"namespace-with-inject"}}}}}}}
}
