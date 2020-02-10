package auth

import (
	k8sapiv1 "k8s.io/api/core/v1"
	rbactypes "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

//go:generate mockgen -source ./interfaces.go -destination ./mocks/mock_auth.go

type ServiceAccountClient interface {
	// create the service account in the namespace on the resource's ObjectMeta
	Create(serviceAccount *k8sapiv1.ServiceAccount) (*k8sapiv1.ServiceAccount, error)

	Get(namespace, name string, options metav1.GetOptions) (*k8sapiv1.ServiceAccount, error)

	// update the service account in the namespace on the resource's ObjectMeta
	Update(serviceAccount *k8sapiv1.ServiceAccount) (*k8sapiv1.ServiceAccount, error)
}

func NewServiceAccountClient(k kubernetes.Interface) ServiceAccountClient {
	return &serviceAccountClient{coreV1: k.CoreV1()}
}

type serviceAccountClient struct {
	coreV1 corev1.CoreV1Interface
}

func (s *serviceAccountClient) Create(serviceAccount *k8sapiv1.ServiceAccount) (*k8sapiv1.ServiceAccount, error) {
	return s.coreV1.ServiceAccounts(serviceAccount.GetNamespace()).Create(serviceAccount)
}

func (s *serviceAccountClient) Get(namespace, name string, options metav1.GetOptions) (*k8sapiv1.ServiceAccount, error) {
	return s.coreV1.ServiceAccounts(namespace).Get(name, options)
}

func (s *serviceAccountClient) Update(serviceAccount *k8sapiv1.ServiceAccount) (*k8sapiv1.ServiceAccount, error) {
	return s.coreV1.ServiceAccounts(serviceAccount.GetNamespace()).Update(serviceAccount)
}

type SecretClient interface {
	Get(namespace, name string, options metav1.GetOptions) (*k8sapiv1.Secret, error)
}

func NewSecretClient(k kubernetes.Interface) SecretClient {
	return &secretClient{coreV1: k.CoreV1()}
}

type secretClient struct {
	coreV1 corev1.CoreV1Interface
}

func (s *secretClient) Get(namespace, name string, options metav1.GetOptions) (*k8sapiv1.Secret, error) {
	return s.coreV1.Secrets(namespace).Get(name, options)
}

type RbacClient interface {
	// bind the given roles to the target service account at cluster scope
	BindClusterRolesToServiceAccount(targetServiceAccount *k8sapiv1.ServiceAccount, roles []*rbactypes.ClusterRole) error
}

type Clients struct {
	ServiceAccountClient ServiceAccountClient
	RbacClient           RbacClient
	SecretClient         SecretClient
}
