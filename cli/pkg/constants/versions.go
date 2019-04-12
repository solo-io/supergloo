package constants

import (
	"github.com/solo-io/supergloo/pkg/install/istio"
	"github.com/solo-io/supergloo/pkg/install/linkerd"
)

var SupportedGlooVersions = []string{
	"latest",
	"v0.13.0",
}

var SupportedIstioVersions = []string{
	istio.IstioVersion103,
	istio.IstioVersion105,
	istio.IstioVersion106,
}

var SupportedLinkerdVersions = []string{
	linkerd.Version_stable221,
}
