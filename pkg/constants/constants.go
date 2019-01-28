package constants

var (
	SuperglooNamespace           = "supergloo-system"
	GlooNamespace                = "gloo-system"
	SuperglooClusterRoleBindings = "cluster-admin-supergloo"
	MeshOptions                  = []string{"istio", "consul", "linkerd2", "appmesh"}
	ConsulInstallPath            = "https://s3.amazonaws.com/supergloo.solo.io/consul.tar.gz"
	IstioInstallPath             = "https://s3.amazonaws.com/supergloo.solo.io/istio-1.0.3.tgz"
	LinkerdInstallPath           = "https://s3.amazonaws.com/supergloo.solo.io/linkerd2-0.1.1.tgz"
)
