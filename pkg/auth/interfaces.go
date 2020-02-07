package auth

import (
	k8sapiv1 "k8s.io/api/core/v1"
	rbactypes "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//go:generate mockgen -destination ./mocks/mock_service_account_client.go github.com/solo-io/mesh-projects/pkg/auth ServiceAccountClient
type ServiceAccountClient interface {
	Create(*k8sapiv1.ServiceAccount) (*k8sapiv1.ServiceAccount, error)
	Get(name string, options metav1.GetOptions) (*k8sapiv1.ServiceAccount, error)
	Update(*k8sapiv1.ServiceAccount) (*k8sapiv1.ServiceAccount, error)
}

//go:generate mockgen -destination ./mocks/mock_secret_client.go github.com/solo-io/mesh-projects/pkg/auth SecretClient
type SecretClient interface {
	Get(name string, options metav1.GetOptions) (*k8sapiv1.Secret, error)
}

//go:generate mockgen -destination ./mocks/mock_rbac_client.go github.com/solo-io/mesh-projects/pkg/auth RbacClient
type RbacClient interface {
	// bind the given roles to the target service account at cluster scope
	BindClusterRolesToServiceAccount(targetServiceAccount *k8sapiv1.ServiceAccount, roles []*rbactypes.ClusterRole) error
}

type Clients struct {
	ServiceAccountClient ServiceAccountClient
	RbacClient           RbacClient
	SecretClient         SecretClient
}
