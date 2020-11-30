// Code generated by MockGen. DO NOT EDIT.
// Source: ./reconciler.go

// Package mock_user is a generated GoMock package.
package mock_user

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	reconcile "github.com/solo-io/skv2/pkg/reconcile"
	v1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
)

// MockmultiClusterReconciler is a mock of multiClusterReconciler interface
type MockmultiClusterReconciler struct {
	ctrl     *gomock.Controller
	recorder *MockmultiClusterReconcilerMockRecorder
}

// MockmultiClusterReconcilerMockRecorder is the mock recorder for MockmultiClusterReconciler
type MockmultiClusterReconcilerMockRecorder struct {
	mock *MockmultiClusterReconciler
}

// NewMockmultiClusterReconciler creates a new mock instance
func NewMockmultiClusterReconciler(ctrl *gomock.Controller) *MockmultiClusterReconciler {
	mock := &MockmultiClusterReconciler{ctrl: ctrl}
	mock.recorder = &MockmultiClusterReconcilerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockmultiClusterReconciler) EXPECT() *MockmultiClusterReconcilerMockRecorder {
	return m.recorder
}

// ReconcileDestinationRule mocks base method
func (m *MockmultiClusterReconciler) ReconcileDestinationRule(clusterName string, obj *v1alpha3.DestinationRule) (reconcile.Result, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ReconcileDestinationRule", clusterName, obj)
	ret0, _ := ret[0].(reconcile.Result)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ReconcileDestinationRule indicates an expected call of ReconcileDestinationRule
func (mr *MockmultiClusterReconcilerMockRecorder) ReconcileDestinationRule(clusterName, obj interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ReconcileDestinationRule", reflect.TypeOf((*MockmultiClusterReconciler)(nil).ReconcileDestinationRule), clusterName, obj)
}

// ReconcileVirtualService mocks base method
func (m *MockmultiClusterReconciler) ReconcileVirtualService(clusterName string, obj *v1alpha3.VirtualService) (reconcile.Result, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ReconcileVirtualService", clusterName, obj)
	ret0, _ := ret[0].(reconcile.Result)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ReconcileVirtualService indicates an expected call of ReconcileVirtualService
func (mr *MockmultiClusterReconcilerMockRecorder) ReconcileVirtualService(clusterName, obj interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ReconcileVirtualService", reflect.TypeOf((*MockmultiClusterReconciler)(nil).ReconcileVirtualService), clusterName, obj)
}

// MocksingleClusterReconciler is a mock of singleClusterReconciler interface
type MocksingleClusterReconciler struct {
	ctrl     *gomock.Controller
	recorder *MocksingleClusterReconcilerMockRecorder
}

// MocksingleClusterReconcilerMockRecorder is the mock recorder for MocksingleClusterReconciler
type MocksingleClusterReconcilerMockRecorder struct {
	mock *MocksingleClusterReconciler
}

// NewMocksingleClusterReconciler creates a new mock instance
func NewMocksingleClusterReconciler(ctrl *gomock.Controller) *MocksingleClusterReconciler {
	mock := &MocksingleClusterReconciler{ctrl: ctrl}
	mock.recorder = &MocksingleClusterReconcilerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MocksingleClusterReconciler) EXPECT() *MocksingleClusterReconcilerMockRecorder {
	return m.recorder
}

// ReconcileDestinationRule mocks base method
func (m *MocksingleClusterReconciler) ReconcileDestinationRule(obj *v1alpha3.DestinationRule) (reconcile.Result, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ReconcileDestinationRule", obj)
	ret0, _ := ret[0].(reconcile.Result)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ReconcileDestinationRule indicates an expected call of ReconcileDestinationRule
func (mr *MocksingleClusterReconcilerMockRecorder) ReconcileDestinationRule(obj interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ReconcileDestinationRule", reflect.TypeOf((*MocksingleClusterReconciler)(nil).ReconcileDestinationRule), obj)
}

// ReconcileVirtualService mocks base method
func (m *MocksingleClusterReconciler) ReconcileVirtualService(obj *v1alpha3.VirtualService) (reconcile.Result, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ReconcileVirtualService", obj)
	ret0, _ := ret[0].(reconcile.Result)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ReconcileVirtualService indicates an expected call of ReconcileVirtualService
func (mr *MocksingleClusterReconcilerMockRecorder) ReconcileVirtualService(obj interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ReconcileVirtualService", reflect.TypeOf((*MocksingleClusterReconciler)(nil).ReconcileVirtualService), obj)
}
