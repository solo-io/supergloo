// Code generated by MockGen. DO NOT EDIT.
// Source: ./reconcilers.go

// Package mock_controller is a generated GoMock package.
package mock_controller

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	v1alpha1 "github.com/solo-io/gloo-mesh/pkg/api/observability.enterprise.mesh.gloo.solo.io/v1alpha1"
	controller "github.com/solo-io/gloo-mesh/pkg/api/observability.enterprise.mesh.gloo.solo.io/v1alpha1/controller"
	reconcile "github.com/solo-io/skv2/pkg/reconcile"
	predicate "sigs.k8s.io/controller-runtime/pkg/predicate"
)

// MockAccessLogCollectionReconciler is a mock of AccessLogCollectionReconciler interface
type MockAccessLogCollectionReconciler struct {
	ctrl     *gomock.Controller
	recorder *MockAccessLogCollectionReconcilerMockRecorder
}

// MockAccessLogCollectionReconcilerMockRecorder is the mock recorder for MockAccessLogCollectionReconciler
type MockAccessLogCollectionReconcilerMockRecorder struct {
	mock *MockAccessLogCollectionReconciler
}

// NewMockAccessLogCollectionReconciler creates a new mock instance
func NewMockAccessLogCollectionReconciler(ctrl *gomock.Controller) *MockAccessLogCollectionReconciler {
	mock := &MockAccessLogCollectionReconciler{ctrl: ctrl}
	mock.recorder = &MockAccessLogCollectionReconcilerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockAccessLogCollectionReconciler) EXPECT() *MockAccessLogCollectionReconcilerMockRecorder {
	return m.recorder
}

// ReconcileAccessLogCollection mocks base method
func (m *MockAccessLogCollectionReconciler) ReconcileAccessLogCollection(obj *v1alpha1.AccessLogCollection) (reconcile.Result, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ReconcileAccessLogCollection", obj)
	ret0, _ := ret[0].(reconcile.Result)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ReconcileAccessLogCollection indicates an expected call of ReconcileAccessLogCollection
func (mr *MockAccessLogCollectionReconcilerMockRecorder) ReconcileAccessLogCollection(obj interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ReconcileAccessLogCollection", reflect.TypeOf((*MockAccessLogCollectionReconciler)(nil).ReconcileAccessLogCollection), obj)
}

// MockAccessLogCollectionDeletionReconciler is a mock of AccessLogCollectionDeletionReconciler interface
type MockAccessLogCollectionDeletionReconciler struct {
	ctrl     *gomock.Controller
	recorder *MockAccessLogCollectionDeletionReconcilerMockRecorder
}

// MockAccessLogCollectionDeletionReconcilerMockRecorder is the mock recorder for MockAccessLogCollectionDeletionReconciler
type MockAccessLogCollectionDeletionReconcilerMockRecorder struct {
	mock *MockAccessLogCollectionDeletionReconciler
}

// NewMockAccessLogCollectionDeletionReconciler creates a new mock instance
func NewMockAccessLogCollectionDeletionReconciler(ctrl *gomock.Controller) *MockAccessLogCollectionDeletionReconciler {
	mock := &MockAccessLogCollectionDeletionReconciler{ctrl: ctrl}
	mock.recorder = &MockAccessLogCollectionDeletionReconcilerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockAccessLogCollectionDeletionReconciler) EXPECT() *MockAccessLogCollectionDeletionReconcilerMockRecorder {
	return m.recorder
}

// ReconcileAccessLogCollectionDeletion mocks base method
func (m *MockAccessLogCollectionDeletionReconciler) ReconcileAccessLogCollectionDeletion(req reconcile.Request) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ReconcileAccessLogCollectionDeletion", req)
	ret0, _ := ret[0].(error)
	return ret0
}

// ReconcileAccessLogCollectionDeletion indicates an expected call of ReconcileAccessLogCollectionDeletion
func (mr *MockAccessLogCollectionDeletionReconcilerMockRecorder) ReconcileAccessLogCollectionDeletion(req interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ReconcileAccessLogCollectionDeletion", reflect.TypeOf((*MockAccessLogCollectionDeletionReconciler)(nil).ReconcileAccessLogCollectionDeletion), req)
}

// MockAccessLogCollectionFinalizer is a mock of AccessLogCollectionFinalizer interface
type MockAccessLogCollectionFinalizer struct {
	ctrl     *gomock.Controller
	recorder *MockAccessLogCollectionFinalizerMockRecorder
}

// MockAccessLogCollectionFinalizerMockRecorder is the mock recorder for MockAccessLogCollectionFinalizer
type MockAccessLogCollectionFinalizerMockRecorder struct {
	mock *MockAccessLogCollectionFinalizer
}

// NewMockAccessLogCollectionFinalizer creates a new mock instance
func NewMockAccessLogCollectionFinalizer(ctrl *gomock.Controller) *MockAccessLogCollectionFinalizer {
	mock := &MockAccessLogCollectionFinalizer{ctrl: ctrl}
	mock.recorder = &MockAccessLogCollectionFinalizerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockAccessLogCollectionFinalizer) EXPECT() *MockAccessLogCollectionFinalizerMockRecorder {
	return m.recorder
}

// ReconcileAccessLogCollection mocks base method
func (m *MockAccessLogCollectionFinalizer) ReconcileAccessLogCollection(obj *v1alpha1.AccessLogCollection) (reconcile.Result, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ReconcileAccessLogCollection", obj)
	ret0, _ := ret[0].(reconcile.Result)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ReconcileAccessLogCollection indicates an expected call of ReconcileAccessLogCollection
func (mr *MockAccessLogCollectionFinalizerMockRecorder) ReconcileAccessLogCollection(obj interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ReconcileAccessLogCollection", reflect.TypeOf((*MockAccessLogCollectionFinalizer)(nil).ReconcileAccessLogCollection), obj)
}

// AccessLogCollectionFinalizerName mocks base method
func (m *MockAccessLogCollectionFinalizer) AccessLogCollectionFinalizerName() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AccessLogCollectionFinalizerName")
	ret0, _ := ret[0].(string)
	return ret0
}

// AccessLogCollectionFinalizerName indicates an expected call of AccessLogCollectionFinalizerName
func (mr *MockAccessLogCollectionFinalizerMockRecorder) AccessLogCollectionFinalizerName() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AccessLogCollectionFinalizerName", reflect.TypeOf((*MockAccessLogCollectionFinalizer)(nil).AccessLogCollectionFinalizerName))
}

// FinalizeAccessLogCollection mocks base method
func (m *MockAccessLogCollectionFinalizer) FinalizeAccessLogCollection(obj *v1alpha1.AccessLogCollection) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FinalizeAccessLogCollection", obj)
	ret0, _ := ret[0].(error)
	return ret0
}

// FinalizeAccessLogCollection indicates an expected call of FinalizeAccessLogCollection
func (mr *MockAccessLogCollectionFinalizerMockRecorder) FinalizeAccessLogCollection(obj interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FinalizeAccessLogCollection", reflect.TypeOf((*MockAccessLogCollectionFinalizer)(nil).FinalizeAccessLogCollection), obj)
}

// MockAccessLogCollectionReconcileLoop is a mock of AccessLogCollectionReconcileLoop interface
type MockAccessLogCollectionReconcileLoop struct {
	ctrl     *gomock.Controller
	recorder *MockAccessLogCollectionReconcileLoopMockRecorder
}

// MockAccessLogCollectionReconcileLoopMockRecorder is the mock recorder for MockAccessLogCollectionReconcileLoop
type MockAccessLogCollectionReconcileLoopMockRecorder struct {
	mock *MockAccessLogCollectionReconcileLoop
}

// NewMockAccessLogCollectionReconcileLoop creates a new mock instance
func NewMockAccessLogCollectionReconcileLoop(ctrl *gomock.Controller) *MockAccessLogCollectionReconcileLoop {
	mock := &MockAccessLogCollectionReconcileLoop{ctrl: ctrl}
	mock.recorder = &MockAccessLogCollectionReconcileLoopMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockAccessLogCollectionReconcileLoop) EXPECT() *MockAccessLogCollectionReconcileLoopMockRecorder {
	return m.recorder
}

// RunAccessLogCollectionReconciler mocks base method
func (m *MockAccessLogCollectionReconcileLoop) RunAccessLogCollectionReconciler(ctx context.Context, rec controller.AccessLogCollectionReconciler, predicates ...predicate.Predicate) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, rec}
	for _, a := range predicates {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "RunAccessLogCollectionReconciler", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// RunAccessLogCollectionReconciler indicates an expected call of RunAccessLogCollectionReconciler
func (mr *MockAccessLogCollectionReconcileLoopMockRecorder) RunAccessLogCollectionReconciler(ctx, rec interface{}, predicates ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, rec}, predicates...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RunAccessLogCollectionReconciler", reflect.TypeOf((*MockAccessLogCollectionReconcileLoop)(nil).RunAccessLogCollectionReconciler), varargs...)
}