package auth

import (
	"context"

	core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	k8s_core "github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/core/v1"
	k8s_core_types "k8s.io/api/core/v1"
	k8s_rbac_types "k8s.io/api/rbac/v1"
	"k8s.io/client-go/rest"
)

//go:generate mockgen -source ./interfaces.go -destination ./mocks/mock_auth.go

type RbacClient interface {
	// bind the given roles to the target service account at cluster scope
	BindClusterRolesToServiceAccount(targetServiceAccount *k8s_core_types.ServiceAccount, roles []*k8s_rbac_types.ClusterRole) error
}

// create a kube config that can authorize to the target cluster as the service account from that target cluster
type RemoteAuthorityConfigCreator interface {

	// Returns a `*rest.Config` that points to the same cluster as `targetClusterCfg`, but authorizes itself using the
	// bearer token belonging to the service account at `serviceAccountRef` in the target cluster
	//
	// NB: This function blocks the current go routine for up to 6 seconds while waiting for the service account's secret
	// to appear, by performing an exponential backoff retry loop
	ConfigFromRemoteServiceAccount(
		ctx context.Context,
		targetClusterCfg *rest.Config,
		serviceAccountRef *core_types.ResourceRef,
	) (*rest.Config, error)
}

// Given a way to authorize to a cluster, produce a new config that can authorize to that same cluster
// using a newly-created service account token in that cluster.
// Creates a service account in the target cluster with the name/namespace of `serviceAccountRef` and cluster-admin permissions
type ClusterAuthorization interface {
	CreateAuthConfigForCluster(
		ctx context.Context,
		targetClusterCfg *rest.Config,
		serviceAccountRef *core_types.ResourceRef,
	) (*rest.Config, error)
}

// Create a service account on a cluster that `targetClusterCfg` can reach
// Set up that service account with the indicated cluster roles
type RemoteAuthorityManager interface {
	// creates a new service account in the cluster pointed to by the cfg at the name/namespace indicated by the ResourceRef,
	// and assigns it the given ClusterRoles
	// NB: if role assignment fails, the service account is left in the cluster; this is not an atomic operation
	ApplyRemoteServiceAccount(
		ctx context.Context,
		newServiceAccountRef *core_types.ResourceRef,
		roles []*k8s_rbac_types.ClusterRole,
	) (*k8s_core_types.ServiceAccount, error)
}

type Clients struct {
	ServiceAccountClient k8s_core.ServiceAccountClient
	RbacClient           RbacClient
	SecretClient         k8s_core.SecretClient
}
