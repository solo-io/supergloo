package flags

import (
	"github.com/spf13/pflag"
)

type Options struct {
	Version    string
	Enterprise bool
	LicenseKey string
}

func (o *Options) AddToFlags(flags *pflag.FlagSet) {
	flags.StringVar(&o.Version, "version", "",
		"Gloo Mesh version to install.\nCommunity defaults to meshctl version, enterprise defaults to latest")
	flags.BoolVar(&o.Enterprise, "enterprise", false, "Install the enterprise features, requires a license key")
	flags.StringVar(&o.LicenseKey, "license", "", "Gloo Mesh Enterprise license key")
}
