package appmesh

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/appmesh"
	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
)

//go:generate mockgen -destination=./client_builder_mock.go -source client_builder.go -package appmesh

func NewAppMeshClientBuilder(secrets gloov1.SecretClient) ClientBuilder {
	return &clientBuilder{
		secrets: secrets,
	}
}

type ClientBuilder interface {
	GetClientInstance(secretRef *core.ResourceRef, region string) (Client, error)
}

type clientBuilder struct {
	secrets gloov1.SecretClient
}

func (c *clientBuilder) GetClientInstance(secretRef *core.ResourceRef, region string) (Client, error) {
	secret, err := c.secrets.Read(secretRef.Namespace, secretRef.Name, clients.ReadOpts{})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to retrieve secret %s", secretRef.Key())
	}

	awsSecret := secret.GetAws()
	if awsSecret == nil {
		return nil, errors.Errorf("expected secret of kind Secret_Aws, but found %s", secret.GetKind())
	}

	awsSession, err := session.NewSession(aws.NewConfig().WithCredentials(
		credentials.NewStaticCredentials(awsSecret.SecretKey, awsSecret.SecretKey, "")))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create aws session with provided credentials")
	}
	appMeshApi := appmesh.New(awsSession, &aws.Config{Region: aws.String(region)})

	return &client{api: appMeshApi}, nil
}
