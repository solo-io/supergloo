package mesh_workload

import (
	"context"
	"fmt"

	"github.com/rotisserie/eris"
	kubernetes_apps "github.com/solo-io/mesh-projects/pkg/clients/kubernetes/apps"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//go:generate mockgen -source ./owner_fetcher.go -destination mocks/mock_owner_fetcher.go

var (
	ControllerOwnerNotFound = func(namespace string, name string, kind string) error {
		return eris.New(fmt.Sprintf("Could not find owner reference with 'controller: true' for %s: %s/%s", kind, namespace, name))
	}
)

type OwnerFetcher interface {
	GetDeployment(ctx context.Context, pod *corev1.Pod) (*appsv1.Deployment, error)
}

func DefaultOwnerFetcher(dynamicClient client.Client) *ownerFetcher {
	return &ownerFetcher{
		deploymentClient: kubernetes_apps.NewDeploymentClient(dynamicClient),
		replicaSetClient: kubernetes_apps.NewReplicaSetClient(dynamicClient),
	}
}

// visible for testing
func NewOwnerFetcher(deploymentClient kubernetes_apps.DeploymentClient, replicaSetClient kubernetes_apps.ReplicaSetClient) *ownerFetcher {
	return &ownerFetcher{deploymentClient: deploymentClient, replicaSetClient: replicaSetClient}
}

type ownerFetcher struct {
	deploymentClient kubernetes_apps.DeploymentClient
	replicaSetClient kubernetes_apps.ReplicaSetClient
}

func (d *ownerFetcher) GetDeployment(ctx context.Context, pod *corev1.Pod) (*appsv1.Deployment, error) {
	namespace := pod.Namespace
	replicaName, err := d.getControllerName(pod, pod.TypeMeta)
	if err != nil {
		return nil, err
	}
	replicaSet, err := d.replicaSetClient.GetReplicaSet(ctx, client.ObjectKey{Namespace: namespace, Name: replicaName})
	if err != nil {
		return nil, err
	}
	deploymentName, err := d.getControllerName(replicaSet, replicaSet.TypeMeta)
	if err != nil {
		return nil, err
	}
	deployment, err := d.deploymentClient.GetDeployment(ctx, client.ObjectKey{Namespace: namespace, Name: deploymentName})
	if err != nil {
		return nil, err
	}
	return deployment, nil
}

func (d *ownerFetcher) getControllerName(obj metav1.Object, resourceType metav1.TypeMeta) (string, error) {
	for _, owner := range obj.GetOwnerReferences() {
		if owner.Controller != nil && *owner.Controller {
			return owner.Name, nil
		}
	}
	return "", ControllerOwnerNotFound(obj.GetNamespace(), obj.GetName(), resourceType.Kind)
}
