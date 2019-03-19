package constants

import (
	"github.com/solo-io/supergloo/pkg/install/istio"
)

var SupportedIstioVersions = []string{
	istio.IstioVersion103,
	istio.IstioVersion105,
	istio.IstioVersion106,
}

var SupportedGlooVersions = []string{
	"latest",
	"v0.11.1",
	"v0.10.5",
}
