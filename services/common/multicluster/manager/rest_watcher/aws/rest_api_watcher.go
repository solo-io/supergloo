package aws

import (
	"github.com/rotisserie/eris"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common/aws_creds"
	k8s_core_types "k8s.io/api/core/v1"
)

// Thread-safe map of REST API name to AppMesh client
//type AppMeshClientMap struct {
//	clients sync.map[string]*appmesh.AppMesh
//}

var (
	AppMeshClientNotFound = func(apiName string) error {
		return eris.Errorf("AppMesh client not found for API: %s", apiName)
	}
)

type awsWatcher struct {
	secretAwsCredsConvert aws_creds.SecretAwsCredsConverter
}

func (a *awsWatcher) RestAPIAdded(secret *k8s_core_types.Secret, apiName string) error {

	return nil
}

func (a *awsWatcher) RestAPIRemoved(apiName string) error {
	panic("implement me")
}
