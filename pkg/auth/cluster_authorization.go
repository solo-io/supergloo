package auth

import (
	"context"

	"github.com/solo-io/mesh-projects/pkg/api/core.zephyr.solo.io/v1alpha1/types"
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

type clusterAuthorization struct {
	configCreator          RemoteAuthorityConfigCreator
	remoteAuthorityManager RemoteAuthorityManager
}

func NewClusterAuthorization(
	configCreator RemoteAuthorityConfigCreator,
	remoteAuthorityManager RemoteAuthorityManager) ClusterAuthorization {
	return &clusterAuthorization{configCreator, remoteAuthorityManager}
}

func (c *clusterAuthorization) CreateAuthConfigForCluster(
	ctx context.Context,
	targetClusterCfg *rest.Config,
	serviceAccountRef *types.ResourceRef,
) (*rest.Config, error) {
	_, err := c.remoteAuthorityManager.ApplyRemoteServiceAccount(ctx, serviceAccountRef, ServiceAccountRoles)
	if err != nil {
		return nil, err
	}

	return c.configCreator.ConfigFromRemoteServiceAccount(ctx, targetClusterCfg, serviceAccountRef)
}
