package kubernetes_core

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

func NewGeneratedServiceAccountClient(client kubernetes.Interface) ServiceAccountClient {
	return &sercviceAccountClient{client: client.CoreV1()}
}

type sercviceAccountClient struct {
	client v1.ServiceAccountsGetter
}

func (s sercviceAccountClient) Create(_ context.Context, serviceAccount *corev1.ServiceAccount) error {
	update, err := s.client.ServiceAccounts(serviceAccount.GetNamespace()).Create(serviceAccount)
	if err != nil {
		return err
	}
	*serviceAccount = *update
	return nil
}

func (s sercviceAccountClient) Get(_ context.Context, name, namespace string) (*corev1.ServiceAccount, error) {
	return s.client.ServiceAccounts(namespace).Get(name, metav1.GetOptions{})
}

func (s sercviceAccountClient) Update(_ context.Context, serviceAccount *corev1.ServiceAccount) error {
	update, err := s.client.ServiceAccounts(serviceAccount.GetNamespace()).Update(serviceAccount)
	if err != nil {
		return err
	}
	*serviceAccount = *update
	return nil
}
