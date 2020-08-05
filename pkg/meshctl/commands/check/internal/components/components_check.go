package components

import (
	"context"
	"fmt"

	"github.com/rotisserie/eris"
	v1 "github.com/solo-io/external-apis/pkg/api/k8s/core/v1"
	"github.com/solo-io/service-mesh-hub/pkg/common/defaults"
	"github.com/solo-io/service-mesh-hub/pkg/meshctl/commands/check/internal"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type componentsCheck struct{}

func NewComponentsCheck() internal.Check {
	return &componentsCheck{}
}

func (s *componentsCheck) GetDescription() string {
	return "deployments are running"
}

func (s *componentsCheck) Run(ctx context.Context, c client.Client, installNamespace string) *internal.Failure {
	namespaceClient := v1.NewNamespaceClient(c)
	_, err := namespaceClient.GetNamespace(ctx, installNamespace)
	if err != nil {
		return &internal.Failure{
			ErrorMessage: eris.Wrapf(err, "specified namespace %s doesn't exist", installNamespace).Error(),
		}
	}

	podClient := v1.NewPodClient(c)
	smhPods, err := podClient.ListPod(ctx, client.InNamespace(installNamespace))
	if err != nil {
		return &internal.Failure{
			ErrorMessage: err.Error(),
		}
	}

	return checkPods(smhPods, installNamespace)
}

func checkPods(smhPods *corev1.PodList, installNamespace string) *internal.Failure {
	if len(smhPods.Items) < 1 {
		return &internal.Failure{
			ErrorMessage: fmt.Sprintf("no pods found in namespace %s", installNamespace),
			Hint: fmt.Sprintf(
				`Service Mesh Hub's installation namespace can be supplied to this cmd with the "--namespace" flag, which defaults to %s`,
				defaults.DefaultPodNamespace),
		}
	}
	for _, pod := range smhPods.Items {
		for _, containerStatus := range pod.Status.ContainerStatuses {
			if containerStatus.State.Terminated != nil {
				return &internal.Failure{
					ErrorMessage: fmt.Sprintf("Container %s in pod %s is terminated: %s (%s)", containerStatus.Name, pod.GetName(), containerStatus.State.Terminated.Reason, containerStatus.State.Terminated.Message),
					Hint:         buildHint(installNamespace, pod.GetName()),
				}
			}

			if containerStatus.State.Waiting != nil {
				return &internal.Failure{
					ErrorMessage: fmt.Sprintf("Container %s in pod %s is waiting: %s (%s)", containerStatus.Name, pod.GetName(), containerStatus.State.Waiting.Reason, containerStatus.State.Waiting.Message),
					Hint:         buildHint(installNamespace, pod.GetName()),
				}
			}
		}
	}
	return nil
}

func buildHint(installNamespace string, podName string) string {
	return fmt.Sprintf("try running either `kubectl -n %s describe pod %s` or `kubectl -n %s logs %s`", installNamespace, podName, installNamespace, podName)
}
