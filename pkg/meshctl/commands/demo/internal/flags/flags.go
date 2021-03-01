package flags

import (
	"github.com/spf13/pflag"
)

type Options struct {
	Enterprise        bool
	LicenseKey        string
	EnterpriseVersion string
}

func (o *Options) AddToFlags(flags *pflag.FlagSet) {
	flags.BoolVar(&o.Enterprise, "enterprise", false, "install the enterprise features, requires a license key")
	flags.StringVar(&o.LicenseKey, "license", "", "Gloo Mesh Enterprise license key")
	flags.StringVar(&o.EnterpriseVersion, "enterprise-version", "", "Gloo Mesh Enterprise version (defaults to latest)")
}
