package istio

import (
	"context"
	"strings"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/common/kubernetes"
	"github.com/solo-io/supergloo/pkg/api/external/istio/authorization/v1alpha1"
	batchv1 "k8s.io/api/batch/v1"
	kubev1 "k8s.io/api/core/v1"
	kubeerrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	batchv1client "k8s.io/client-go/kubernetes/typed/batch/v1"
)

type PilotDeployment struct {
	Version, Namespace string
}

var PostInstallJobs = []string{
	"istio-security-post-install",
}

type IstioResourceDetector interface {
	DetectPilotDeployments(ctx context.Context, deployments kubernetes.DeploymentList) []PilotDeployment
	DetectMeshPolicyCrd(crdGetter CrdGetter) (bool, error)
	DetectPostInstallJobComplete(jobGetter batchv1client.JobsGetter, pilotNamespace string) (bool, error)
	DetectInjectedIstioPods(ctx context.Context, pods kubernetes.PodList) map[string]kubernetes.PodList
}

type istioResourceDetector struct{}

var _ IstioResourceDetector = istioResourceDetector{}

func NewIstioResourceDetector() istioResourceDetector {
	return istioResourceDetector{}
}

func (i istioResourceDetector) DetectPilotDeployments(ctx context.Context, deployments kubernetes.DeploymentList) []PilotDeployment {
	var pilots []PilotDeployment
	for _, deployment := range deployments {
		for _, container := range deployment.Spec.Template.Spec.Containers {
			if (strings.Contains(container.Image, "istio") || strings.Contains(container.Image, "openshift")) && strings.Contains(container.Image, "pilot") {
				split := strings.Split(container.Image, ":")
				if len(split) != 2 {
					contextutils.LoggerFrom(ctx).Errorf("invalid or unexpected image format for pilot: %v", container.Image)
					continue
				}
				pilots = append(pilots, PilotDeployment{Version: split[1], Namespace: deployment.Namespace})
			}
		}
	}
	return pilots
}

func (i istioResourceDetector) DetectMeshPolicyCrd(crdGetter CrdGetter) (bool, error) {
	_, err := crdGetter.Get(v1alpha1.MeshPolicyCrd.FullName(), metav1.GetOptions{})
	if err == nil {
		return true, nil
	}
	if kubeerrs.IsNotFound(err) {
		return false, nil
	}
	return false, err
}

func (i istioResourceDetector) DetectPostInstallJobComplete(jobGetter batchv1client.JobsGetter, pilotNamespace string) (bool, error) {
	jobsList, err := jobGetter.Jobs(pilotNamespace).List(metav1.ListOptions{})
	if err != nil {
		return false, err
	}

	getJobFromPrefix := func(prefix string) *batchv1.Job {
		for _, job := range jobsList.Items {
			if strings.HasPrefix(job.Name, prefix) {
				return &job
			}
		}
		return nil
	}

	for _, jobName := range PostInstallJobs {
		job := getJobFromPrefix(jobName)
		if job == nil {
			return false, nil
		}

		var jobComplete bool
		for _, condition := range job.Status.Conditions {
			if condition.Type == batchv1.JobComplete && condition.Status == kubev1.ConditionTrue {
				jobComplete = true
				break
			}
		}
		if !jobComplete {
			return false, nil
		}
	}
	return true, nil
}

func (i istioResourceDetector) DetectInjectedIstioPods(ctx context.Context, pods kubernetes.PodList) map[string]kubernetes.PodList {
	injectedPods := make(map[string]kubernetes.PodList)
	for _, pod := range pods {
		discoveryNamespace, ok := detectInjectedIstioPod(ctx, pod)
		if ok {
			injectedPods[discoveryNamespace] = append(injectedPods[discoveryNamespace], pod)
		}
	}
	return injectedPods
}

func detectInjectedIstioPod(ctx context.Context, pod *kubernetes.Pod) (string, bool) {
	for _, container := range pod.Spec.Containers {
		if container.Name == "istio-proxy" {
			for i, arg := range container.Args {
				if arg == "--discoveryAddress" {
					if i == len(container.Args) {
						contextutils.LoggerFrom(ctx).Errorf("invalid args for istio-proxy sidecar for pod %v.%v: "+
							"expected to find --discoveryAddress with a parameter. instead, found %v", pod.Namespace, pod.Name,
							container.Args)
						return "", false
					}
					discoveryAddress := container.Args[i+1]
					discoveryService := strings.Split(discoveryAddress, ":")[0]
					discoveryNamespace := strings.TrimPrefix(discoveryService, "istio-pilot.")

					return discoveryNamespace, true
				}
			}
		}
	}
	return "", false
}
