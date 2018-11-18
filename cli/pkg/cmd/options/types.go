package options

type Options struct {
	Top      Top
	Install  Install
	MeshTool MeshTool
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

type MeshTool struct {
	MeshId    string
	ServiceId string
}
