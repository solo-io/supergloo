package mocks_common

import (
	v1 "github.com/solo-io/mesh-projects/pkg/api/v1"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
)

type MockMeshReconciler struct {
	ReconcileCalledWith []v1.MeshList
}

func (r *MockMeshReconciler) Reconcile(namespace string, desiredResources v1.MeshList, transition v1.TransitionMeshFunc, opts clients.ListOpts) error {
	r.ReconcileCalledWith = append(r.ReconcileCalledWith, desiredResources)
	return nil
}

type MockMeshIngressReconciler struct {
	ReconcileCalledWith []v1.MeshIngressList
}

func (r *MockMeshIngressReconciler) Reconcile(namespace string, desiredResources v1.MeshIngressList, transition v1.TransitionMeshIngressFunc, opts clients.ListOpts) error {
	r.ReconcileCalledWith = append(r.ReconcileCalledWith, desiredResources)
	return nil
}
