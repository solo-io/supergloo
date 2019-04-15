package common

import (
	"sync"
)

const SelectorPrefix = "discovered_by"

type EnabledConfigLoops struct {
	istio   bool
	appMesh bool
	linkerd bool

	lock sync.RWMutex
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
	Run() (<-chan error, error)
	HandleError(err error)
	Selector() string
}
