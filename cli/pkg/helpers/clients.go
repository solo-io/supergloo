package helpers

import (
	"context"

	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/memory"
	"github.com/solo-io/solo-kit/pkg/errors"
	"github.com/solo-io/solo-kit/pkg/utils/kubeutils"
	"github.com/solo-io/solo-kit/pkg/utils/log"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func MustGetNamespaces() []string {
	ns, err := GetNamespaces()
	if err != nil {
		log.Fatalf("failed to list namespaces: %v", err)
	}
	return ns
}

// Note: requires RBAC permission to list namespaces at the cluster level
func GetNamespaces() ([]string, error) {
	if memoryResourceClient != nil {
		return []string{"default", "supergloo-system"}, nil
	}

	kubeClient, err := KubeClient()
	if err != nil {
		return nil, errors.Wrapf(err, "getting kube client")
	}
	var namespaces []string
	nsList, err := kubeClient.CoreV1().Namespaces().List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	for _, ns := range nsList.Items {
		namespaces = append(namespaces, ns.Name)
	}
	return namespaces, nil
}

var memoryResourceClient *factory.MemoryResourceClientFactory

func UseMemoryClients() {
	memoryResourceClient = &factory.MemoryResourceClientFactory{
		Cache: memory.NewInMemoryResourceCache(),
	}
}

func MustInstallClient() v1.InstallClient {
	client, err := InstallClient()
	if err != nil {
		log.Fatalf("failed to create install client: %v", err)
	}
	return client
}

func InstallClient() (v1.InstallClient, error) {
	if memoryResourceClient != nil {
		return v1.NewInstallClient(memoryResourceClient)
	}

	cfg, err := kubeutils.GetConfig("", "")
	if err != nil {
		return nil, errors.Wrapf(err, "getting kube config")
	}
	cache := kube.NewKubeCache(context.TODO())
	installClient, err := v1.NewInstallClient(&factory.KubeResourceClientFactory{
		Crd:         v1.InstallCrd,
		Cfg:         cfg,
		SharedCache: cache,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "creating install client")
	}
	if err := installClient.Register(); err != nil {
		return nil, err
	}
	return installClient, nil
}

func MustMeshClient() v1.MeshClient {
	client, err := MeshClient()
	if err != nil {
		log.Fatalf("failed to create mesh client: %v", err)
	}
	return client
}

func MeshClient() (v1.MeshClient, error) {
	if memoryResourceClient != nil {
		return v1.NewMeshClient(memoryResourceClient)
	}

	cfg, err := kubeutils.GetConfig("", "")
	if err != nil {
		return nil, errors.Wrapf(err, "getting kube config")
	}
	cache := kube.NewKubeCache(context.TODO())
	meshClient, err := v1.NewMeshClient(&factory.KubeResourceClientFactory{
		Crd:         v1.MeshCrd,
		Cfg:         cfg,
		SharedCache: cache,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "creating mesh client")
	}
	if err := meshClient.Register(); err != nil {
		return nil, err
	}
	return meshClient, nil
}

func MustUpstreamClient() gloov1.UpstreamClient {
	client, err := UpstreamClient()
	if err != nil {
		log.Fatalf("failed to create upstream client: %v", err)
	}
	return client
}

func UpstreamClient() (gloov1.UpstreamClient, error) {
	if memoryResourceClient != nil {
		return gloov1.NewUpstreamClient(memoryResourceClient)
	}

	cfg, err := kubeutils.GetConfig("", "")
	if err != nil {
		return nil, errors.Wrapf(err, "getting kube config")
	}
	cache := kube.NewKubeCache(context.TODO())
	upstreamClient, err := gloov1.NewUpstreamClient(&factory.KubeResourceClientFactory{
		Crd:         gloov1.UpstreamCrd,
		Cfg:         cfg,
		SharedCache: cache,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "creating upstreams client")
	}
	if err := upstreamClient.Register(); err != nil {
		return nil, err
	}
	return upstreamClient, nil
}

func MustKubeClient() kubernetes.Interface {
	client, err := KubeClient()
	if err != nil {
		log.Fatalf("failed to create kube client: %v", err)
	}
	return client
}

func KubeClient() (kubernetes.Interface, error) {
	cfg, err := kubeutils.GetConfig("", "")
	if err != nil {
		return nil, errors.Wrapf(err, "getting kube config")
	}
	return kubernetes.NewForConfig(cfg)
}
