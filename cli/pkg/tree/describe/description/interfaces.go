package description

import (
	"context"
	"fmt"
	"io"

	"github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
)

//go:generate mockgen -source ./interfaces.go -destination mocks/mock_interfaces.go

type ResourceDescriber interface {
	DescribeService(ctx context.Context, kubeResourceIdentifier FullyQualifiedKubeResource) (*ExplorationResult, error)
	DescribeWorkload(ctx context.Context, kubeResourceIdentifier FullyQualifiedKubeResource) (*ExplorationResult, error)
}

type Printer interface {
	PrintForService(out io.Writer, serviceName *FullyQualifiedKubeResource, result *ExplorationResult) error
}

type ExplorationResult struct {
	Policies *Policies
}

type Policies struct {
	AccessControlPolicies []*v1alpha1.AccessControlPolicy
	TrafficPolicies       []*v1alpha1.TrafficPolicy
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
