package aws

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/service-mesh-hub/pkg/common/aws/aws_creds"
	"github.com/solo-io/service-mesh-hub/pkg/common/aws/clients"
	"github.com/solo-io/service-mesh-hub/pkg/common/aws/cloud"
	compute_target "github.com/solo-io/service-mesh-hub/pkg/common/compute-target"
	"github.com/solo-io/service-mesh-hub/pkg/common/kube"
	"go.uber.org/zap"
	k8s_core_types "k8s.io/api/core/v1"
)

const (
	ReconcileIntervalSeconds = 1 // TODO make this configurable by user
)

type awsCredsHandler struct {
	reconcilerCancelFuncs map[string]context.CancelFunc // Map of computeTargetName -> RestAPIDiscoveryReconciler's cancelFunc
	secretCredsConverter  aws_creds.SecretAwsCredsConverter
	reconcilers           []RestAPIDiscoveryReconciler
	stsClientFactory      clients.STSClientFactory
	awsCloudStore         cloud.AwsCloudStore
}

type AwsCredsHandler compute_target.ComputeTargetCredentialsHandler

func NewAwsAPIHandler(
	secretCredsConverter aws_creds.SecretAwsCredsConverter,
	reconcilers []RestAPIDiscoveryReconciler,
	stsClientFactory clients.STSClientFactory,
	awsCloudStore cloud.AwsCloudStore,
) AwsCredsHandler {
	return &awsCredsHandler{
		reconcilerCancelFuncs: make(map[string]context.CancelFunc),
		secretCredsConverter:  secretCredsConverter,
		reconcilers:           reconcilers,
		stsClientFactory:      stsClientFactory,
		awsCloudStore:         awsCloudStore,
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
	// Store mapping of computeTargetName to cancelFunc so associated reconcilers can be canceled
	a.reconcilerCancelFuncs[secret.GetName()] = cancelFunc
	creds, accountId, err := a.convertToCreds(secret)
	if err != nil {
		return err
	}
	// Store credentials for account.
	a.awsCloudStore.Add(accountId, creds)
	a.runReconcilers(logger, reconcilerCtx, accountId)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				a.runReconcilers(logger, reconcilerCtx, accountId)
			case <-reconcilerCtx.Done():
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
	_, accountId, err := a.convertToCreds(secret)
	if err != nil {
		return err
	}
	a.awsCloudStore.Remove(accountId)
	cancelFunc()
	return nil
}

func (a *awsCredsHandler) buildContext(parentCtx context.Context, computeTargetName string) (context.Context, context.CancelFunc) {
	return context.WithCancel(contextutils.WithLoggerValues(
		context.WithValue(parentCtx, kube.COMPUTE_TARGET, computeTargetName),
		zap.String(kube.COMPUTE_TARGET, computeTargetName),
	))
}

func (a *awsCredsHandler) runReconcilers(
	logger *zap.SugaredLogger,
	reconcilerCtx context.Context,
	accountID string,
) {
	for _, reconciler := range a.reconcilers {
		if err := reconciler.Reconcile(reconcilerCtx, accountID); err != nil {
			logger.Errorw(fmt.Sprintf("Error during reconcile for %s", reconciler.GetName()), zap.Error(err))
		}
	}
}

func (a *awsCredsHandler) fetchAWSAccount(creds *credentials.Credentials) (string, error) {
	// Region does not matter for constructing the STS client because the account identity is region agnostic.
	stsClient, err := a.stsClientFactory(creds, "us-east-1")
	if err != nil {
		return "", err
	}
	callerIdentity, err := stsClient.GetCallerIdentity()
	if err != nil {
		return "", err
	}
	return aws.StringValue(callerIdentity.Account), nil
}

func (a *awsCredsHandler) convertToCreds(secret *k8s_core_types.Secret) (*credentials.Credentials, string, error) {
	creds, err := a.secretCredsConverter.SecretToCreds(secret)
	if err != nil {
		return nil, "", err
	}
	accountID, err := a.fetchAWSAccount(creds)
	if err != nil {
		return nil, "", err
	}
	return creds, accountID, nil
}
