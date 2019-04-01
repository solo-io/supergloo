package flagutils

import (
	"github.com/solo-io/supergloo/cli/pkg/options"
	"github.com/spf13/pflag"
)

func AddProxyFlags(set *pflag.FlagSet, opts *options.GetMeshIngress) {
	set.StringVar(&opts.Proxy.Name, "name", "gateway-proxy", "the name of the proxy service/deployment to use")
	set.StringVar(&opts.Proxy.Port, "port", "http", "the name of the service port to connect to")
}
