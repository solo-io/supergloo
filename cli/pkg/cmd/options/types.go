package options

type Options struct {
	Top         Top
	Install     Install
	MeshTool    MeshTool
	IngressTool IngressTool
	Get         Get
}

type Top struct {
	Static bool
}

type Install struct {
	Filename  string
	MeshType  string
	Namespace string
	Mtls      bool
}

type MeshTool struct {
	MeshId    string
	ServiceId string
}

type IngressTool struct {
	IngressId string
	RouteId   string
}

type Get struct {
	Output string
}
