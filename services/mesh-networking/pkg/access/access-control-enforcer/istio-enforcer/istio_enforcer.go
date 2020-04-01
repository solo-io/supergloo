package istio_enforcer

import (
	"context"

	discovery_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/mesh-projects/pkg/clients/istio/security"
	"github.com/solo-io/mesh-projects/services/common/constants"
	mc_manager "github.com/solo-io/mesh-projects/services/common/multicluster/manager"
	global_access_control_enforcer "github.com/solo-io/mesh-projects/services/mesh-networking/pkg/access/access-control-enforcer"
	security_v1beta1 "istio.io/api/security/v1beta1"
	client_security_v1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const GlobalAccessControlAuthPolicyName = "global-access-control"

type istioEnforcer struct {
	dynamicClientGetter     mc_manager.DynamicClientGetter
	authPolicyClientFactory security.AuthorizationPolicyClientFactory
}

type IstioEnforcer global_access_control_enforcer.AccessPolicyMeshEnforcer

func NewIstioEnforcer(
	dynamicClientGetter mc_manager.DynamicClientGetter,
	authPolicyClientFactory security.AuthorizationPolicyClientFactory,
) IstioEnforcer {
	return &istioEnforcer{
		authPolicyClientFactory: authPolicyClientFactory,
		dynamicClientGetter:     dynamicClientGetter,
	}
}

func (i *istioEnforcer) StartEnforcing(ctx context.Context, meshes []*discovery_v1alpha1.Mesh) error {
	for _, mesh := range meshes {
		if mesh.Spec.GetIstio() == nil {
			continue
		}
		if err := i.startEnforcingForMesh(ctx, mesh); err != nil {
			return err
		}
	}
	return nil
}

func (i *istioEnforcer) StopEnforcing(ctx context.Context, meshes []*discovery_v1alpha1.Mesh) error {
	for _, mesh := range meshes {
		if mesh.Spec.GetIstio() == nil {
			continue
		}
		if err := i.stopEnforcingForMesh(ctx, mesh); err != nil {
			return err
		}
	}
	return nil
}

func (i *istioEnforcer) startEnforcingForMesh(
	ctx context.Context,
	mesh *discovery_v1alpha1.Mesh,
) error {
	// The following config denies all traffic in the mesh because it defaults to an ALLOW rule that doesn't match any requests,
	// thereby denying traffic unless explicitly allowed by the user through additional AuthorizationPolicies.
	globalAccessControlAuthPolicy := &client_security_v1beta1.AuthorizationPolicy{
		ObjectMeta: v1.ObjectMeta{
			Name:      GlobalAccessControlAuthPolicyName,
			Namespace: mesh.Spec.GetIstio().GetInstallation().GetInstallationNamespace(),
			Labels:    constants.OwnedBySMHLabel,
		},
		Spec: security_v1beta1.AuthorizationPolicy{},
	}
	clientForCluster, err := i.dynamicClientGetter.GetClientForCluster(mesh.Spec.GetCluster().GetName())
	if err != nil {
		return err
	}
	authPolicyClient := i.authPolicyClientFactory(clientForCluster)
	err = authPolicyClient.UpsertSpec(ctx, globalAccessControlAuthPolicy)
	if err != nil {
		return err
	}
	return nil
}

func (i *istioEnforcer) stopEnforcingForMesh(
	ctx context.Context,
	mesh *discovery_v1alpha1.Mesh,
) error {
	clientForCluster, err := i.dynamicClientGetter.GetClientForCluster(mesh.Spec.GetCluster().GetName())
	if err != nil {
		return err
	}
	authPolicyClient := i.authPolicyClientFactory(clientForCluster)
	authPolicyKey := client.ObjectKey{
		Name:      GlobalAccessControlAuthPolicyName,
		Namespace: mesh.Spec.GetIstio().GetInstallation().GetInstallationNamespace(),
	}
	_, err = authPolicyClient.Get(ctx, authPolicyKey)
	if errors.IsNotFound(err) {
		return nil
	} else {
		return authPolicyClient.Delete(ctx, authPolicyKey)
	}
}
