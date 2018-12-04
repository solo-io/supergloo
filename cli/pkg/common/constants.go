package common

const (
	SuperglooGroupName     = "supergloo.solo.io"
	HelmSetupFileName      = "https://raw.githubusercontent.com/solo-io/supergloo/master/hack/install/helm/helm-service-account.yaml"
	SuperglooSetupFileName = "https://raw.githubusercontent.com/solo-io/supergloo/master/hack/install/supergloo.yaml"

	// Mesh types
	Istio    = "istio"
	Consul   = "consul"
	Linkerd2 = "linkerd2"
	AppMesh  = "appmesh"

	// The character used as a separator when specifying CLI options that are to be interpreted as a list of values
	ListOptionSeparator = ","

	// The character used as a separator for elements of a CLI lis option that are lists themselves
	// E.g. "--matchers="prefix=/root,methods=GET|PUT"
	SubListOptionSeparator = "|"

	// The character used as a separator when specifying CLI options values that are namespace-scoped
	// E.g. "--source=my-namespace:my-upstream-name"
	NamespacedResourceSeparator = ":"

	ValidMatcherHttpMethods = "GET|HEAD|POST|PUT|PATCH|DELETE|CONNECT|OPTIONS|TRACE"
)
