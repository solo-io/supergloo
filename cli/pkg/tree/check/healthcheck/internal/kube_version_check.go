package internal

import (
	"context"
	"fmt"
	"regexp"
	"strconv"

	healthcheck_types "github.com/solo-io/service-mesh-hub/cli/pkg/tree/check/healthcheck/types"
	"github.com/solo-io/service-mesh-hub/pkg/version"
)

func NewK8sServerVersionCheck() healthcheck_types.HealthCheck {
	return &k8sServerVersionCheck{}
}

type k8sServerVersionCheck struct{}

func (*k8sServerVersionCheck) GetDescription() string {
	return fmt.Sprintf("running the minimum supported Kubernetes version (required: >=1.%d)", version.MinimumSupportedKubernetesMinorVersion)
}

func (*k8sServerVersionCheck) Run(_ context.Context, _ string, clients healthcheck_types.Clients) (runFailure *healthcheck_types.RunFailure, checkApplies bool) {
	versionInfo, err := clients.ServerVersionClient.Get()
	if err != nil {
		return &healthcheck_types.RunFailure{
			ErrorMessage: KubernetesApiServerUnreachable(err).Error(),
		}, true
	}

	// GKE may have nonstandard minor version numbers
	// our own GKE clusters report a minor version of something like "15+"
	minorVersionNumericPartPattern := regexp.MustCompile("^([0-9]+)")
	matches := minorVersionNumericPartPattern.FindStringSubmatch(versionInfo.Minor)
	if len(matches) != 2 {
		return &healthcheck_types.RunFailure{
			ErrorMessage: KubernetesServerVersionUnsupported(versionInfo.Minor).Error(),
		}, true
	}
	minorVersionNumericPart := matches[1]

	minorVersion, err := strconv.Atoi(minorVersionNumericPart)
	if err != nil || minorVersion < version.MinimumSupportedKubernetesMinorVersion {
		return &healthcheck_types.RunFailure{
			ErrorMessage: KubernetesServerVersionUnsupported(versionInfo.Minor).Error(),
		}, true
	}

	return nil, true
}
