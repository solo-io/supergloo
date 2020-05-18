package translation_framework

import (
	"context"

	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/reconciliation"
)

func NewTranslationReconciler(meshServiceClient *zephyr_discovery.MeshServiceClient) reconciliation.Reconciler {
	return &translationReconciler{
		meshServiceClient: meshServiceClient,
	}
}

type translationReconciler struct {
	meshServiceClient *zephyr_discovery.MeshServiceClient
}

func (t *translationReconciler) GetName() string {
	return "traffic-policy-translation-reconciler"
}

func (t *translationReconciler) Reconcile(context.Context) error {
	
}

