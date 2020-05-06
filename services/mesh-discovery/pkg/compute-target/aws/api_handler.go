package aws

import (
	"context"
	"time"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common/aws_creds"
	compute_target "github.com/solo-io/service-mesh-hub/services/common/compute-target"
	"github.com/solo-io/service-mesh-hub/services/common/constants"
	"go.uber.org/zap"
	k8s_core_types "k8s.io/api/core/v1"
)

const (
	// TODO make this configurable by user
	ReconcileIntervalSeconds = 1
	Region                   = "us-east-2" // TODO remove hardcode and replace with configuration
)

type awsCredsHandler struct {
	reconcilerCancelFuncs      map[string]context.CancelFunc // Map of computeTargetName -> RestAPIDiscoveryReconciler's cancelFunc
	secretCredsConverter       aws_creds.SecretAwsCredsConverter
	appMeshDiscoveryReconciler AppMeshDiscoveryReconciler
	eksReconciler              EksDiscoveryReconciler
}

type AwsCredsHandler compute_target.ComputeTargetCredentialsHandler

func NewAwsAPIHandler(
	secretCredsConverter aws_creds.SecretAwsCredsConverter,
	appMeshDiscoveryReconciler AppMeshDiscoveryReconciler,
	eksReconciler EksDiscoveryReconciler,
) AwsCredsHandler {
	return &awsCredsHandler{
		reconcilerCancelFuncs:      make(map[string]context.CancelFunc),
		secretCredsConverter:       secretCredsConverter,
		appMeshDiscoveryReconciler: appMeshDiscoveryReconciler,
		eksReconciler:              eksReconciler,
	}
}

func (a *awsCredsHandler) ComputeTargetAdded(ctx context.Context, secret *k8s_core_types.Secret) error {
	// Only handle AWS REST APIs
	if secret.Type != aws_creds.AWSSecretType {
		return nil
	}
	logger := contextutils.LoggerFrom(ctx)
	logger.Debugf("New REST API added for compute target %s", secret.GetName())
	ticker := time.NewTicker(ReconcileIntervalSeconds * time.Second)
	reconcilerCtx, cancelFunc := a.buildContext(ctx, secret.GetName())
	// Store mapping of computeTargetName to cancelFunc so reconciler can be canceled
	a.reconcilerCancelFuncs[secret.GetName()] = cancelFunc
	creds, err := a.secretCredsConverter.SecretToCreds(secret)
	if err != nil {
		return err
	}
	go func() {
		for {
			select {
			case <-ticker.C:
				var err error
				logger.Debugf("Reconciling AppMesh with secret %s.%s", secret.GetName(), secret.GetNamespace())
				if err = a.appMeshDiscoveryReconciler.Reconcile(reconcilerCtx, creds, Region); err != nil {
					logger.Error(err)
				}
				if err = a.eksReconciler.Reconcile(reconcilerCtx, creds, Region); err != nil {
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

func (a *awsCredsHandler) ComputeTargetRemoved(ctx context.Context, secret *k8s_core_types.Secret) error {
	// Only handle AWS REST APIs
	if secret.Type != aws_creds.AWSSecretType {
		return nil
	}
	logger := contextutils.LoggerFrom(ctx)
	logger.Debugf("REST API removed for compute target %s", secret)
	cancelFunc, ok := a.reconcilerCancelFuncs[secret.GetName()]
	if !ok {
		logger.Errorf("Error stopping RestAPIDiscoveryReconciler for compute target %s, entry not found in map.", secret)
	}
	cancelFunc()
	return nil
}

func (a *awsCredsHandler) buildContext(parentCtx context.Context, computeTargetName string) (context.Context, context.CancelFunc) {
	return context.WithCancel(contextutils.WithLoggerValues(
		context.WithValue(parentCtx, constants.COMPUTE_TARGET, computeTargetName),
		zap.String(constants.COMPUTE_TARGET, computeTargetName),
	))
}
