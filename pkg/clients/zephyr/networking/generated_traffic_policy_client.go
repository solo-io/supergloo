package zephyr_networking

import (
	"context"

	networking_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1/clientset/versioned"
	networkingv1alpha1 "github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1/clientset/versioned/typed/networking.zephyr.solo.io/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type generatedTrafficPolicyClient struct {
	client networkingv1alpha1.NetworkingV1alpha1Interface
}

func NewGeneratedTrafficPolicyClient(cfg *rest.Config) TrafficPolicyClient {
	clientSet, _ := versioned.NewForConfig(cfg)
	return &generatedTrafficPolicyClient{client: clientSet.NetworkingV1alpha1()}
}

func (g *generatedTrafficPolicyClient) Get(_ context.Context, name string, namespace string) (*networking_v1alpha1.TrafficPolicy, error) {
	return g.client.TrafficPolicies(namespace).Get(name, metav1.GetOptions{})
}

func (g *generatedTrafficPolicyClient) Create(_ context.Context, trafficPolicy *networking_v1alpha1.TrafficPolicy, options ...client.CreateOption) error {
	newVirtualMesh, err := g.client.TrafficPolicies(trafficPolicy.GetNamespace()).Create(trafficPolicy)
	if err != nil {
		return err
	}
	*trafficPolicy = *newVirtualMesh
	return nil
}

func (g *generatedTrafficPolicyClient) UpdateStatus(_ context.Context, trafficPolicy *networking_v1alpha1.TrafficPolicy, options ...client.UpdateOption) error {
	updated, err := g.client.TrafficPolicies(trafficPolicy.GetNamespace()).UpdateStatus(trafficPolicy)
	if err != nil {
		return err
	}
	trafficPolicy.Status = updated.Status
	return nil
}

func (g *generatedTrafficPolicyClient) List(ctx context.Context, options ...client.ListOption) (*networking_v1alpha1.TrafficPolicyList, error) {
	listOptions := &client.ListOptions{}
	for _, v := range options {
		v.ApplyToList(listOptions)
	}
	raw := metav1.ListOptions{}
	if converted := listOptions.AsListOptions(); converted != nil {
		raw = *converted
	}
	return g.client.TrafficPolicies(listOptions.Namespace).List(raw)
}
