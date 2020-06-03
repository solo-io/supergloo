package appmesh

import (
	"context"

	access_control_enforcer "github.com/solo-io/service-mesh-hub/pkg/access-control/enforcer"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/aws/translation"
)

const (
	EnforcerId = "appmesh_enforcer"
)

type appmeshEnforcer struct {
	appmeshTranslationReconciler translation.AppmeshTranslationReconciler
}

type AppmeshEnforcer access_control_enforcer.AccessPolicyMeshEnforcer

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
	mesh *zephyr_discovery.Mesh,
	virtualMesh *zephyr_networking.VirtualMesh,
) error {
	if mesh.Spec.GetAwsAppMesh() == nil {
		return nil
	}
	return a.appmeshTranslationReconciler.Reconcile(ctx, mesh, virtualMesh)
}
