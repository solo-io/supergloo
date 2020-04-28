package appmesh

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/appmesh"
	"github.com/aws/aws-sdk-go/service/appmesh/appmeshiface"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common/aws_creds"
	v1 "k8s.io/api/core/v1"
)

type appMeshClientFactory struct {
	secretConverter aws_creds.SecretAwsCredsConverter
}

func NewAppMeshClientFactory(secretConverter aws_creds.SecretAwsCredsConverter) AppMeshClientFactory {
	return &appMeshClientFactory{secretConverter: secretConverter}
}

func (a *appMeshClientFactory) Build(secret *v1.Secret, region string) (appmeshiface.AppMeshAPI, error) {
	creds, err := a.secretConverter.SecretToCreds(secret)
	if err != nil {
		return nil, err
	}
	sess, err := session.NewSession(&aws.Config{
		Credentials: creds,
		Region:      aws.String(region),
	})
	if err != nil {
		return nil, err
	}
	return appmesh.New(sess), nil
}
