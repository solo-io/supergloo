package options

type Options struct {
	Top     Top
	Install Install
}

type Top struct {
	Static bool
}

type Install struct {
	Filename  string
	MeshType  string
	Namespace string
	TopOpts   Top
}
