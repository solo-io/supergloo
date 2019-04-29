package clients

import (
	"context"
	"sync"

	"github.com/solo-io/supergloo/pkg/util"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/solo-io/go-utils/kubeutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

//go:generate mockgen -destination=./clientset_mock.go -source clientset.go -package clients

// Exposes only the methods that are actually used. Used to generate mocks.
type WebhookResourceClient interface {
	ListMeshes(namespace string, opts clients.ListOpts) (v1.MeshList, error)
	GetConfigMap(namespace, name string) (*corev1.ConfigMap, error)
	GetSuperglooNamespace() (string, error)
}

// Contains the resource clients required by the webhook
type webhookClientSet struct {
	meshClient v1.MeshClient
	kubeClient kubernetes.Interface
}

func (c *webhookClientSet) ListMeshes(namespace string, opts clients.ListOpts) (v1.MeshList, error) {
	return c.meshClient.List(namespace, opts)
}

func (c *webhookClientSet) GetConfigMap(namespace, name string) (*corev1.ConfigMap, error) {
	return c.kubeClient.CoreV1().ConfigMaps(namespace).Get(name, metav1.GetOptions{})
}

func (c *webhookClientSet) GetSuperglooNamespace() (string, error) {
	return util.GetSuperglooNamespace(c.kubeClient)
}

var (
	mutex           sync.Mutex
	globalClientSet WebhookResourceClient
)

func SetClientSet(clientSet WebhookResourceClient) {
	mutex.Lock()
	defer mutex.Unlock()
	globalClientSet = clientSet
}

func GetClientSet() WebhookResourceClient {
	mutex.Lock()
	defer mutex.Unlock()
	return globalClientSet
}

func InitClientSet(ctx context.Context) error {
	mutex.Lock()
	defer mutex.Unlock()

	if globalClientSet != nil {
		return nil
	}

	restConfig, err := kubeutils.GetConfig("", "")
	if err != nil {
		return err
	}
	kubeClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return err
	}

	mesh, err := v1.NewMeshClient(&factory.KubeResourceClientFactory{
		Crd:         v1.MeshCrd,
		Cfg:         restConfig,
		SharedCache: kube.NewKubeCache(ctx),
	})
	if err != nil {
		return err
	}
	if err := mesh.Register(); err != nil {
		return err
	}

	globalClientSet = &webhookClientSet{
		meshClient: mesh,
		kubeClient: kubeClient,
	}
	return nil
}
