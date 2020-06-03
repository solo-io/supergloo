package aws

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/service-mesh-hub/pkg/aws/aws_creds"
	"github.com/solo-io/service-mesh-hub/pkg/aws/clients"
	credentials2 "github.com/solo-io/service-mesh-hub/pkg/aws/credentials"
	aws2 "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/compute-target/aws"
	v1 "k8s.io/api/core/v1"
)

type networkingAwsCredsHandler struct {
	stsClientFactory     clients.STSClientFactory
	credentialsMap       credentials2.AwsCredentialsGetter
	secretCredsConverter aws_creds.SecretAwsCredsConverter
}

func NewNetworkingAwsCredsHandler(
	credentialsGetter credentials2.AwsCredentialsGetter,
	stsClientFactory clients.STSClientFactory,
	secretCredsConverter aws_creds.SecretAwsCredsConverter,
) aws2.AwsCredsHandler {
	return &networkingAwsCredsHandler{
		credentialsMap:       credentialsGetter,
		stsClientFactory:     stsClientFactory,
		secretCredsConverter: secretCredsConverter,
	}
}

func (n *networkingAwsCredsHandler) ComputeTargetAdded(ctx context.Context, secret *v1.Secret) error {
	// Only handle AWS REST APIs
	if secret.Type != aws_creds.AWSSecretType {
		return nil
	}
	logger := contextutils.LoggerFrom(ctx)
	logger.Debugf("New REST API added for compute target %s", secret.GetName())
	creds, err := n.secretCredsConverter.SecretToCreds(secret)
	if err != nil {
		return err
	}
	accountID, err := n.getAccountID(ctx, creds)
	if err != nil {
		return err
	}
	n.credentialsMap.Set(accountID, creds)
	return nil
}

func (n *networkingAwsCredsHandler) ComputeTargetRemoved(ctx context.Context, secret *v1.Secret) error {
	// Only handle AWS REST APIs
	if secret.Type != aws_creds.AWSSecretType {
		return nil
	}
	logger := contextutils.LoggerFrom(ctx)
	logger.Debugf("REST API removed for compute target %s", secret)
	creds, err := n.secretCredsConverter.SecretToCreds(secret)
	if err != nil {
		return err
	}
	accountID, err := n.getAccountID(ctx, creds)
	if err != nil {
		return err
	}
	n.credentialsMap.Remove(accountID)
	return nil
}

func (n *networkingAwsCredsHandler) getAccountID(ctx context.Context, creds *credentials.Credentials) (string, error) {
	// Region does not matter for constructing the STS client because the account identity is region agnostic.
	stsClient, err := n.stsClientFactory(creds, "us-east-1")
	if err != nil {
		return "", err
	}
	callerIdentity, err := stsClient.GetCallerIdentity()
	if err != nil {
		return "", err
	}
	return aws.StringValue(callerIdentity.Account), nil
}
