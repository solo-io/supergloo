package description

import (
	"context"
	"fmt"
	"io"

	smh_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
)

//go:generate mockgen -source ./interfaces.go -destination mocks/mock_interfaces.go

type ResourceDescriber interface {
	DescribeService(ctx context.Context, kubeResourceIdentifier FullyQualifiedKubeResource) (*DescriptionResult, error)
	DescribeWorkload(ctx context.Context, kubeResourceIdentifier FullyQualifiedKubeResource) (*DescriptionResult, error)
}

type Printer interface {
	PrintForService(out io.Writer, serviceName *FullyQualifiedKubeResource, result *DescriptionResult) error
}

type DescriptionResult struct {
	Policies *Policies
}

type Policies struct {
	AccessControlPolicies []*smh_networking.AccessControlPolicy
	TrafficPolicies       []*smh_networking.TrafficPolicy
}

// the name/namespace/cluster of a kube-native resource, like a Service or a Pod
type FullyQualifiedKubeResource struct {
	Name        string
	Namespace   string
	ClusterName string
}

func (f *FullyQualifiedKubeResource) String() string {
	return fmt.Sprintf("%s.%s.%s", f.Name, f.Namespace, f.ClusterName)
}
