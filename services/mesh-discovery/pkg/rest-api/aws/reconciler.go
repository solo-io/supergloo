package aws

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appmesh"
	zephyr_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/env"
	rest_api "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/rest-api"
	k8s_meta_types "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	ObjectNamePrefix   = "appmesh"
	NumItemsPerRequest = aws.Int64(100)
	Region             = "us-east-2" // TODO remove hardcode and replace with configuration
)

type appMeshDiscoveryReconciler struct {
	meshPlatformName   string
	meshClient         zephyr_discovery.MeshClient
	meshWorkloadClient zephyr_discovery.MeshWorkloadClient
	meshServiceClient  zephyr_discovery.MeshServiceClient
	appMeshClient      *appmesh.AppMesh
}

type AppMeshDiscoveryReconcilerFactory func(
	meshPlatformName string,
	appMeshClient *appmesh.AppMesh,
) rest_api.RestAPIDiscoveryReconciler

func NewAppMeshDiscoveryReconcilerFactory(
	masterClient client.Client,
	meshClientFactory zephyr_discovery.MeshClientFactory,
	meshWorkloadClientFactory zephyr_discovery.MeshWorkloadClientFactory,
	meshServiceClientFactory zephyr_discovery.MeshServiceClientFactory,
) AppMeshDiscoveryReconcilerFactory {
	return func(
		meshPlatformName string,
		appMeshClient *appmesh.AppMesh,
	) rest_api.RestAPIDiscoveryReconciler {
		return NewAppMeshDiscoveryReconciler(
			meshClientFactory(masterClient),
			meshWorkloadClientFactory(masterClient),
			meshServiceClientFactory(masterClient),
			appMeshClient,
			meshPlatformName,
		)
	}
}

func NewAppMeshDiscoveryReconciler(
	meshClient zephyr_discovery.MeshClient,
	meshWorkloadClient zephyr_discovery.MeshWorkloadClient,
	meshServiceClient zephyr_discovery.MeshServiceClient,
	appMeshClient *appmesh.AppMesh,
	meshPlatformName string,
) rest_api.RestAPIDiscoveryReconciler {
	return &appMeshDiscoveryReconciler{
		meshClient:         meshClient,
		meshWorkloadClient: meshWorkloadClient,
		meshServiceClient:  meshServiceClient,
		appMeshClient:      appMeshClient,
		meshPlatformName:   meshPlatformName,
	}
}

func (a *appMeshDiscoveryReconciler) Reconcile(ctx context.Context) error {
	var err error
	meshes, smhMeshNames, err := a.reconcileMeshes(ctx, a.appMeshClient)
	if err != nil {
		return err
	}
	if err = a.reconcileMeshServices(ctx, a.appMeshClient, smhMeshNames, meshes); err != nil {
		return err
	}
	if err = a.reconcileMeshWorkloads(ctx, a.appMeshClient, smhMeshNames, meshes); err != nil {
		return err
	}
	return nil
}

func (a *appMeshDiscoveryReconciler) reconcileMeshes(
	ctx context.Context,
	appMeshClient *appmesh.AppMesh,
) (map[string]*zephyr_discovery.Mesh, sets.String, error) {
	var nextToken *string
	input := &appmesh.ListMeshesInput{
		Limit:     NumItemsPerRequest,
		NextToken: nextToken,
	}
	discoveredSMHMeshNames := sets.NewString() // Set containing SMH unique identifiers for AppMesh instances (the Mesh CRD name) for comparison
	discoveredMeshes := make(map[string]*zephyr_discovery.Mesh)
	for {
		appMeshes, err := appMeshClient.ListMeshes(input)
		if err != nil {
			return nil, nil, err
		}
		for _, appMeshRef := range appMeshes.Meshes {
			discoveredMesh := a.convertAppMeshMesh(appMeshRef)
			discoveredSMHMeshNames.Insert(discoveredMesh.GetName())
			err := a.meshClient.UpsertMeshSpec(ctx, discoveredMesh)
			discoveredMeshes[aws.StringValue(appMeshRef.MeshName)] = discoveredMesh
			if err != nil {
				return nil, nil, err
			}
		}
		if appMeshes.NextToken == nil {
			break
		}
		input.SetNextToken(aws.StringValue(appMeshes.NextToken))
	}

	meshList, err := a.meshClient.ListMesh(ctx)
	if err != nil {
		return nil, nil, err
	}
	for _, mesh := range meshList.Items {
		mesh := mesh
		if mesh.Spec.GetAwsAppMesh() == nil {
			continue
		}
		// Remove Meshes that no longer exist in AppMesh
		if !discoveredSMHMeshNames.Has(mesh.GetName()) {
			err := a.meshClient.DeleteMesh(ctx, client.ObjectKey{Name: mesh.GetName(), Namespace: mesh.GetNamespace()})
			if err != nil {
				return nil, nil, err
			}
		}
	}
	return discoveredMeshes, discoveredSMHMeshNames, nil
}

func (a *appMeshDiscoveryReconciler) reconcileMeshServices(
	ctx context.Context,
	appMeshClient *appmesh.AppMesh,
	smhMeshNames sets.String,
	meshes map[string]*zephyr_discovery.Mesh,
) error {
	input := &appmesh.ListVirtualServicesInput{
		Limit: NumItemsPerRequest,
	}
	discoveredMeshServiceSMHNames := sets.NewString() // Set containing SMH names for AppMesh VirtualServices for reconciling
	for appMeshName, mesh := range meshes {
		input.NextToken = nil
		input.SetMeshName(appMeshName)
		for {
			appMeshVirtualServices, err := appMeshClient.ListVirtualServices(input)
			if err != nil {
				return err
			}
			for _, appMeshVirtualService := range appMeshVirtualServices.VirtualServices {
				discoveredMeshService := a.convertAppMeshVirtualService(appMeshVirtualService, mesh)
				discoveredMeshServiceSMHNames.Insert(discoveredMeshService.GetName())
				err := a.meshServiceClient.UpsertMeshServiceSpec(ctx, discoveredMeshService)
				if err != nil {
					return err
				}
			}
			if appMeshVirtualServices.NextToken == nil {
				break
			}
			input.SetNextToken(aws.StringValue(appMeshVirtualServices.NextToken))
		}
	}
	meshServiceList, err := a.meshServiceClient.ListMeshService(ctx)
	if err != nil {
		return err
	}
	for _, meshService := range meshServiceList.Items {
		meshService := meshService
		// Skip MeshService if not owned by an AppMesh instance
		if !smhMeshNames.Has(meshService.Spec.GetMesh().GetName()) {
			continue
		}
		// Remove MeshServices whose corresponding VirtualServices no longer exist in AppMesh
		if !discoveredMeshServiceSMHNames.Has(meshService.GetName()) {
			err := a.meshServiceClient.DeleteMeshService(
				ctx,
				client.ObjectKey{Name: meshService.GetName(), Namespace: meshService.GetNamespace()})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (a *appMeshDiscoveryReconciler) reconcileMeshWorkloads(
	ctx context.Context,
	appMeshClient *appmesh.AppMesh,
	smhMeshNames sets.String,
	meshes map[string]*zephyr_discovery.Mesh,
) error {
	input := &appmesh.ListVirtualNodesInput{
		Limit: NumItemsPerRequest,
	}
	discoveredMeshWorkloadSMHNames := sets.NewString() // Set containing SMH identifiers for AppMesh MeshWorkloads for reconciling
	for appMeshName, mesh := range meshes {
		input.NextToken = nil
		input.SetMeshName(appMeshName)
		for {
			virtualNodes, err := appMeshClient.ListVirtualNodes(input)
			if err != nil {
				return err
			}
			for _, virtualNodeRef := range virtualNodes.VirtualNodes {
				discoveredMeshWorkload := a.convertAppMeshVirtualNode(virtualNodeRef, mesh)
				discoveredMeshWorkloadSMHNames.Insert(discoveredMeshWorkload.GetName())
				err = a.meshWorkloadClient.UpsertMeshWorkloadSpec(ctx, discoveredMeshWorkload)
				if err != nil {
					return err
				}
			}
			if virtualNodes.NextToken == nil {
				break
			}
			input.SetNextToken(aws.StringValue(virtualNodes.NextToken))
		}
	}
	meshWorkloadList, err := a.meshWorkloadClient.ListMeshWorkload(ctx)
	if err != nil {
		return err
	}
	for _, meshWorkload := range meshWorkloadList.Items {
		meshWorkload := meshWorkload
		// Skip MeshWorkload if not owned by an AppMesh instance
		if !smhMeshNames.Has(meshWorkload.Spec.GetMesh().GetName()) {
			continue
		}
		// Remove MeshWorkloads whose corresponding VirtualNodes no longer exist in AppMesh
		if !discoveredMeshWorkloadSMHNames.Has(meshWorkload.GetName()) {
			err := a.meshWorkloadClient.DeleteMeshWorkload(
				ctx,
				client.ObjectKey{Name: meshWorkload.GetName(), Namespace: meshWorkload.GetNamespace()})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (a *appMeshDiscoveryReconciler) convertAppMeshMesh(appMeshMesh *appmesh.MeshRef) *zephyr_discovery.Mesh {
	meshName := a.buildAppMeshMeshName(appMeshMesh)
	return &zephyr_discovery.Mesh{
		ObjectMeta: k8s_meta_types.ObjectMeta{
			Name:      meshName,
			Namespace: env.GetWriteNamespace(),
		},
		Spec: zephyr_discovery_types.MeshSpec{
			MeshType: &zephyr_discovery_types.MeshSpec_AwsAppMesh_{
				AwsAppMesh: &zephyr_discovery_types.MeshSpec_AwsAppMesh{
					Name:    aws.StringValue(appMeshMesh.MeshName),
					ApiName: a.meshPlatformName,
					Region:  Region,
				},
			},
		},
	}
}

func (a *appMeshDiscoveryReconciler) convertAppMeshVirtualService(
	appMeshVirtualService *appmesh.VirtualServiceRef,
	parentMesh *zephyr_discovery.Mesh,
) *zephyr_discovery.MeshService {
	meshServiceName := a.buildAppMeshVirtualServiceName(appMeshVirtualService)
	return &zephyr_discovery.MeshService{
		ObjectMeta: k8s_meta_types.ObjectMeta{
			Name:      meshServiceName,
			Namespace: env.GetWriteNamespace(),
		},
		Spec: zephyr_discovery_types.MeshServiceSpec{
			Mesh: &zephyr_core_types.ResourceRef{
				Name:      parentMesh.GetName(),
				Namespace: parentMesh.GetNamespace(),
				Cluster:   parentMesh.Spec.GetAwsAppMesh().GetAws,
			},
		},
	}
}

func (a *appMeshDiscoveryReconciler) convertAppMeshVirtualNode(
	appMeshVirtualNode *appmesh.VirtualNodeRef,
	parentMesh *zephyr_discovery.Mesh,
) *zephyr_discovery.MeshWorkload {
	virtualNodeName := a.buildAppMeshVirtualNodeName(appMeshVirtualNode)
	return &zephyr_discovery.MeshWorkload{
		ObjectMeta: k8s_meta_types.ObjectMeta{
			Name:      virtualNodeName,
			Namespace: env.GetWriteNamespace(),
		},
		Spec: zephyr_discovery_types.MeshWorkloadSpec{
			Mesh: &zephyr_core_types.ResourceRef{
				Name:      parentMesh.GetName(),
				Namespace: parentMesh.GetNamespace(),
				Cluster:   parentMesh.Spec.GetAwsAppMesh().GetAwsAccountName(), // TODO this should be renamed to capture the broader semantics introduced by AppMesh
			},
		},
	}
}

// TODO: https://github.com/solo-io/service-mesh-hub/issues/141
// Secret name identifies the user-supplied AWS Account registration name
// Format: <mesh_type>-<mesh_entity_name>-<parent_mesh_name>-<aws_account_registration_name>
func (a *appMeshDiscoveryReconciler) buildAppMeshMeshName(mesh *appmesh.MeshRef) string {
	return fmt.Sprintf("%s-%s-%s",
		ObjectNamePrefix,
		convertToKubeName(aws.StringValue(mesh.MeshName)),
		a.meshPlatformName)
}

func (a *appMeshDiscoveryReconciler) buildAppMeshVirtualNodeName(virtualNode *appmesh.VirtualNodeRef) string {
	return fmt.Sprintf("%s-%s-%s-%s",
		ObjectNamePrefix,
		aws.StringValue(virtualNode.VirtualNodeName),
		convertToKubeName(aws.StringValue(virtualNode.MeshName)),
		a.meshPlatformName)
}

func (a *appMeshDiscoveryReconciler) buildAppMeshVirtualServiceName(virtualService *appmesh.VirtualServiceRef) string {
	return fmt.Sprintf("%s-%s-%s-%s",
		ObjectNamePrefix,
		aws.StringValue(virtualService.VirtualServiceName),
		convertToKubeName(aws.StringValue(virtualService.MeshName)),
		a.meshPlatformName)
}

// AppMesh entity names only contain "Alphanumeric characters, dashes, and underscores are allowed." (taken from AppMesh GUI)
// So just replace underscores with a k8s name friendly delimiter
func convertToKubeName(appmeshName string) string {
	return strings.ReplaceAll(appmeshName, "_", "-")
}
