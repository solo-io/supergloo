package aws

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/appmesh"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common/aws_creds"
	zephyr_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/env"
	rest_api "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/rest-api"
	k8s_core_types "k8s.io/api/core/v1"
	k8s_meta_types "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	ObjectNamePrefix   = "appmesh"
	NumItemsPerRequest = aws.Int64(100)
	Region             = "us-east-2" // TODO remove hardcode and replace with configuration
)

type appMeshAPIReconciler struct {
	secretAwsCredsConverter aws_creds.SecretAwsCredsConverter
	secret                  *k8s_core_types.Secret
	meshClient              zephyr_discovery.MeshClient
	meshWorkloadClient      zephyr_discovery.MeshWorkloadClient
	meshServiceClient       zephyr_discovery.MeshServiceClient
}

type AppMeshReconcilerFactory func(secret *k8s_core_types.Secret) rest_api.RestAPIReconciler

func NewAppMeshReconcilerFactory(
	secretAwsCredsConverter aws_creds.SecretAwsCredsConverter,
	masterClient client.Client,
	meshClientFactory zephyr_discovery.MeshClientFactory,
	meshWorkloadClientFactory zephyr_discovery.MeshWorkloadClientFactory,
	meshServiceClientFactory zephyr_discovery.MeshServiceClientFactory,
) AppMeshReconcilerFactory {
	return func(
		secret *k8s_core_types.Secret,
	) rest_api.RestAPIReconciler {
		return NewAppMeshReconciler(
			secretAwsCredsConverter,
			meshClientFactory(masterClient),
			meshWorkloadClientFactory(masterClient),
			meshServiceClientFactory(masterClient),
			secret,
		)
	}
}

func NewAppMeshReconciler(
	secretAwsCredsConverter aws_creds.SecretAwsCredsConverter,
	meshClient zephyr_discovery.MeshClient,
	meshWorkloadClient zephyr_discovery.MeshWorkloadClient,
	meshServiceClient zephyr_discovery.MeshServiceClient,
	secret *k8s_core_types.Secret,
) rest_api.RestAPIReconciler {
	return &appMeshAPIReconciler{
		secretAwsCredsConverter: secretAwsCredsConverter,
		meshClient:              meshClient,
		meshWorkloadClient:      meshWorkloadClient,
		meshServiceClient:       meshServiceClient,
		secret:                  secret,
	}
}

func (a *appMeshAPIReconciler) Reconcile(ctx context.Context) error {
	var err error
	var appMeshClient *appmesh.AppMesh
	if appMeshClient, err = a.buildAppMeshClient(a.secret); err != nil {
		return err
	}
	meshes, smhMeshNames, err := a.reconcileMeshes(ctx, appMeshClient)
	if err != nil {
		return err
	}
	if err = a.reconcileMeshServices(ctx, appMeshClient, smhMeshNames, meshes); err != nil {
		return err
	}
	if err = a.reconcileMeshWorkloads(ctx, appMeshClient, smhMeshNames, meshes); err != nil {
		return err
	}
	return nil
}

func (a *appMeshAPIReconciler) reconcileMeshes(
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

func (a *appMeshAPIReconciler) reconcileMeshServices(
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

func (a *appMeshAPIReconciler) reconcileMeshWorkloads(
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

func (a *appMeshAPIReconciler) buildAppMeshClient(secret *k8s_core_types.Secret) (*appmesh.AppMesh, error) {
	creds, err := a.secretAwsCredsConverter.SecretToCreds(secret)
	if err != nil {
		return nil, err
	}
	sess, err := session.NewSession(&aws.Config{
		Credentials: creds,
		Region:      aws.String(Region),
	})
	if err != nil {
		return nil, err
	}
	return appmesh.New(sess), nil
}

func (a *appMeshAPIReconciler) convertAppMeshMesh(appMeshMesh *appmesh.MeshRef) *zephyr_discovery.Mesh {
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
					ApiName: a.secret.GetName(),
					Region:  Region,
				},
			},
		},
	}
}

func (a *appMeshAPIReconciler) convertAppMeshVirtualService(
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
				Cluster:   parentMesh.Spec.GetAwsAppMesh().GetApiName(),
			},
		},
	}
}

func (a *appMeshAPIReconciler) convertAppMeshVirtualNode(
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
				Cluster:   parentMesh.Spec.GetAwsAppMesh().GetApiName(), // TODO this should be renamed to capture the broader semantics introduced by AppMesh
			},
		},
	}
}

// TODO: https://github.com/solo-io/service-mesh-hub/issues/141
// Secret name identifies the user-supplied AWS Account registration name
// Format: <mesh_type>-<mesh_entity_name>-<parent_mesh_name>-<aws_account_registration_name>
func (a *appMeshAPIReconciler) buildAppMeshMeshName(mesh *appmesh.MeshRef) string {
	return strings.Join([]string{
		ObjectNamePrefix,
		aws.StringValue(mesh.MeshName),
		a.secret.GetName()},
		"-")
}

func (a *appMeshAPIReconciler) buildAppMeshVirtualNodeName(virtualNode *appmesh.VirtualNodeRef) string {
	return strings.Join([]string{
		ObjectNamePrefix,
		aws.StringValue(virtualNode.VirtualNodeName),
		aws.StringValue(virtualNode.MeshName),
		a.secret.GetName()},
		"-")
}

func (a *appMeshAPIReconciler) buildAppMeshVirtualServiceName(virtualService *appmesh.VirtualServiceRef) string {
	return strings.Join([]string{
		ObjectNamePrefix,
		aws.StringValue(virtualService.VirtualServiceName),
		aws.StringValue(virtualService.MeshName),
		a.secret.GetName()},
		"-")
}
