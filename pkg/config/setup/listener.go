package setup

import (
	"context"

	"github.com/solo-io/solo-kit/pkg/api/v1/eventloop"
	"github.com/solo-io/supergloo/pkg/api/clientset"
	"github.com/solo-io/supergloo/pkg/registration"
)

func StartSuperglooConfigLoop(ctx context.Context, cs *clientset.Clientset, manager *registration.Manager) {
	sgConfigLoop := &superglooConfigLoop{cs: cs}
	sgListener := registration.NewSubscriber(ctx, manager, sgConfigLoop)
	sgListener.Listen(ctx)
}

type superglooConfigLoop struct {
	cs *clientset.Clientset
}

func (scl *superglooConfigLoop) Enabled(enabled registration.EnabledConfigLoops) bool {
	return enabled.Linkerd || enabled.Istio || enabled.AppMesh
}

func (scl *superglooConfigLoop) Start(ctx context.Context, enabled registration.EnabledConfigLoops) (eventloop.EventLoop, error) {
	return createConfigStarters(scl.cs)(ctx, enabled)
}
