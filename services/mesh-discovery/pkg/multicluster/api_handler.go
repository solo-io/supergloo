package multicluster

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/appmesh"
	"github.com/solo-io/service-mesh-hub/pkg/k8s_secrets/aws_creds"
	"github.com/solo-io/service-mesh-hub/services/common/multicluster/manager/rest_manager"
	k8s_core_types "k8s.io/api/core/v1"
)

type apiHandler struct {
	secretAwsCredsConvert aws_creds.SecretAwsCredsConverter
}

func NewRestAPIHandler(secretAwsCredsConvert aws_creds.SecretAwsCredsConverter) rest_manager.RestAPIHandler {
	return &apiHandler{secretAwsCredsConvert: secretAwsCredsConvert}
}

func (a *apiHandler) APIAdded(ctx context.Context, apiProvider rest_manager.RestAPIProvider) error {
	err := apiProvider.IsProviderValid()
	if err != nil {
		return err
	}
	switch apiProvider {
	case rest_manager.AppMesh:

	}
	return nil
}

func (a *apiHandler) APIRemoved(ctx context.Context, apiProvider rest_manager.RestAPIProvider) error {
	panic("implement me")
}

func (a *apiHandler) initAppMeshClient(secret *k8s_core_types.Secret) (*appmesh.AppMesh, error) {
	creds, err := a.secretAwsCredsConvert.SecretToCreds(secret)
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
