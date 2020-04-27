package appmesh

import (
	"github.com/aws/aws-sdk-go/service/appmesh/appmeshiface"
	k8s_core_types "k8s.io/api/core/v1"
)

//go:generate mockgen -source interfaces.go -destination ./mocks/interfaces.go

type AppMeshClientFactory interface {
	Build(secret *k8s_core_types.Secret, region string) (appmeshiface.AppMeshAPI, error)
}
