package appmesh

import (
	"context"
	"fmt"
	"strings"

	"github.com/solo-io/go-utils/contextutils"
	smh_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1/types"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	smh_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/types"
	aws_utils "github.com/solo-io/service-mesh-hub/pkg/aws/parser"
	container_runtime "github.com/solo-io/service-mesh-hub/pkg/container-runtime"
	"github.com/solo-io/service-mesh-hub/pkg/kube"
	"github.com/solo-io/service-mesh-hub/pkg/kube/metadata"
	meshworkload_discovery "github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/discovery/mesh-workload/k8s"
	k8s_apps_types "k8s.io/api/apps/v1"
	k8s_core_types "k8s.io/api/core/v1"
	k8s_meta_types "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	DiscoveryLabels = func() map[string]string {
		return map[string]string{
			kube.MESH_TYPE: strings.ToLower(smh_core_types.MeshType_APPMESH.String()),
		}
	}
)

func AppMeshWorkloadScannerFactoryProvider(
	appMeshParser aws_utils.AppMeshScanner,
	awsAccountIdFetcher aws_utils.AwsAccountIdFetcher,
) meshworkload_discovery.MeshWorkloadScannerFactory {
	return func(
		ownerFetcher meshworkload_discovery.OwnerFetcher,
		meshClient smh_discovery.MeshClient,
		remoteClient client.Client,
	) meshworkload_discovery.MeshWorkloadScanner {
		return NewAppMeshWorkloadScanner(
			ownerFetcher,
			appMeshParser,
			meshClient,
			awsAccountIdFetcher,
			remoteClient,
		)
	}
}

// visible for testing
func NewAppMeshWorkloadScanner(
	ownerFetcher meshworkload_discovery.OwnerFetcher,
	appMeshParser aws_utils.AppMeshScanner,
	meshClient smh_discovery.MeshClient,
	awsAccountIdFetcher aws_utils.AwsAccountIdFetcher,
	remoteClient client.Client,
) meshworkload_discovery.MeshWorkloadScanner {
	return &appMeshWorkloadScanner{
		ownerFetcher:        ownerFetcher,
		meshClient:          meshClient,
		appmeshScanner:      appMeshParser,
		remoteClient:        remoteClient,
		awsAccountIdFetcher: awsAccountIdFetcher,
	}
}

type appMeshWorkloadScanner struct {
	ownerFetcher        meshworkload_discovery.OwnerFetcher
	appmeshScanner      aws_utils.AppMeshScanner
	meshClient          smh_discovery.MeshClient
	remoteClient        client.Client
	awsAccountIdFetcher aws_utils.AwsAccountIdFetcher
}

func (a *appMeshWorkloadScanner) ScanPod(ctx context.Context, pod *k8s_core_types.Pod, clusterName string) (*smh_discovery.MeshWorkload, error) {
	awsAccountId, err := a.awsAccountIdFetcher.GetEksAccountId(ctx, a.remoteClient)
	if err != nil {
		contextutils.LoggerFrom(ctx).Warnf("Error fetching AWS Account ID from ConfigMap: %+v", err)
	}
	appMeshPod, err := a.appmeshScanner.ScanPodForAppMesh(pod, awsAccountId)
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
			Namespace: container_runtime.GetWriteNamespace(),
		},
	)
	if err != nil {
		return nil, err
	}
	containerPorts := a.getContainerPorts(deployment)
	return &smh_discovery.MeshWorkload{
		ObjectMeta: k8s_meta_types.ObjectMeta{
			Name:      a.buildMeshWorkloadName(deployment.GetName(), deployment.GetNamespace(), clusterName),
			Namespace: container_runtime.GetWriteNamespace(),
			Labels:    DiscoveryLabels(),
		},
		Spec: smh_discovery_types.MeshWorkloadSpec{
			KubeController: &smh_discovery_types.MeshWorkloadSpec_KubeController{
				KubeControllerRef: &smh_core_types.ResourceRef{
					Name:      deployment.GetName(),
					Namespace: deployment.GetNamespace(),
					Cluster:   clusterName,
				},
				Labels:             pod.GetLabels(),
				ServiceAccountName: pod.Spec.ServiceAccountName,
			},
			Mesh: &smh_core_types.ResourceRef{
				Name:      mesh.GetName(),
				Namespace: mesh.GetNamespace(),
			},
			Appmesh: &smh_discovery_types.MeshWorkloadSpec_Appmesh{
				VirtualNodeName: appMeshPod.VirtualNodeName,
				Ports:           containerPorts,
			},
		},
	}, nil
}

func (a *appMeshWorkloadScanner) buildMeshWorkloadName(deploymentName string, namespace string, clusterName string) string {
	// TODO: https://github.com/solo-io/service-mesh-hub/issues/141
	return fmt.Sprintf("%s-%s-%s-%s", "appmesh", deploymentName, namespace, clusterName)
}

func (a *appMeshWorkloadScanner) getContainerPorts(
	deployment *k8s_apps_types.Deployment,
) []*smh_discovery_types.MeshWorkloadSpec_Appmesh_ContainerPort {
	var ports []*smh_discovery_types.MeshWorkloadSpec_Appmesh_ContainerPort
	for _, container := range deployment.Spec.Template.Spec.Containers {
		for _, containerPort := range container.Ports {
			ports = append(ports, &smh_discovery_types.MeshWorkloadSpec_Appmesh_ContainerPort{
				Port:     uint32(containerPort.ContainerPort),
				Protocol: string(containerPort.Protocol),
			})
		}
	}
	return ports
}
