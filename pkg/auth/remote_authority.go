package auth

import (
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	k8sapiv1 "k8s.io/api/core/v1"
	rbactypes "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

// Create a service account on a cluster that `targetClusterCfg` can reach
// Set up that service account with the indicated cluster roles
//go:generate mockgen -destination ./mocks/mock_remote_authority_manager.go github.com/solo-io/mesh-projects/pkg/auth RemoteAuthorityManager
type RemoteAuthorityManager interface {
	// creates a new service account in the cluster pointed to by the cfg at the name/namespace indicated by the ResourceRef,
	// and assigns it the given ClusterRoles
	// NB: if role assignment fails, the service account is left in the cluster; this is not an atomic operation
	CreateRemoteServiceAccount(targetClusterCfg *rest.Config, newServiceAccountRef *core.ResourceRef, roles []*rbactypes.ClusterRole) (*k8sapiv1.ServiceAccount, error)
}

func NewRemoteAuthorityManager(clientFactory ClientFactory) RemoteAuthorityManager {
	return &remoteAuthorityManager{clientFactory}
}

type remoteAuthorityManager struct {
	clientFactory ClientFactory
}

func (r *remoteAuthorityManager) CreateRemoteServiceAccount(targetClusterCfg *rest.Config, newServiceAccountRef *core.ResourceRef, roles []*rbactypes.ClusterRole) (*k8sapiv1.ServiceAccount, error) {
	clients, err := r.clientFactory(targetClusterCfg, newServiceAccountRef.Namespace)
	if err != nil {
		return nil, err
	}

	newServiceAccount, err := clients.ServiceAccountClient.Create(&k8sapiv1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name: newServiceAccountRef.Name,
		},
	})

	if err != nil {
		return nil, err
	}

	err = clients.RbacClient.BindClusterRolesToServiceAccount(newServiceAccount, roles)
	if err != nil {
		return nil, err
	}

	return newServiceAccount, nil
}
