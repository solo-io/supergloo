package flagutils

import (
	"github.com/solo-io/supergloo/cli/pkg/options"
	"github.com/spf13/pflag"
)

func AddSetRootCertFlags(set *pflag.FlagSet, in *options.SetRootCert) {
	set.Var(&in.TargetMesh, "target-mesh", "resource reference the mesh for which you wish to set the root cert. format must be <NAMESPACE>.<NAME>")
	set.Var(&in.TlsSecret, "tls-secret", "resource reference the TLS Secret (created with supergloo CLI) which you wish to use as the custom root cert for the mesh. if empty, the any existing custom root cert will be removed from this mesh. format must be <NAMESPACE>.<NAME>")
}
