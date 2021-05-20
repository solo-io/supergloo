// Code generated by MockGen. DO NOT EDIT.
// Source: ./reconcilers.go

// Package mock_controller is a generated GoMock package.
package mock_controller

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	v1 "github.com/solo-io/gloo-mesh/pkg/api/observability.enterprise.mesh.gloo.solo.io/v1"
	controller "github.com/solo-io/gloo-mesh/pkg/api/observability.enterprise.mesh.gloo.solo.io/v1/controller"
	reconcile "github.com/solo-io/skv2/pkg/reconcile"
	predicate "sigs.k8s.io/controller-runtime/pkg/predicate"
)

// MockAccessLogRecordReconciler is a mock of AccessLogRecordReconciler interface.
type MockAccessLogRecordReconciler struct {
	ctrl     *gomock.Controller
	recorder *MockAccessLogRecordReconcilerMockRecorder
}

// MockAccessLogRecordReconcilerMockRecorder is the mock recorder for MockAccessLogRecordReconciler.
type MockAccessLogRecordReconcilerMockRecorder struct {
	mock *MockAccessLogRecordReconciler
}

// NewMockAccessLogRecordReconciler creates a new mock instance.
func NewMockAccessLogRecordReconciler(ctrl *gomock.Controller) *MockAccessLogRecordReconciler {
	mock := &MockAccessLogRecordReconciler{ctrl: ctrl}
	mock.recorder = &MockAccessLogRecordReconcilerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockAccessLogRecordReconciler) EXPECT() *MockAccessLogRecordReconcilerMockRecorder {
	return m.recorder
}

// ReconcileAccessLogRecord mocks base method.
func (m *MockAccessLogRecordReconciler) ReconcileAccessLogRecord(obj *v1.AccessLogRecord) (reconcile.Result, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ReconcileAccessLogRecord", obj)
	ret0, _ := ret[0].(reconcile.Result)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ReconcileAccessLogRecord indicates an expected call of ReconcileAccessLogRecord.
func (mr *MockAccessLogRecordReconcilerMockRecorder) ReconcileAccessLogRecord(obj interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ReconcileAccessLogRecord", reflect.TypeOf((*MockAccessLogRecordReconciler)(nil).ReconcileAccessLogRecord), obj)
}

// MockAccessLogRecordDeletionReconciler is a mock of AccessLogRecordDeletionReconciler interface.
type MockAccessLogRecordDeletionReconciler struct {
	ctrl     *gomock.Controller
	recorder *MockAccessLogRecordDeletionReconcilerMockRecorder
}

// MockAccessLogRecordDeletionReconcilerMockRecorder is the mock recorder for MockAccessLogRecordDeletionReconciler.
type MockAccessLogRecordDeletionReconcilerMockRecorder struct {
	mock *MockAccessLogRecordDeletionReconciler
}

// NewMockAccessLogRecordDeletionReconciler creates a new mock instance.
func NewMockAccessLogRecordDeletionReconciler(ctrl *gomock.Controller) *MockAccessLogRecordDeletionReconciler {
	mock := &MockAccessLogRecordDeletionReconciler{ctrl: ctrl}
	mock.recorder = &MockAccessLogRecordDeletionReconcilerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockAccessLogRecordDeletionReconciler) EXPECT() *MockAccessLogRecordDeletionReconcilerMockRecorder {
	return m.recorder
}

// ReconcileAccessLogRecordDeletion mocks base method.
func (m *MockAccessLogRecordDeletionReconciler) ReconcileAccessLogRecordDeletion(req reconcile.Request) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ReconcileAccessLogRecordDeletion", req)
	ret0, _ := ret[0].(error)
	return ret0
}

// ReconcileAccessLogRecordDeletion indicates an expected call of ReconcileAccessLogRecordDeletion.
func (mr *MockAccessLogRecordDeletionReconcilerMockRecorder) ReconcileAccessLogRecordDeletion(req interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ReconcileAccessLogRecordDeletion", reflect.TypeOf((*MockAccessLogRecordDeletionReconciler)(nil).ReconcileAccessLogRecordDeletion), req)
}

// MockAccessLogRecordFinalizer is a mock of AccessLogRecordFinalizer interface.
type MockAccessLogRecordFinalizer struct {
	ctrl     *gomock.Controller
	recorder *MockAccessLogRecordFinalizerMockRecorder
}

// MockAccessLogRecordFinalizerMockRecorder is the mock recorder for MockAccessLogRecordFinalizer.
type MockAccessLogRecordFinalizerMockRecorder struct {
	mock *MockAccessLogRecordFinalizer
}

// NewMockAccessLogRecordFinalizer creates a new mock instance.
func NewMockAccessLogRecordFinalizer(ctrl *gomock.Controller) *MockAccessLogRecordFinalizer {
	mock := &MockAccessLogRecordFinalizer{ctrl: ctrl}
	mock.recorder = &MockAccessLogRecordFinalizerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockAccessLogRecordFinalizer) EXPECT() *MockAccessLogRecordFinalizerMockRecorder {
	return m.recorder
}

// AccessLogRecordFinalizerName mocks base method.
func (m *MockAccessLogRecordFinalizer) AccessLogRecordFinalizerName() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AccessLogRecordFinalizerName")
	ret0, _ := ret[0].(string)
	return ret0
}

// AccessLogRecordFinalizerName indicates an expected call of AccessLogRecordFinalizerName.
func (mr *MockAccessLogRecordFinalizerMockRecorder) AccessLogRecordFinalizerName() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AccessLogRecordFinalizerName", reflect.TypeOf((*MockAccessLogRecordFinalizer)(nil).AccessLogRecordFinalizerName))
}

// FinalizeAccessLogRecord mocks base method.
func (m *MockAccessLogRecordFinalizer) FinalizeAccessLogRecord(obj *v1.AccessLogRecord) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FinalizeAccessLogRecord", obj)
	ret0, _ := ret[0].(error)
	return ret0
}

// FinalizeAccessLogRecord indicates an expected call of FinalizeAccessLogRecord.
func (mr *MockAccessLogRecordFinalizerMockRecorder) FinalizeAccessLogRecord(obj interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FinalizeAccessLogRecord", reflect.TypeOf((*MockAccessLogRecordFinalizer)(nil).FinalizeAccessLogRecord), obj)
}

// ReconcileAccessLogRecord mocks base method.
func (m *MockAccessLogRecordFinalizer) ReconcileAccessLogRecord(obj *v1.AccessLogRecord) (reconcile.Result, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ReconcileAccessLogRecord", obj)
	ret0, _ := ret[0].(reconcile.Result)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ReconcileAccessLogRecord indicates an expected call of ReconcileAccessLogRecord.
func (mr *MockAccessLogRecordFinalizerMockRecorder) ReconcileAccessLogRecord(obj interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ReconcileAccessLogRecord", reflect.TypeOf((*MockAccessLogRecordFinalizer)(nil).ReconcileAccessLogRecord), obj)
}

// MockAccessLogRecordReconcileLoop is a mock of AccessLogRecordReconcileLoop interface.
type MockAccessLogRecordReconcileLoop struct {
	ctrl     *gomock.Controller
	recorder *MockAccessLogRecordReconcileLoopMockRecorder
}

// MockAccessLogRecordReconcileLoopMockRecorder is the mock recorder for MockAccessLogRecordReconcileLoop.
type MockAccessLogRecordReconcileLoopMockRecorder struct {
	mock *MockAccessLogRecordReconcileLoop
}

// NewMockAccessLogRecordReconcileLoop creates a new mock instance.
func NewMockAccessLogRecordReconcileLoop(ctrl *gomock.Controller) *MockAccessLogRecordReconcileLoop {
	mock := &MockAccessLogRecordReconcileLoop{ctrl: ctrl}
	mock.recorder = &MockAccessLogRecordReconcileLoopMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockAccessLogRecordReconcileLoop) EXPECT() *MockAccessLogRecordReconcileLoopMockRecorder {
	return m.recorder
}

// RunAccessLogRecordReconciler mocks base method.
func (m *MockAccessLogRecordReconcileLoop) RunAccessLogRecordReconciler(ctx context.Context, rec controller.AccessLogRecordReconciler, predicates ...predicate.Predicate) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, rec}
	for _, a := range predicates {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "RunAccessLogRecordReconciler", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// RunAccessLogRecordReconciler indicates an expected call of RunAccessLogRecordReconciler.
func (mr *MockAccessLogRecordReconcileLoopMockRecorder) RunAccessLogRecordReconciler(ctx, rec interface{}, predicates ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, rec}, predicates...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RunAccessLogRecordReconciler", reflect.TypeOf((*MockAccessLogRecordReconcileLoop)(nil).RunAccessLogRecordReconciler), varargs...)
}
