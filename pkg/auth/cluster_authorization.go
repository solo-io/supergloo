package auth

import (
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	rbacapiv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

var (
	// visible for testing
	ServiceAccountRoles = []*rbacapiv1.ClusterRole{{
		ObjectMeta: metav1.ObjectMeta{Name: "cluster-admin"},
	}}
)

// Given a way to authorize to a cluster, produce a new config that can authorize to that same cluster
// using a newly-created service account token in that cluster.
// Creates a service account in the target cluster with the name/namespace of `serviceAccountRef` and cluster-admin permissions
//go:generate mockgen -destination ./mocks/mock_cluster_authorization.go github.com/solo-io/mesh-projects/pkg/auth ClusterAuthorization
type ClusterAuthorization interface {
	CreateAuthConfigForCluster(targetClusterCfg *rest.Config, serviceAccountRef *core.ResourceRef) (*rest.Config, error)
}

type clusterAuthorization struct {
	configCreator          RemoteAuthorityConfigCreator
	remoteAuthorityManager RemoteAuthorityManager
}

func NewClusterAuthorization(configCreator RemoteAuthorityConfigCreator, remoteAuthorityManager RemoteAuthorityManager) ClusterAuthorization {
	return &clusterAuthorization{configCreator, remoteAuthorityManager}
}

func (c *clusterAuthorization) CreateAuthConfigForCluster(targetClusterCfg *rest.Config, serviceAccountRef *core.ResourceRef) (*rest.Config, error) {
	_, err := c.remoteAuthorityManager.CreateRemoteServiceAccount(targetClusterCfg, serviceAccountRef, ServiceAccountRoles)
	if err != nil {
		return nil, err
	}

	return c.configCreator.ConfigFromRemoteServiceAccount(targetClusterCfg, serviceAccountRef)
}
