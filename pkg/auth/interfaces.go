package auth

import (
	k8sapiv1 "k8s.io/api/core/v1"
	rbactypes "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sclientv1 "k8s.io/client-go/kubernetes/typed/core/v1"
	k8srbacv1 "k8s.io/client-go/kubernetes/typed/rbac/v1"
	"k8s.io/client-go/rest"
)

//go:generate mockgen -destination ./mocks/mock_service_account_client.go github.com/solo-io/mesh-projects/pkg/auth ServiceAccountClient
type ServiceAccountClient interface {
	Create(*k8sapiv1.ServiceAccount) (*k8sapiv1.ServiceAccount, error)
	Get(name string, options metav1.GetOptions) (*k8sapiv1.ServiceAccount, error)
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

// given a config for a cluster and a namespace where resources should be written, create clients capable of writing there
type ClientFactory func(cfg *rest.Config, writeNamespace string) (*Clients, error)

func DefaultClients(cfg *rest.Config, writeNamespace string) (*Clients, error) {
	k8sClient, err := k8sclientv1.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	serviceAccountClient := k8sClient.ServiceAccounts(writeNamespace)

	k8sRbacClient, err := k8srbacv1.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}
	rbacClient := &defaultRbacClient{
		clusterRoleBindingClient: k8sRbacClient.ClusterRoleBindings(),
	}

	secretClient := k8sClient.Secrets(writeNamespace)

	return &Clients{serviceAccountClient, rbacClient, secretClient}, nil
}
