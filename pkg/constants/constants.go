package constants

import "time"

const (
	ServiceMeshHubChartUriTemplate = "https://storage.googleapis.com/service-mesh-hub/management-plane/service-mesh-hub-%s.tgz"
	ServiceMeshHubReleaseName      = "service-mesh-hub"
	ServiceMeshHubApplicationName  = "service-mesh-hub"
	DefaultReleaseTag              = "latest"
	DefaultKubeClientTimeout       = 5 * time.Second

	ServiceMeshHubApiGroupSuffix = "zephyr.solo.io"

	CsrAgentReleaseName = "csr-agent"

	// https://kubernetes.io/docs/concepts/overview/working-with-objects/common-labels/
	ManagedByLabel = "app.kubernetes.io/managed-by"
)
