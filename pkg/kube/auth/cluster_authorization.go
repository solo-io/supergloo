package auth

import (
	"context"

	zephyr_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	k8s_rbac_types "k8s.io/api/rbac/v1"
	k8s_meta_types "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

var (
	// visible for testing
	ServiceAccountRoles = []*k8s_rbac_types.ClusterRole{{
		ObjectMeta: k8s_meta_types.ObjectMeta{Name: "cluster-admin"},
	}}
)

type clusterAuthorization struct {
	configCreator          RemoteAuthorityConfigCreator
	remoteAuthorityManager RemoteAuthorityManager
}

type ClusterAuthorizationFactory func(
	configCreator RemoteAuthorityConfigCreator,
	remoteAuthorityManager RemoteAuthorityManager,
) ClusterAuthorization

func ClusterAuthorizationFactoryProvider() ClusterAuthorizationFactory {
	return NewClusterAuthorization
}

func NewClusterAuthorization(
	configCreator RemoteAuthorityConfigCreator,
	remoteAuthorityManager RemoteAuthorityManager) ClusterAuthorization {
	return &clusterAuthorization{configCreator, remoteAuthorityManager}
}

func (c *clusterAuthorization) BuildRemoteBearerToken(
	ctx context.Context,
	targetClusterCfg *rest.Config,
	serviceAccountRef *zephyr_core_types.ResourceRef,
) (bearerToken string, err error) {
	_, err = c.remoteAuthorityManager.ApplyRemoteServiceAccount(ctx, serviceAccountRef, ServiceAccountRoles)
	if err != nil {
		return "", err
	}

	saConfig, err := c.configCreator.ConfigFromRemoteServiceAccount(ctx, targetClusterCfg, serviceAccountRef)
	if err != nil {
		return "", err
	}

	// we only want the bearer token for that service account
	return saConfig.BearerToken, nil
}
