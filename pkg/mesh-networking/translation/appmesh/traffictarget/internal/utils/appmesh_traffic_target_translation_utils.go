package utils

import (
	appmeshv1beta2 "github.com/aws/aws-app-mesh-controller-for-k8s/apis/appmesh/v1beta2"
	"github.com/rotisserie/eris"
	discoveryv1alpha2 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1alpha2"
	discoveryv1alpha2sets "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1alpha2/sets"
	"github.com/solo-io/skv2/contrib/pkg/sets"
	"k8s.io/apimachinery/pkg/types"
)

func GetAppMeshMeshRef(trafficTarget *discoveryv1alpha2.TrafficTarget, meshes discoveryv1alpha2sets.MeshSet) (*appmeshv1beta2.MeshReference, error) {
	if trafficTarget.Spec.GetKubeService() == nil {
		// TODO non kube services currently unsupported
		return nil, eris.New("Cannot get App Mesh mesh ref for non-kube service")
	}

	mesh, err := meshes.Find(trafficTarget.Spec.Mesh)
	if err != nil {
		return nil, err
	}

	appMesh := mesh.Spec.GetAwsAppMesh()
	if appMesh == nil {
		return nil, eris.Errorf("Cannot access App Mesh Mesh ref from non App Mesh mesh %s", sets.Key(mesh))
	}

	if appMesh.ClusterMeshResources == nil {
		return nil, eris.Errorf("%s has no discovered clusters.", sets.Key(mesh))
	}

	cluster := trafficTarget.Spec.GetKubeService().Ref.ClusterName
	if ref, ok := appMesh.ClusterMeshResources[cluster]; ok {
		return &appmeshv1beta2.MeshReference{
			Name: ref.Name,
			UID:  types.UID(ref.Uid),
		}, nil
	}
	return nil, eris.Errorf("%s does not manage cluster %s.", sets.Key(mesh), cluster)
}
