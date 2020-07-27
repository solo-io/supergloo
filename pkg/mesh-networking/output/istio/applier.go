package istio

import (
	"context"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/snapshot/output"

	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/snapshot/input"
	"github.com/solo-io/skv2/pkg/ezkube"
	"github.com/solo-io/skv2/pkg/multicluster"
)

// the istio Applier applies a Snapshot of resources across clusters
type Applier interface {
	Apply(ctx context.Context, cli multicluster.Client, in input.Snapshot, out output.Snapshot) error
}

type applier struct {
}

func (a applier) HandleWriteError(resource ezkube.Object, err error) {

}

func (a applier) HandleDeleteError(resource ezkube.Object, err error) {
	panic("implement me")
}

func (a applier) HandleListError(err error) {
	panic("implement me")
}
