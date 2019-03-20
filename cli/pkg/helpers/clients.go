package helpers

import (
	"context"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/cache"

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
	sharedCache := kube.NewKubeCache(context.TODO())
	installClient, err := v1.NewInstallClient(&factory.KubeResourceClientFactory{
		Crd:         v1.InstallCrd,
		Cfg:         cfg,
		SharedCache: sharedCache,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "creating install client")
	}
	if err := installClient.Register(); err != nil {
		return nil, err
	}
	return installClient, nil
}

func MustRoutingRuleClient() v1.RoutingRuleClient {
	client, err := RoutingRuleClient()
	if err != nil {
		log.Fatalf("failed to create RoutingRule client: %v", err)
	}
	return client
}

func RoutingRuleClient() (v1.RoutingRuleClient, error) {
	if memoryResourceClient != nil {
		return v1.NewRoutingRuleClient(memoryResourceClient)
	}

	cfg, err := kubeutils.GetConfig("", "")
	if err != nil {
		return nil, errors.Wrapf(err, "getting kube config")
	}
	sharedCache := kube.NewKubeCache(context.TODO())
	RoutingRuleClient, err := v1.NewRoutingRuleClient(&factory.KubeResourceClientFactory{
		Crd:         v1.RoutingRuleCrd,
		Cfg:         cfg,
		SharedCache: sharedCache,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "creating RoutingRule client")
	}
	if err := RoutingRuleClient.Register(); err != nil {
		return nil, err
	}
	return RoutingRuleClient, nil
}

func MustMeshIngressClient() v1.MeshIngressClient {
	client, err := MeshIngressClient()
	if err != nil {
		log.Fatalf("failed to create mesh client: %v", err)
	}
	return client
}

func MeshIngressClient() (v1.MeshIngressClient, error) {
	if memoryResourceClient != nil {
		return v1.NewMeshIngressClient(memoryResourceClient)
	}

	cfg, err := kubeutils.GetConfig("", "")
	if err != nil {
		return nil, errors.Wrapf(err, "getting kube config")
	}
	sharedCache := kube.NewKubeCache(context.TODO())
	meshClient, err := v1.NewMeshIngressClient(&factory.KubeResourceClientFactory{
		Crd:         v1.MeshIngressCrd,
		Cfg:         cfg,
		SharedCache: sharedCache,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "creating MeshIngress client")
	}
	if err := meshClient.Register(); err != nil {
		return nil, err
	}
	return meshClient, nil
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
	sharedCache := kube.NewKubeCache(context.TODO())
	meshClient, err := v1.NewMeshClient(&factory.KubeResourceClientFactory{
		Crd:         v1.MeshCrd,
		Cfg:         cfg,
		SharedCache: sharedCache,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "creating mesh client")
	}
	if err := meshClient.Register(); err != nil {
		return nil, err
	}
	return meshClient, nil
}

func MustTlsSecretClient() v1.TlsSecretClient {
	client, err := TlsSecretClient()
	if err != nil {
		log.Fatalf("failed to create tlsSecret client: %v", err)
	}
	return client
}

func TlsSecretClient() (v1.TlsSecretClient, error) {
	if memoryResourceClient != nil {
		return v1.NewTlsSecretClient(memoryResourceClient)
	}

	kubeClient := MustKubeClient()
	kubeCache, err := cache.NewKubeCoreCache(context.TODO(), kubeClient)
	if err != nil {
		return nil, errors.Wrapf(err, "creating kube cache")
	}

	tlsSecretClient, err := v1.NewTlsSecretClient(&factory.KubeSecretClientFactory{
		Clientset:    kubeClient,
		PlainSecrets: true,
		Cache:        kubeCache,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "creating tlsSecret client")
	}
	if err := tlsSecretClient.Register(); err != nil {
		return nil, err
	}
	return tlsSecretClient, nil
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
	sharedCache := kube.NewKubeCache(context.TODO())
	upstreamClient, err := gloov1.NewUpstreamClient(&factory.KubeResourceClientFactory{
		Crd:         gloov1.UpstreamCrd,
		Cfg:         cfg,
		SharedCache: sharedCache,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "creating upstreams client")
	}
	if err := upstreamClient.Register(); err != nil {
		return nil, err
	}
	return upstreamClient, nil
}

func MustSecretClient() gloov1.SecretClient {
	client, err := SecretClient()
	if err != nil {
		log.Fatalf("failed to create secret client: %v", err)
	}
	return client
}

func SecretClient() (gloov1.SecretClient, error) {
	if memoryResourceClient != nil {
		return gloov1.NewSecretClient(memoryResourceClient)
	}

	kubeClient := MustKubeClient()
	kubeCoreCache, err := cache.NewKubeCoreCache(context.TODO(), kubeClient)
	if err != nil {
		return nil, err
	}
	secretClient, err := gloov1.NewSecretClient(&factory.KubeSecretClientFactory{
		Clientset:    kubeClient,
		Cache:        kubeCoreCache,
		PlainSecrets: true,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "creating secret client")
	}
	if err := secretClient.Register(); err != nil {
		return nil, err
	}
	return secretClient, nil
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
