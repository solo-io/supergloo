package hostutils

import (
	"fmt"

	v1alpha2sets "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1alpha2/sets"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1alpha2"
	"github.com/solo-io/gloo-mesh/pkg/common/defaults"
	"github.com/solo-io/gloo-mesh/pkg/mesh-discovery/translation/utils"
	skv1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	skv1alpha1sets "github.com/solo-io/skv2/pkg/api/multicluster.solo.io/v1alpha1/sets"
	"github.com/solo-io/skv2/pkg/ezkube"
)

//go:generate mockgen -source ./cluster_domain.go -destination mocks/cluster_domain_mocks.go

const (
	// this suffix is used as the default for federated fqdns; this is due to the
	// fact that istio Coredns comes with the *.global suffix already configured:
	// https://istio.io/latest/docs/setup/install/multicluster/gateways/
	defaultHostnameSuffix = "global"
)

// ClusterDomainRegistry retrieves known cluster domain suffixes for
// registered clusters. Returns the default 'cluster.local' when
// domain cannot be found
type ClusterDomainRegistry interface {
	// get the domain suffix used by the cluster
	GetClusterDomain(clusterName string) string

	// get the local FQDN of a service in a given cluster.
	// this is the Kubernetes DNS name that clients within that cluster
	// can use to communicate to this cluster.
	GetLocalFQDN(serviceRef ezkube.ClusterResourceId) string

	// get the remote FQDN of a service in a given cluster.
	// this is the DNS name used by Gloo Mesh
	// to establish cross-cluster connectivity.
	GetFederatedFQDN(serviceRef ezkube.ClusterResourceId) string

	// get the FQDN of a service which is being addressed as a
	// destination, e.g. for a traffic split or mirror policy.
	// this will either be the Local or Remote FQDN, depending on the
	// originating cluster.
	GetDestinationFQDN(originatingCluster string, serviceRef ezkube.ClusterResourceId) string
}

type clusterDomainRegistry struct {
	clusters       skv1alpha1sets.KubernetesClusterSet
	trafficTargets v1alpha2sets.TrafficTargetSet
}

func NewClusterDomainRegistry(
	clusters skv1alpha1sets.KubernetesClusterSet,
	trafficTargets v1alpha2sets.TrafficTargetSet,
) ClusterDomainRegistry {
	return &clusterDomainRegistry{
		clusters:       clusters,
		trafficTargets: trafficTargets,
	}
}

func (c *clusterDomainRegistry) GetClusterDomain(clusterName string) string {
	cluster, err := c.clusters.Find(&skv1.ObjectRef{
		Name:      clusterName,
		Namespace: defaults.GetPodNamespace(),
	})
	if err != nil {
		return defaults.DefaultClusterDomain
	}
	clusterDomain := cluster.Spec.ClusterDomain
	if clusterDomain == "" {
		clusterDomain = defaults.DefaultClusterDomain
	}
	return clusterDomain
}

func (c *clusterDomainRegistry) GetLocalFQDN(trafficTargetRef ezkube.ClusterResourceId) string {
	return fmt.Sprintf("%s.%s.svc.%s", trafficTargetRef.GetName(), trafficTargetRef.GetNamespace(), c.GetClusterDomain(trafficTargetRef.GetClusterName()))
}

func (c *clusterDomainRegistry) GetFederatedFQDN(trafficTargetRef ezkube.ClusterResourceId) string {
	trafficTarget, err := c.trafficTargets.Find(&skv1.ObjectRef{
		Name:      utils.DiscoveredResourceName(trafficTargetRef),
		Namespace: defaults.GetPodNamespace(),
	})
	if err != nil || trafficTarget.Status.GetAppliedFederation().GetFederatedHostname() == "" {
		return fmt.Sprintf("%s.%s.svc.%s.%v", trafficTargetRef.GetName(), trafficTargetRef.GetNamespace(), trafficTargetRef.GetClusterName(), defaultHostnameSuffix)
	} else {
		return trafficTarget.Status.GetAppliedFederation().GetFederatedHostname()
	}
}

func (c *clusterDomainRegistry) GetDestinationFQDN(originatingCluster string, destination ezkube.ClusterResourceId) string {
	if destination.GetClusterName() == originatingCluster {
		// hostname will use the cluster local domain if the destination is in the same cluster as the target TrafficTarget
		return c.GetLocalFQDN(destination)
	} else {
		// hostname will use the cross-cluster domain if the destination is in a different cluster than the target TrafficTarget
		return c.GetFederatedFQDN(destination)
	}
}

// Construct a federated FQDN for the given service, using the provided hostname suffix if provided, otherwise use default suffix.
func BuildFederatedFQDN(serviceRef ezkube.ClusterResourceId, virtualMeshSpec *v1alpha2.VirtualMeshSpec) string {
	return fmt.Sprintf(
		"%s.%s.svc.%s.%v",
		serviceRef.GetName(),
		serviceRef.GetNamespace(),
		serviceRef.GetClusterName(),
		GetFederatedHostnameSuffix(virtualMeshSpec),
	)
}

func GetFederatedHostnameSuffix(virtualMeshSpec *v1alpha2.VirtualMeshSpec) string {
	federatedHostnameSuffix := virtualMeshSpec.GetFederation().GetHostnameSuffix()
	if federatedHostnameSuffix == "" {
		federatedHostnameSuffix = defaultHostnameSuffix
	}
	return federatedHostnameSuffix
}
