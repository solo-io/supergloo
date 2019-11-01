package setup

import (
	v1 "github.com/solo-io/mesh-projects/pkg/api/v1"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/eventloop"
)

type LoopSet struct {
	OperatorLoop    eventloop.EventLoop
	WatchOpts       clients.WatchOpts
	WatchNamespaces []string
}

func NewLoopSet(operatorLoop eventloop.EventLoop, opts clients.WatchOpts, watchNamespaces []string) LoopSet {
	return LoopSet{
		OperatorLoop:    operatorLoop,
		WatchOpts:       opts,
		WatchNamespaces: watchNamespaces,
	}
}

func NewNetworkBridgeSnapshotEmitter(emitter v1.NetworkBridgeEmitter) v1.NetworkBridgeSnapshotEmitter {
	return emitter
}
