package install

type InstallationConfig struct {
	CreateNamespace            bool
	CreateIstioControlPlaneCRD bool

	// will be defaulted to istio-operator if left blank
	InstallNamespace string

	// will be defaulted to `DefaultIstioOperatorVersion` if left blank
	IstioOperatorVersion string
}
