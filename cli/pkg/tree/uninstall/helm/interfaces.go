package helm_uninstall

import (
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/release"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

//go:generate mockgen -source ./interfaces.go -destination ./mocks/mock_uninstaller.go

type UninstallerFactory func(getter genericclioptions.RESTClientGetter, namespace string, log action.DebugLog) (Uninstaller, error)

type Uninstaller interface {
	Run(releaseName string) (*release.UninstallReleaseResponse, error)
}
