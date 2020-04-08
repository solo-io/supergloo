package internal

import (
	"context"

	healthcheck_types "github.com/solo-io/service-mesh-hub/cli/pkg/tree/check/healthcheck/types"
)

func NewKubeConnectivityCheck() healthcheck_types.HealthCheck {
	return &kubeConnectivityCheck{}
}

type kubeConnectivityCheck struct{}

func (*kubeConnectivityCheck) GetDescription() string {
	return "Kubernetes API server is reachable"
}

func (*kubeConnectivityCheck) Run(_ context.Context, _ string, clients healthcheck_types.Clients) (runFailure *healthcheck_types.RunFailure, checkApplies bool) {
	_, err := clients.ServerVersionClient.Get()
	if err != nil {
		return &healthcheck_types.RunFailure{
			ErrorMessage: KubernetesApiServerUnreachable(err).Error(),
		}, true
	}
	return nil, true
}
