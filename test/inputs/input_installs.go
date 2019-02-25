package inputs

import (
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
)

func IstioInstall(name, namespace, installNs, version string, disabled bool) *v1.Install {

	return &v1.Install{
		Metadata: core.Metadata{Name: name, Namespace: namespace},
		Disabled: disabled,
		InstallType: &v1.Install_Istio_{
			Istio: &v1.Install_Istio{
				InstallationNamespace: installNs,
				IstioVersion:          version,
			},
		},
	}
}
