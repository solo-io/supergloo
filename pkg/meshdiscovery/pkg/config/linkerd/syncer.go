package linkerd

import (
	"context"

	"github.com/solo-io/solo-kit/pkg/api/v1/eventloop"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/meshdiscovery/pkg/clientset"
)

const (
	injectionLabel = "istio-injection"
)

func NewLinkerdConfigDiscoveryRunner(ctx context.Context, cs *clientset.Clientset) (eventloop.EventLoop, error) {
	istioClient, err := clientset.IstioClientsetFromContext(ctx)
	if err != nil {
		return nil, err
	}

	emitter := v1.NewIstioDiscoveryEmitter(
		cs.Discovery.Mesh,
		cs.Input.Install,
		istioClient.MeshPolicies,
	)
	syncer := newLinkerdConfigDiscoverSyncer(cs)
	el := v1.NewIstioDiscoveryEventLoop(emitter, syncer)

	return el, nil
}

type linkerdConfigDiscoverSyncer struct {
	cs *clientset.Clientset
}

func newLinkerdConfigDiscoverSyncer(cs *clientset.Clientset) *linkerdConfigDiscoverSyncer {
	return &linkerdConfigDiscoverSyncer{cs: cs}
}
