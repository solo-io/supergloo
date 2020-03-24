package helm_uninstall

import (
	"helm.sh/helm/v3/pkg/action"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func NewUninstallerFactory() UninstallerFactory {
	return func(getter genericclioptions.RESTClientGetter, namespace string, log action.DebugLog) (Uninstaller, error) {
		actionConfig := new(action.Configuration)
		err := actionConfig.Init(getter, namespace, "secrets", log)
		if err != nil {
			return nil, err
		}

		// just directly use the Helm implementation, hidden behind our interface
		return action.NewUninstall(actionConfig), nil
	}
}
