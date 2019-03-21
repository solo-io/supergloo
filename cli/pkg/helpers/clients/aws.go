package clients

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/appmesh"
	"github.com/solo-io/go-utils/errors"
)

//go:generate mockgen -destination=./../mocks/aws.go -source aws.go -package mocks

// Functions must have the same signatures at the ones in the embedded appmesh.AppMesh struct.
type Appmesh interface {
	ListMeshes(input *appmesh.ListMeshesInput) (*appmesh.ListMeshesOutput, error)
}

// We wrap the AWS App Mesh client in order to be able to define an interface (with only the function we need)
// that we can use to generate mocks.
type AppmeshClient struct {
	*appmesh.AppMesh
}

var mock Appmesh

func UseAppmeshMock(mockClient Appmesh) {
	mock = mockClient
}

func NewAppmeshClient(accessKeyId, secretAccessKey, region string) (Appmesh, error) {
	if mock != nil {
		return mock, nil
	}
	awsSession, err := session.NewSession(aws.NewConfig().WithCredentials(
		credentials.NewStaticCredentials(accessKeyId, secretAccessKey, "")))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create aws session with provided credentials")
	}
	return appmesh.New(awsSession, &aws.Config{Region: aws.String(region)}), nil
}
