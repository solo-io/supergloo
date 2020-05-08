package kubeconfig

import (
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

//go:generate mockgen -source ./interfaces.go -destination ./mocks/mock_interfaces.go

// given a path to a kube config file, convert it into either creds for hitting the API server of the cluster it points to,
// or return the contexts/clusters it is aware of
type KubeLoader interface {
	GetConfigWithContext(masterURL, kubeconfigPath, context string) (clientcmd.ClientConfig, error)
	GetRestConfigForContext(path string, context string) (*rest.Config, error)
	GetRawConfigForContext(path, context string) (clientcmdapi.Config, error)
	RESTClientGetter(path, context string) resource.RESTClientGetter
	GetRestConfigFromBytes(config []byte) (*rest.Config, error)
}
