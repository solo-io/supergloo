package zephyr_networking

import (
	"context"

	networkingv1alpha1 "github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1/clientset/versioned"
	networking_client "github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1/clientset/versioned/typed/networking.zephyr.solo.io/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewGeneratedAccessControlPolicyClient(config *rest.Config) (AccessControlPolicyClient, error) {
	clientSet, err := versioned.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &generatedAccessControlPolicyClient{clientSet: clientSet.NetworkingV1alpha1()}, nil
}

type generatedAccessControlPolicyClient struct {
	clientSet networking_client.NetworkingV1alpha1Interface
}

func (g *generatedAccessControlPolicyClient) List(ctx context.Context, opts ...client.ListOption) (*networkingv1alpha1.AccessControlPolicyList, error) {
	listOptions := &client.ListOptions{}
	for _, v := range opts {
		v.ApplyToList(listOptions)
	}
	raw := metav1.ListOptions{}
	if converted := listOptions.AsListOptions(); converted != nil {
		raw = *converted
	}
	return g.clientSet.AccessControlPolicies(listOptions.Namespace).List(raw)
}

func (g *generatedAccessControlPolicyClient) UpdateStatus(ctx context.Context, acp *networkingv1alpha1.AccessControlPolicy, options ...client.UpdateOption) error {
	updated, err := g.clientSet.AccessControlPolicies(acp.GetNamespace()).UpdateStatus(acp)
	if err != nil {
		return err
	}
	acp.Status = updated.Status
	return nil
}
