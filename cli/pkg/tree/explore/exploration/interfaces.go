package exploration

import (
	"context"
	"fmt"
	"io"

	"github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
)

type ResourceExplorer interface {
	ExploreService(ctx context.Context, kubeResourceIdentifier FullyQualifiedKubeResource) (*ExplorationResult, error)
	ExploreWorkload(ctx context.Context, kubeResourceIdentifier FullyQualifiedKubeResource) (*ExplorationResult, error)
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
