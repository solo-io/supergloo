package inputs

import (
	v1 "github.com/solo-io/mesh-discovery/pkg/api/v1"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
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
