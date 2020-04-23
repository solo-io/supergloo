package aws

import (
	"context"
	"time"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common/aws_creds"
	"github.com/solo-io/service-mesh-hub/services/common/constants"
	rest_api "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/rest-api"
	appmesh_client "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/rest-api/aws/appmesh-client"
	"go.uber.org/zap"
	k8s_core_types "k8s.io/api/core/v1"
)

const (
	// TODO make this configurable by user
	ReconcileIntervalSeconds = 1
)

type awsCredsHandler struct {
	appMeshClientFactory              appmesh_client.AppMeshClientFactory
	appMeshDiscoveryReconcilerFactory AppMeshDiscoveryReconcilerFactory
	reconcilerCancelFuncs             map[string]context.CancelFunc // Map of meshPlatformName -> RestAPIDiscoveryReconciler's cancelFunc
}

type AwsCredsHandler rest_api.RestAPICredsHandler

func NewAwsCredsHandler(
	appMeshClientFactory appmesh_client.AppMeshClientFactory,
	appMeshReconcilerFactory AppMeshDiscoveryReconcilerFactory,
) AwsCredsHandler {
	return &awsCredsHandler{
		appMeshClientFactory:              appMeshClientFactory,
		appMeshDiscoveryReconcilerFactory: appMeshReconcilerFactory,
		reconcilerCancelFuncs:             make(map[string]context.CancelFunc),
	}
}

func (a *awsCredsHandler) RestAPIAdded(ctx context.Context, secret *k8s_core_types.Secret) error {
	// Only handle AWS REST APIs
	if secret.Type != aws_creds.AWSSecretType {
		return nil
	}
	logger := contextutils.LoggerFrom(ctx)
	logger.Debugf("New REST API added for meshPlatform %s", secret.GetName())
	// Periodically run API Reconciler to ensure AppMesh state is consistent with SMH
	ticker := time.NewTicker(ReconcileIntervalSeconds * time.Second)
	appMeshClient, err := a.appMeshClientFactory.Build(secret, Region)
	if err != nil {
		return err
	}
	reconcilerCtx, cancelFunc := a.buildContext(ctx, secret.GetName())
	appMeshDiscoveryReconciler := a.appMeshDiscoveryReconcilerFactory(secret.GetName(), appMeshClient)
	// Store mapping of meshPlatformName to cancelFunc so reconciler can be canceled
	a.reconcilerCancelFuncs[secret.GetName()] = cancelFunc
	go func() {
		for {
			select {
			case <-ticker.C:
				logger.Debugf("Reconciling AppMesh with secret %s.%s", secret.GetName(), secret.GetNamespace())
				err := appMeshDiscoveryReconciler.Reconcile(reconcilerCtx)
				if err != nil {
					logger.Error(err)
				}
			case <-reconcilerCtx.Done():
				ticker.Stop()
				return
			}
		}
	}()
	return nil
}

func (a *awsCredsHandler) RestAPIRemoved(ctx context.Context, secret *k8s_core_types.Secret) error {
	// Only handle AWS REST APIs
	if secret.Type != aws_creds.AWSSecretType {
		return nil
	}
	logger := contextutils.LoggerFrom(ctx)
	logger.Debugf("REST API removed for meshPlatform %s", secret)
	cancelFunc, ok := a.reconcilerCancelFuncs[secret.GetName()]
	if !ok {
		logger.Errorf("Error stopping RestAPIDiscoveryReconciler for meshPlatform %s, entry not found in map.", secret)
	}
	cancelFunc()
	return nil
}

func (a *awsCredsHandler) buildContext(parentCtx context.Context, meshPlatformName string) (context.Context, context.CancelFunc) {
	return context.WithCancel(contextutils.WithLoggerValues(
		context.WithValue(parentCtx, constants.MESH_PLATFORM, meshPlatformName),
		zap.String(constants.MESH_PLATFORM, meshPlatformName),
	))
}
