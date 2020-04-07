package internal

import (
	"context"
	"fmt"

	healthcheck_types "github.com/solo-io/service-mesh-hub/cli/pkg/tree/check/healthcheck/types"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewSmhComponentsHealthCheck() healthcheck_types.HealthCheck {
	return &smhComponentsHealthCheck{}
}

type smhComponentsHealthCheck struct {
	smhNotInstalled bool
	problemPodName  string
}

func (*smhComponentsHealthCheck) GetDescription() string {
	return "components are running"
}

func (s *smhComponentsHealthCheck) Run(ctx context.Context, installNamespace string, clients healthcheck_types.Clients) (runFailure *healthcheck_types.RunFailure, checkApplies bool) {
	smhPods, err := clients.PodClient.List(ctx, client.InNamespace(installNamespace))
	if err != nil {
		return &healthcheck_types.RunFailure{
			ErrorMessage: GenericCheckFailed(err).Error(),
		}, true
	}

	return s.findProblemSmhPod(smhPods, installNamespace)
}

func (s *smhComponentsHealthCheck) findProblemSmhPod(smhPods *corev1.PodList, installNamespace string) (runFailure *healthcheck_types.RunFailure, checkApplies bool) {
	// default to assuming nothing is installed
	smhInstalled := false

	for _, pod := range smhPods.Items {
		// we found SMH pods
		smhInstalled = true

		for _, containerStatus := range pod.Status.ContainerStatuses {
			if containerStatus.State.Terminated != nil {
				return &healthcheck_types.RunFailure{
					ErrorMessage: fmt.Sprintf("Container %s in pod %s is terminated: %s (%s)", containerStatus.Name, pod.GetName(), containerStatus.State.Terminated.Reason, containerStatus.State.Terminated.Message),
					Hint:         s.buildHint(installNamespace, pod.GetName()),
				}, true
			}

			if containerStatus.State.Waiting != nil {
				return &healthcheck_types.RunFailure{
					ErrorMessage: fmt.Sprintf("Container %s in pod %s is waiting: %s (%s)", containerStatus.Name, pod.GetName(), containerStatus.State.Waiting.Reason, containerStatus.State.Waiting.Message),
					Hint:         s.buildHint(installNamespace, pod.GetName()),
				}, true
			}
		}
	}

	if smhInstalled {
		return nil, true
	} else {
		return &healthcheck_types.RunFailure{
			ErrorMessage: NoServiceMeshHubComponentsExist.Error(),
			Hint:         "you can install Service Mesh Hub with `meshctl install`",
		}, true
	}
}

func (s *smhComponentsHealthCheck) buildHint(installNamespace string, podName string) string {
	return fmt.Sprintf("try running either `kubectl -n %s describe pod %s` or `kubectl -n %s logs %s`", installNamespace, podName, installNamespace, podName)
}
