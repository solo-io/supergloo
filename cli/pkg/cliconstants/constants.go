package cliconstants

import "time"

const (
	ServiceMeshHubChartUriTemplate = "https://storage.googleapis.com/service-mesh-hub/management-plane/service-mesh-hub-%s.tgz"
	ServiceMeshHubReleaseName      = "service-mesh-hub"
	DefaultReleaseTag              = "latest"
	DefaultKubeClientTimeout       = 5 * time.Second

	DefaultIstioOperatorNamespace      = "istio-operator"
	DefaultIstioOperatorDeploymentName = "istio-operator"

	ServiceMeshHubApiGroupSuffix = "zephyr.solo.io"

	CsrAgentReleaseName = "csr-agent"
)
