package constants

import (
	"github.com/solo-io/supergloo/pkg/install/mesh"
)

var SupportedIstioVersions = []string{
	mesh.IstioVersion103,
	mesh.IstioVersion105,
	mesh.IstioVersion106,
}
