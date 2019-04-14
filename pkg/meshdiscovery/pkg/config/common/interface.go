package common

type EnabledConfigLoops struct {
	Istio   bool
	AppMesh bool
	Linkerd bool
}

type AdvancedDiscoverySyncer interface {
	Run() (<-chan error, error)
	HandleError(err error)
	Enabled(enabled EnabledConfigLoops) bool
}
