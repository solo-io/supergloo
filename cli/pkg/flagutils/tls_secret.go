package flagutils

import (
	"github.com/solo-io/supergloo/cli/pkg/options"
	"github.com/spf13/pflag"
)

func AddCreateTlsSecretFlags(set *pflag.FlagSet, secret *options.CreateTlsSecret) {
	set.StringVar(&secret.RootCaFilename, "rootcert", "", "path to root-cert file")
	set.StringVar(&secret.CaCertFilename, "cacert", "", "path to ca-cert file")
	set.StringVar(&secret.PrivateKeyFilename, "cakey", "", "path to ca-key file")
	set.StringVar(&secret.CertChainFilename, "certchain", "", "path to cert-chain file")
}
