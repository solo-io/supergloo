package auth

import (
	k8s_core "github.com/solo-io/external-apis/pkg/api/k8s/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type ClusterAuthClientFromConfigFactory func(remoteAuthConfig *rest.Config) (ClusterAuthorization, error)

func ClusterAuthClientFromConfigFactoryProvider() ClusterAuthClientFromConfigFactory {
	return DefaultClusterAuthClientFromConfig
}

var DefaultClusterAuthClientFromConfig = func(remoteAuthConfig *rest.Config) (ClusterAuthorization, error) {
	remoteClientset, err := k8s_core.NewClientsetFromConfig(remoteAuthConfig)
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(remoteAuthConfig)
	if err != nil {
		return nil, err
	}
	rbacClient := RbacClientProvider(clientset)
	remoteAuthorityConfigCreator := NewRemoteAuthorityConfigCreator(remoteClientset.Secrets(), remoteClientset.ServiceAccounts())
	remoteAuthorityManager := NewRemoteAuthorityManager(remoteClientset.ServiceAccounts(), rbacClient)
	return NewClusterAuthorization(remoteAuthorityConfigCreator, remoteAuthorityManager), nil
}
