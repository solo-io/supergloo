package istio

import (
	"context"

	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	smh_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/types"
	istio_security "github.com/solo-io/service-mesh-hub/pkg/api/istio/security/v1beta1"
	smh_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1/types"
	access_control_enforcer "github.com/solo-io/service-mesh-hub/pkg/common/access-control/enforcer"
	istio_federation "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/federation/resolver/meshes/istio"
	"github.com/solo-io/service-mesh-hub/pkg/common/kube"
	"github.com/solo-io/service-mesh-hub/pkg/common/kube/multicluster"
	istio_api_security "istio.io/api/security/v1beta1"
	"istio.io/api/type/v1beta1"
	client_security_v1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	EnforcerId                        = "istio_enforcer"
	GlobalAccessControlAuthPolicyName = "global-access-control"
	IngressGatewayAuthPolicy          = "ingress-policy"
)

type istioEnforcer struct {
	dynamicClientGetter     multicluster.DynamicClientGetter
	authPolicyClientFactory istio_security.AuthorizationPolicyClientFactory
}

type IstioEnforcer access_control_enforcer.AccessPolicyMeshEnforcer

func NewIstioEnforcer(
	dynamicClientGetter multicluster.DynamicClientGetter,
	authPolicyClientFactory istio_security.AuthorizationPolicyClientFactory,
) IstioEnforcer {
	return &istioEnforcer{
		authPolicyClientFactory: authPolicyClientFactory,
		dynamicClientGetter:     dynamicClientGetter,
	}
}

func (i *istioEnforcer) Name() string {
	return EnforcerId
}

func (i *istioEnforcer) ReconcileAccessControl(
	ctx context.Context,
	mesh *smh_discovery.Mesh,
	virtualMesh *smh_networking.VirtualMesh,
) error {
	if mesh.Spec.GetIstio1_6() == nil && mesh.Spec.GetIstio1_5() == nil {
		return nil
	}
	switch virtualMesh.Spec.GetEnforceAccessControl() {
	case types.VirtualMeshSpec_DISABLED, types.VirtualMeshSpec_MESH_DEFAULT:
		return i.stopEnforcing(ctx, mesh)
	case types.VirtualMeshSpec_ENABLED:
		return i.startEnforcing(ctx, mesh)
	}
	return nil
}

func (i *istioEnforcer) startEnforcing(ctx context.Context, mesh *smh_discovery.Mesh) error {
	installationNamespace := ""
	if mesh.Spec.GetIstio1_5() != nil {
		installationNamespace = mesh.Spec.GetIstio1_5().GetMetadata().GetInstallation().GetInstallationNamespace()
	} else {
		installationNamespace = mesh.Spec.GetIstio1_6().GetMetadata().GetInstallation().GetInstallationNamespace()
	}

	clientForCluster, err := i.dynamicClientGetter.GetClientForCluster(ctx, mesh.Spec.GetCluster().GetName())
	if err != nil {
		return err
	}
	authPolicyClient := i.authPolicyClientFactory(clientForCluster)
	if err := i.ensureIngressGatewayPolicy(ctx, installationNamespace, authPolicyClient); err != nil {
		return err
	}
	if err := i.ensureGlobalAuthPolicy(ctx, installationNamespace, authPolicyClient); err != nil {
		return err
	}
	return nil
}

func (i *istioEnforcer) stopEnforcing(ctx context.Context, mesh *smh_discovery.Mesh) error {
	installationNamespace := ""
	if mesh.Spec.GetIstio1_5() != nil {
		installationNamespace = mesh.Spec.GetIstio1_5().GetMetadata().GetInstallation().GetInstallationNamespace()
	} else {
		installationNamespace = mesh.Spec.GetIstio1_6().GetMetadata().GetInstallation().GetInstallationNamespace()
	}
	if err := i.stopEnforcingForMesh(ctx, installationNamespace, mesh); err != nil {
		return err
	}
	return nil
}

// returns nil if not an Istio installation
func (*istioEnforcer) getIstioInstallation(mesh *smh_discovery.Mesh) *smh_discovery_types.MeshSpec_MeshInstallation {
	var istioInstallation *smh_discovery_types.MeshSpec_MeshInstallation
	if mesh.Spec.GetIstio1_6() != nil {
		istioInstallation = mesh.Spec.GetIstio1_6().GetMetadata().GetInstallation()
	} else if mesh.Spec.GetIstio1_5() != nil {
		istioInstallation = mesh.Spec.GetIstio1_5().GetMetadata().GetInstallation()
	}

	return istioInstallation
}

func (i *istioEnforcer) ensureGlobalAuthPolicy(
	ctx context.Context,
	installationNamespace string,
	authPolicyClient istio_security.AuthorizationPolicyClient,
) error {
	// The following config denies all traffic in the mesh because it defaults to an ALLOW rule that doesn't match any requests,
	// thereby denying traffic unless explicitly allowed by the user through additional AuthorizationPolicies.
	// https://istio.io/docs/reference/config/security/authorization-policy/#AuthorizationPolicy
	globalAccessControlAuthPolicy := &client_security_v1beta1.AuthorizationPolicy{
		ObjectMeta: v1.ObjectMeta{
			Name:      GlobalAccessControlAuthPolicyName,
			Namespace: installationNamespace,
			Labels:    kube.OwnedBySMHLabel,
		},
		Spec: istio_api_security.AuthorizationPolicy{},
	}
	return authPolicyClient.UpsertAuthorizationPolicySpec(ctx, globalAccessControlAuthPolicy)
}

func (i *istioEnforcer) ensureIngressGatewayPolicy(
	ctx context.Context,
	installationNamespace string,
	authPolicyClient istio_security.AuthorizationPolicyClient,
) error {
	// The following config allows all traffic into the "istio-ingressgateway", which in Service Mesh Hub is
	// the gateway used for multi cluster traffic. Authorization is then handled by the individual workloads which traffic
	// is forwarded to.
	ingressGatewayAllowAllPolicy := &client_security_v1beta1.AuthorizationPolicy{
		ObjectMeta: v1.ObjectMeta{
			Name:      IngressGatewayAuthPolicy,
			Namespace: installationNamespace,
			Labels:    kube.OwnedBySMHLabel,
		},
		Spec: istio_api_security.AuthorizationPolicy{
			Action: istio_api_security.AuthorizationPolicy_ALLOW,
			// According to the Istio docs on AuthorizationPolicy a single empty rule allows all traffic
			// https://istio.io/docs/reference/config/security/authorization-policy/#AuthorizationPolicy
			Rules: []*istio_api_security.Rule{{}},
			Selector: &v1beta1.WorkloadSelector{
				MatchLabels: istio_federation.BuildGatewayWorkloadSelector(),
			},
		},
	}
	return authPolicyClient.UpsertAuthorizationPolicySpec(ctx, ingressGatewayAllowAllPolicy)
}

func (i *istioEnforcer) stopEnforcingForMesh(
	ctx context.Context,
	installationNamespace string,
	mesh *smh_discovery.Mesh,
) error {
	clientForCluster, err := i.dynamicClientGetter.GetClientForCluster(ctx, mesh.Spec.GetCluster().GetName())
	if err != nil {
		return err
	}
	authPolicyClient := i.authPolicyClientFactory(clientForCluster)
	globalAuthPolicyKey := client.ObjectKey{
		Name:      GlobalAccessControlAuthPolicyName,
		Namespace: installationNamespace,
	}
	if err = i.deleteIfExists(ctx, globalAuthPolicyKey, authPolicyClient); err != nil {
		return err
	}
	gatewayAuthPolicyKey := client.ObjectKey{
		Name:      IngressGatewayAuthPolicy,
		Namespace: installationNamespace,
	}
	return i.deleteIfExists(ctx, gatewayAuthPolicyKey, authPolicyClient)
}

func (i *istioEnforcer) deleteIfExists(
	ctx context.Context,
	objKey client.ObjectKey,
	policyClient istio_security.AuthorizationPolicyClient,
) error {
	_, err := policyClient.GetAuthorizationPolicy(ctx, objKey)
	if err != nil {
		// If it cannot be found, do not attempt to delete, and return no error
		if errors.IsNotFound(err) {
			return nil
		}
		return err
	} else {
		return policyClient.DeleteAuthorizationPolicy(ctx, objKey)
	}
}
