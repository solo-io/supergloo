package appmesh

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/appmesh"
	"github.com/aws/aws-sdk-go/service/appmesh/appmeshiface"
)

type AppmeshRawClientFactory func(creds *credentials.Credentials, region string) (appmeshiface.AppMeshAPI, error)

func AppmeshRawClientFactoryProvider() AppmeshRawClientFactory {
	return func(creds *credentials.Credentials, region string) (appmeshiface.AppMeshAPI, error) {
		sess, err := session.NewSession(&aws.Config{
			Credentials: creds,
			Region:      aws.String(region),
		})
		if err != nil {
			return nil, err
		}
		return appmesh.New(sess), nil
	}
}
