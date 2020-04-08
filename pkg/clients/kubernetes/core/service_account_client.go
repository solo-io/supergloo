package kubernetes_core

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewGeneratedServiceAccountClient(client kubernetes.Interface) ServiceAccountClient {
	return &serviceAccountClient{client: client.CoreV1()}
}

type serviceAccountClient struct {
	client v1.ServiceAccountsGetter
}

func (s serviceAccountClient) Create(_ context.Context, serviceAccount *corev1.ServiceAccount) error {
	update, err := s.client.ServiceAccounts(serviceAccount.GetNamespace()).Create(serviceAccount)
	if err != nil {
		return err
	}
	*serviceAccount = *update
	return nil
}

func (s serviceAccountClient) Get(_ context.Context, name, namespace string) (*corev1.ServiceAccount, error) {
	return s.client.ServiceAccounts(namespace).Get(name, metav1.GetOptions{})
}

func (s serviceAccountClient) Update(_ context.Context, serviceAccount *corev1.ServiceAccount) error {
	update, err := s.client.ServiceAccounts(serviceAccount.GetNamespace()).Update(serviceAccount)
	if err != nil {
		return err
	}
	*serviceAccount = *update
	return nil
}

func (s *serviceAccountClient) List(_ context.Context, options ...client.ListOption) (*corev1.ServiceAccountList, error) {
	listOptions := &client.ListOptions{}
	for _, v := range options {
		v.ApplyToList(listOptions)
	}
	raw := metav1.ListOptions{}
	if converted := listOptions.AsListOptions(); converted != nil {
		raw = *converted
	}
	return s.client.ServiceAccounts(listOptions.Namespace).List(raw)
}
