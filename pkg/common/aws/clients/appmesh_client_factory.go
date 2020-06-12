package clients

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/appmesh"
	"github.com/aws/aws-sdk-go/service/appmesh/appmeshiface"
)

type AppmeshRawClientFactory func(sess *session.Session) (appmeshiface.AppMeshAPI, error)

func AppmeshRawClientFactoryProvider() AppmeshRawClientFactory {
	return func(sess *session.Session) (appmeshiface.AppMeshAPI, error) {
		return appmesh.New(sess), nil
	}
}
