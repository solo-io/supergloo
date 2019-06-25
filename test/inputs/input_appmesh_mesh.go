package inputs

import (
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
)

func AppMeshMesh(namespace, region string, secretRef *core.ResourceRef) *v1.Mesh {
	return &v1.Mesh{
		Metadata: core.Metadata{
			Namespace: namespace,
			Name:      "appmesh",
		},
		MeshType: &v1.Mesh_AwsAppMesh{
			AwsAppMesh: &v1.AwsAppMesh{
				AwsSecret: secretRef,
				Region:    region,
			},
		},
	}
}
