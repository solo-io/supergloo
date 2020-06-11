package appmesh

import (
	"context"

	access_policy_enforcer "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/access/access-control-enforcer"

	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	smh_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/common/aws/translation"
)

const (
	EnforcerId = "appmesh_enforcer"
)

type appmeshEnforcer struct {
	appmeshTranslationReconciler translation.AppmeshTranslationReconciler
}

type AppmeshEnforcer access_policy_enforcer.AccessPolicyMeshEnforcer

func NewAppmeshEnforcer(
	appmeshTranslationReconciler translation.AppmeshTranslationReconciler,
) AppmeshEnforcer {
	return &appmeshEnforcer{
		appmeshTranslationReconciler: appmeshTranslationReconciler,
	}
}

func (a *appmeshEnforcer) Name() string {
	return EnforcerId
}

func (a *appmeshEnforcer) ReconcileAccessControl(
	ctx context.Context,
	mesh *smh_discovery.Mesh,
	virtualMesh *smh_networking.VirtualMesh,
) error {
	if mesh.Spec.GetAwsAppMesh() == nil {
		return nil
	}
	return a.appmeshTranslationReconciler.Reconcile(ctx, mesh, virtualMesh)
}
