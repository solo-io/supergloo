package constants

import (
	"github.com/solo-io/supergloo/pkg/install/istio"
	"github.com/solo-io/supergloo/pkg/install/linkerd"
)

var SupportedGlooVersions = []string{
	"latest",
	"v0.13.0",
	"v0.14.1",
}

var SupportedIstioVersions = []string{
	istio.IstioVersion103,
	istio.IstioVersion105,
	istio.IstioVersion106,
	//istio.IstioVersion113, //TODO: enable when istio1.1x is ready to ship
}

var SupportedLinkerdVersions = []string{
	linkerd.Version_stable230,
}
