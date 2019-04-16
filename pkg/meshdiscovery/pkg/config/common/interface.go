package common

import (
	"context"
	"sync"
)

const SelectorPrefix = "discovered_by"

type EnabledConfigLoops struct {
	istio   bool
	appMesh bool
	linkerd bool

	lock sync.RWMutex
}

func Diff(old, new *EnabledConfigLoops) bool {
	return old.Linkerd() != new.Linkerd() || old.Istio() != new.Istio() || old.Appmesh() != new.Appmesh()
}

func (e *EnabledConfigLoops) SetIstio(state bool) {
	e.lock.Lock()
	defer e.lock.Unlock()
	e.istio = state
}

func (e *EnabledConfigLoops) SetAppmesh(state bool) {
	e.lock.Lock()
	defer e.lock.Unlock()
	e.linkerd = state
}

func (e *EnabledConfigLoops) SetLinkerd(state bool) {
	e.lock.Lock()
	defer e.lock.Unlock()
	e.linkerd = state
}

func (e *EnabledConfigLoops) Istio() bool {
	e.lock.RLock()
	defer e.lock.RUnlock()
	return e.istio
}

func (e *EnabledConfigLoops) Appmesh() bool {
	e.lock.RLock()
	defer e.lock.RUnlock()
	return e.appMesh
}

func (e *EnabledConfigLoops) Linkerd() bool {
	e.lock.RLock()
	defer e.lock.RUnlock()
	return e.linkerd
}

type AdvancedDiscoverySycnerList []AdvancedDiscoverySyncer
type AdvancedDiscoverySyncer interface {
	Run(ctx context.Context) chan<- *EnabledConfigLoops
}
