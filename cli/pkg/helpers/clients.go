package helpers

import (
	"context"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/memory"
	"github.com/solo-io/solo-kit/pkg/errors"
	"github.com/solo-io/solo-kit/pkg/utils/kubeutils"
	"github.com/solo-io/solo-kit/pkg/utils/log"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
)

var memoryResourceClient *factory.MemoryResourceClientFactory

func UseMemoryClients() {
	memoryResourceClient = &factory.MemoryResourceClientFactory{
		Cache: memory.NewInMemoryResourceCache(),
	}
}

func MustInstallClient() v1.InstallClient {
	client, err := InstallClient()
	if err != nil {
		log.Fatalf("failed to create upstream client: %v", err)
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
