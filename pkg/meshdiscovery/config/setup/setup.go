package setup

import (
	"context"

	"github.com/solo-io/solo-kit/pkg/api/v1/eventloop"
	"github.com/solo-io/supergloo/pkg/meshdiscovery/clientset"
	"github.com/solo-io/supergloo/pkg/meshdiscovery/config/istio"
	"github.com/solo-io/supergloo/pkg/meshdiscovery/config/linkerd"
	"github.com/solo-io/supergloo/pkg/registration"
)

func NewDiscoveryConfigLoopStarters(clientset *clientset.Clientset) registration.ConfigLoopStarters {
	starters := createConfigStarters(clientset)
	return starters
}

func createConfigStarters(cs *clientset.Clientset) registration.ConfigLoopStarters {
	return registration.ConfigLoopStarters{
		createIstioConfigLoopStarter(cs),
		createLinkerdConfigLoopStarter(cs),
	}
}

func createIstioConfigLoopStarter(cs *clientset.Clientset) registration.ConfigLoopStarter {
	return func(ctx context.Context, enabled registration.EnabledConfigLoops) (eventloop.EventLoop, error) {
		if enabled.Istio {
			return istio.NewIstioConfigDiscoveryRunner(ctx, cs)
		}
		return nil, nil
	}
}

func createLinkerdConfigLoopStarter(cs *clientset.Clientset) registration.ConfigLoopStarter {
	return func(ctx context.Context, enabled registration.EnabledConfigLoops) (eventloop.EventLoop, error) {
		if enabled.Linkerd {
			return linkerd.NewLinkerdConfigDiscoveryRunner(ctx, cs)
		}
		return nil, nil
	}
}
