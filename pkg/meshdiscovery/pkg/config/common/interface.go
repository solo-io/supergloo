package common

type EnabledConfigLoops struct {
	Istio   bool
	AppMesh bool
	Linkerd bool
}

type AdvancedDiscoverySycnerList []AdvancedDiscoverySyncer
type AdvancedDiscoverySyncer interface {
	Run() (<-chan error, error)
	HandleError(err error)
}
