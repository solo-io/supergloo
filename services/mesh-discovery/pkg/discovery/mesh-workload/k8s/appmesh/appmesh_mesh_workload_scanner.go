package appmesh

import (
	"context"
	"fmt"
	"strings"

	zephyr_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/env"
	"github.com/solo-io/service-mesh-hub/pkg/metadata"
	"github.com/solo-io/service-mesh-hub/services/common/constants"
	aws_utils "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/compute-target/aws/parser"
	meshworkload_discovery "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/mesh-workload/k8s"
	k8s_core_types "k8s.io/api/core/v1"
	k8s_meta_types "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	DiscoveryLabels = func() map[string]string {
		return map[string]string{
			constants.MESH_TYPE: strings.ToLower(zephyr_core_types.MeshType_APPMESH.String()),
		}
	}
)

func AppMeshWorkloadScannerFactoryProvider(
	appMeshParser aws_utils.AppMeshScanner,
) meshworkload_discovery.MeshWorkloadScannerFactory {
	return func(
		ownerFetcher meshworkload_discovery.OwnerFetcher,
		meshClient zephyr_discovery.MeshClient,
		remoteClient client.Client,
	) meshworkload_discovery.MeshWorkloadScanner {
		return NewAppMeshWorkloadScanner(
			ownerFetcher,
			appMeshParser,
			meshClient,
			remoteClient,
		)
	}
}

// visible for testing
func NewAppMeshWorkloadScanner(
	ownerFetcher meshworkload_discovery.OwnerFetcher,
	appMeshParser aws_utils.AppMeshScanner,
	meshClient zephyr_discovery.MeshClient,
	remoteClient client.Client,
) meshworkload_discovery.MeshWorkloadScanner {
	return &appMeshWorkloadScanner{
		ownerFetcher:   ownerFetcher,
		meshClient:     meshClient,
		appmeshScanner: appMeshParser,
		remoteClient:   remoteClient,
	}
}

type appMeshWorkloadScanner struct {
	ownerFetcher   meshworkload_discovery.OwnerFetcher
	appmeshScanner aws_utils.AppMeshScanner
	meshClient     zephyr_discovery.MeshClient
	remoteClient   client.Client
}

func (a *appMeshWorkloadScanner) ScanPod(ctx context.Context, pod *k8s_core_types.Pod, clusterName string) (*zephyr_discovery.MeshWorkload, error) {
	appMeshPod, err := a.appmeshScanner.ScanPodForAppMesh(ctx, pod, a.remoteClient)
	if err != nil {
		return nil, err
	}
	if appMeshPod == nil {
		return nil, nil
	}
	deployment, err := a.ownerFetcher.GetDeployment(ctx, pod)
	if err != nil {
		return nil, err
	}
	mesh, err := a.meshClient.GetMesh(
		ctx, client.ObjectKey{
			Name:      metadata.BuildAppMeshName(appMeshPod.AppMeshName, appMeshPod.Region, appMeshPod.AwsAccountID),
			Namespace: env.GetWriteNamespace(),
		},
	)
	if err != nil {
		return nil, err
	}
	return &zephyr_discovery.MeshWorkload{
		ObjectMeta: k8s_meta_types.ObjectMeta{
			Name:      a.buildMeshWorkloadName(deployment.GetName(), deployment.GetNamespace(), clusterName),
			Namespace: env.GetWriteNamespace(),
			Labels:    DiscoveryLabels(),
		},
		Spec: zephyr_discovery_types.MeshWorkloadSpec{
			KubeController: &zephyr_discovery_types.MeshWorkloadSpec_KubeController{
				KubeControllerRef: &zephyr_core_types.ResourceRef{
					Name:      deployment.GetName(),
					Namespace: deployment.GetNamespace(),
					Cluster:   clusterName,
				},
				Labels:             pod.GetLabels(),
				ServiceAccountName: pod.Spec.ServiceAccountName,
			},
			Mesh: &zephyr_core_types.ResourceRef{
				Name:      mesh.GetName(),
				Namespace: mesh.GetNamespace(),
			},
		},
	}, nil
}

func (a *appMeshWorkloadScanner) buildMeshWorkloadName(deploymentName string, namespace string, clusterName string) string {
	// TODO: https://github.com/solo-io/service-mesh-hub/issues/141
	return fmt.Sprintf("%s-%s-%s-%s", "appmesh", deploymentName, namespace, clusterName)
}
