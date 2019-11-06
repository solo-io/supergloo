package istio

import (
	"context"
	"strings"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/mesh-projects/pkg/api/external/istio/authorization/v1alpha1"
	"github.com/solo-io/mesh-projects/services/mesh-discovery/pkg/common/injectedpods"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/common/kubernetes"
	sk_errors "github.com/solo-io/solo-kit/pkg/errors"
	batchv1 "k8s.io/api/batch/v1"
	kubev1 "k8s.io/api/core/v1"
)

type PilotDeployment struct {
	Version, Namespace, Cluster string
}

// TODO merge with linkerd controller type
func (c PilotDeployment) Name() string {
	if c.Cluster == "" {
		return "istio-" + c.Namespace
	}
	//TODO cluster is not restricted to kube name scheme, kebab it
	return "istio-" + c.Namespace + "-" + c.Cluster
}

var PostInstallJobs = []string{
	"istio-security-post-install",
}

type IstioResourceDetector interface {
	DetectPilotDeployments(ctx context.Context, deployments kubernetes.DeploymentList) []PilotDeployment
	DetectMeshPolicyCrd(crdGetter CrdGetter, cluster string) (bool, error)
	DetectPostInstallJobComplete(jobClient kubernetes.JobClient, pilotNamespace, pilotCluster string) (bool, error)
	DetectInjectedIstioPods(ctx context.Context, pods kubernetes.PodList) injectedpods.InjectedPods
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
			if strings.Contains(container.Image, "istio") && strings.Contains(container.Image, "pilot") {
				split := strings.Split(container.Image, ":")
				if len(split) != 2 {
					contextutils.LoggerFrom(ctx).Errorf("invalid or unexpected image format for pilot: %v", container.Image)
					continue
				}
				pilots = append(pilots, PilotDeployment{Version: split[1], Namespace: deployment.Namespace, Cluster: deployment.ClusterName})
			}
		}
	}
	return pilots
}

func (i istioResourceDetector) DetectMeshPolicyCrd(crdGetter CrdGetter, cluster string) (bool, error) {
	_, err := crdGetter.Read(v1alpha1.MeshPolicyCrd.FullName(), clients.ReadOpts{Cluster: cluster})
	if err == nil {
		return true, nil
	}
	if sk_errors.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func (i istioResourceDetector) DetectPostInstallJobComplete(jobClient kubernetes.JobClient, pilotNamespace, pilotCluster string) (bool, error) {
	jobsList, err := jobClient.List(pilotNamespace, clients.ListOpts{Cluster: pilotCluster})
	if err != nil {
		return false, err
	}

	getJobFromPrefix := func(prefix string) *kubernetes.Job {
		for _, job := range jobsList {
			if strings.HasPrefix(job.Name, prefix) {
				return job
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

func (i istioResourceDetector) DetectInjectedIstioPods(ctx context.Context, pods kubernetes.PodList) injectedpods.InjectedPods {
	return injectedpods.NewDetector(detectInjectedIstioPod).DetectInjectedPods(ctx, pods)
}

func detectInjectedIstioPod(ctx context.Context, pod *kubernetes.Pod) (string, string, bool) {
	for _, container := range pod.Spec.Containers {
		if container.Name == "istio-proxy" {
			for i, arg := range container.Args {
				if arg == "--discoveryAddress" {
					if i == len(container.Args) {
						contextutils.LoggerFrom(ctx).Errorf("invalid args for istio-proxy sidecar for pod %v.%v: "+
							"expected to find --discoveryAddress with a parameter. instead, found %v", pod.Namespace, pod.Name,
							container.Args)
						return "", "", false
					}
					discoveryAddress := container.Args[i+1]
					discoveryService := strings.Split(discoveryAddress, ":")[0]
					discoveryNamespace := strings.TrimPrefix(discoveryService, "istio-pilot.")

					return pod.ClusterName, discoveryNamespace, true
				}
			}
		}
	}
	return pod.ClusterName, "", false
}
