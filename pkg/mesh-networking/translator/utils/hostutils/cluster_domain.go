package hostutils

import (
	"fmt"
	skv1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	skv1alpha1sets "github.com/solo-io/skv2/pkg/api/multicluster.solo.io/v1alpha1/sets"
	"github.com/solo-io/skv2/pkg/ezkube"
	"github.com/solo-io/smh/pkg/common/defaults"
)

const defaultClusterDomain = "cluster.local"

// ClusterDomainRegistry retrieves known cluster domain suffixes for
// registered clusters. Returns the default 'cluster.local' when
// domain cannot be found
type ClusterDomainRegistry interface {
	// get the domain suffix used by the cluster
	GetClusterDomain(clusterName string) string

	// get the local FQDN of a service in a given cluster.
	// this is the Kubernetes DNS name that clients within that cluster
	// can use to communicate to this cluster.
	GetLocalServiceFQDN(serviceRef ezkube.ResourceId) string

	// get the remote FQDN of a service in a given cluster.
	// this is the DNS name used by Service Mesh Hub
	// to establish cross-cluster connectivity.
	GetRemoteServiceFQDN(serviceRef ezkube.ResourceId) string

	// get the FQDN of a service which is being addressed as a
	// destination, e.g. for a traffic split or mirror policy.
	// this will either be the Local or Remote FQDN, depending on the
	// originating cluster.
	GetDestinationServiceFQDN(originatingCluster string, serviceRef ezkube.ResourceId) string
}

type clusterDomainRegistry struct {
	clusters skv1alpha1sets.KubernetesClusterSet
}

func NewClusterDomainRegistry(clusters skv1alpha1sets.KubernetesClusterSet) ClusterDomainRegistry {
	return &clusterDomainRegistry{clusters: clusters}
}

func (c *clusterDomainRegistry) GetClusterDomain(clusterName string) string {
	cluster, err := c.clusters.Find(&skv1.ClusterObjectRef{
		Name:      clusterName,
		Namespace: defaults.GetPodNamespace(),
	})
	if err != nil {
		return defaultClusterDomain
	}
	clusterDomain := cluster.Spec.ClusterDomain
	if clusterDomain == "" {
		clusterDomain = defaultClusterDomain
	}
	return clusterDomain
}

func (c *clusterDomainRegistry) GetLocalServiceFQDN(serviceRef ezkube.ResourceId) string {
	return fmt.Sprintf("%s.%s.svc.%s", serviceRef.GetName(), serviceRef.GetNamespace(), c.GetClusterDomain(serviceRef.GetClusterName()))
}

func (c *clusterDomainRegistry) GetRemoteServiceFQDN(serviceRef ezkube.ResourceId) string {
	return fmt.Sprintf("%s.%s.svc.%s", serviceRef.GetName(), serviceRef.GetNamespace(), serviceRef.GetClusterName())
}

func (c *clusterDomainRegistry) GetDestinationServiceFQDN(originatingCluster string, destination ezkube.ResourceId) string {
	if destination.GetClusterName() == originatingCluster {
		// hostname will use the cluster local domain if the destination is in the same cluster as the target MeshService
		return c.GetLocalServiceFQDN(destination)
	} else {
		// hostname will use the cross-cluster domain if the destination is in a different cluster than the target MeshService
		return c.GetRemoteServiceFQDN(destination)
	}
}