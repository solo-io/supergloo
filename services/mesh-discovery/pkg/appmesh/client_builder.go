package appmesh

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/appmesh"
	"github.com/solo-io/go-utils/errors"
)

func NewAppMeshClientBuilder() ClientBuilder {
	return &clientBuilder{}
}

type ClientBuilder interface {
	GetClientInstance(region string) (Client, error)
}

type clientBuilder struct {
}

func (c *clientBuilder) GetClientInstance(region string) (Client, error) {
	chainedProvider := credentials.NewChainCredentials([]credentials.Provider{
		&credentials.SharedCredentialsProvider{},
		&credentials.EnvProvider{},
	})
	awsSession, err := session.NewSession(aws.NewConfig().WithCredentials(chainedProvider))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create aws session with provided credentials")
	}
	appMeshApi := appmesh.New(awsSession, &aws.Config{Region: aws.String(region)})

	return &client{api: appMeshApi}, nil
}
