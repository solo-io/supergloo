package multicluster

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/appmesh"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/k8s_secrets/aws_creds"
	"github.com/solo-io/service-mesh-hub/services/common/multicluster/manager/rest_watcher"
	k8s_core_types "k8s.io/api/core/v1"
)

type appMeshAPIReconciler struct {
	secretAwsCredsConverter aws_creds.SecretAwsCredsConverter
	secret                  *k8s_core_types.Secret
	meshClient              zephyr_discovery.MeshClient
	meshWorkloadClient      zephyr_discovery.MeshWorkloadClient
	meshServiceClient       zephyr_discovery.MeshServiceClient
}

type AppMeshAPIReconcilerFactory func(secret *k8s_core_types.Secret) rest_watcher.RestAPIReconciler

func NewAppMeshAPIReconcilerFactory(
	secretAwsCredsConverter aws_creds.SecretAwsCredsConverter,
	meshClient zephyr_discovery.MeshClient,
	meshWorkloadClient zephyr_discovery.MeshWorkloadClient,
	meshServiceClient zephyr_discovery.MeshServiceClient,
) AppMeshAPIReconcilerFactory {
	return func(secret *k8s_core_types.Secret) rest_watcher.RestAPIReconciler {
		return NewAppMeshAPIReconciler(
			secretAwsCredsConverter,
			meshClient,
			meshWorkloadClient,
			meshServiceClient,
			secret,
		)
	}
}

func NewAppMeshAPIReconciler(
	secretAwsCredsConverter aws_creds.SecretAwsCredsConverter,
	meshClient zephyr_discovery.MeshClient,
	meshWorkloadClient zephyr_discovery.MeshWorkloadClient,
	meshServiceClient zephyr_discovery.MeshServiceClient,
	secret *k8s_core_types.Secret,
) rest_watcher.RestAPIReconciler {
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
	if err = a.reconcileMeshes(ctx, appMeshClient); err != nil {
		return err
	}
	if err = a.reconcileMeshWorkloads(appMeshClient); err != nil {
		return err
	}
	if err = a.reconcileMeshServices(appMeshClient); err != nil {
		return err
	}
	return nil
}

func (a *appMeshAPIReconciler) reconcileMeshes(ctx context.Context, appMeshClient *appmesh.AppMesh) error {
	//appMeshClient.ListMeshes()

	meshList, err := a.meshClient.ListMesh(ctx)
	if err != nil {
		return err
	}
	for _, mesh := range meshList.Items {
		mesh := mesh
		if mesh.Spec.GetAwsAppMesh() == nil {
			continue
		}
		// SMH meshes
	}
	return nil
}

func (a *appMeshAPIReconciler) reconcileMeshWorkloads(appMeshClient *appmesh.AppMesh) error {
	return nil
}

func (a *appMeshAPIReconciler) reconcileMeshServices(appMeshClient *appmesh.AppMesh) error {
	return nil
}

func (a *appMeshAPIReconciler) buildAppMeshClient(secret *k8s_core_types.Secret) (*appmesh.AppMesh, error) {
	creds, err := a.secretAwsCredsConverter.SecretToCreds(secret)
	if err != nil {
		return nil, err
	}
	sess, err := session.NewSession(&aws.Config{
		Credentials: creds,
	})
	if err != nil {
		return nil, err
	}
	return appmesh.New(sess), nil
}
