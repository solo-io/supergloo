package clients

import (
	k8s_core "github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/core/v1"
	"github.com/solo-io/service-mesh-hub/pkg/kube/auth"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type ClusterAuthClientFromConfigFactory func(remoteAuthConfig *rest.Config) (auth.ClusterAuthorization, error)

func ClusterAuthClientFromConfigFactoryProvider() ClusterAuthClientFromConfigFactory {
	return DefaultClusterAuthClientFromConfig
}

var DefaultClusterAuthClientFromConfig = func(remoteAuthConfig *rest.Config) (auth.ClusterAuthorization, error) {
	remoteClientset, err := k8s_core.NewClientsetFromConfig(remoteAuthConfig)
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(remoteAuthConfig)
	if err != nil {
		return nil, err
	}
	rbacClient := auth.RbacClientProvider(clientset)
	remoteAuthorityConfigCreator := auth.NewRemoteAuthorityConfigCreator(remoteClientset.Secrets(), remoteClientset.ServiceAccounts())
	remoteAuthorityManager := auth.NewRemoteAuthorityManager(remoteClientset.ServiceAccounts(), rbacClient)
	return auth.NewClusterAuthorization(remoteAuthorityConfigCreator, remoteAuthorityManager), nil
}
