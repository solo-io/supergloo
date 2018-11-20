package common

const (
	SuperglooGroupName = "supergloo.solo.io"

	// Mesh types
	Istio    = "istio"
	Consul   = "consul"
	Linkerd2 = "linkerd2"

	// The character used as a separator when specifying CLI options that are to be interpreted as a list of values
	ListOptionSeparator = ","

	// The character used as a separator when specifying CLI options values that are namespace-scoped
	// E.g. "--source=my-namespace:my-upstream-name"
	NamespacedResourceSeparator = ":"

	ValidMatcherHttpMethods = "GET|HEAD|POST|PUT|PATCH|DELETE|CONNECT|OPTIONS|TRACE"
)
