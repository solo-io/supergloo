// Code generated by MockGen. DO NOT EDIT.
// Source: interfaces.go

// Package mock_aws is a generated GoMock package.
package mock_aws

import (
	context "context"
	reflect "reflect"

	credentials "github.com/aws/aws-sdk-go/aws/credentials"
	gomock "github.com/golang/mock/gomock"
)

// MockRestAPIDiscoveryReconciler is a mock of RestAPIDiscoveryReconciler interface.
type MockRestAPIDiscoveryReconciler struct {
	ctrl     *gomock.Controller
	recorder *MockRestAPIDiscoveryReconcilerMockRecorder
}

// MockRestAPIDiscoveryReconcilerMockRecorder is the mock recorder for MockRestAPIDiscoveryReconciler.
type MockRestAPIDiscoveryReconcilerMockRecorder struct {
	mock *MockRestAPIDiscoveryReconciler
}

// NewMockRestAPIDiscoveryReconciler creates a new mock instance.
func NewMockRestAPIDiscoveryReconciler(ctrl *gomock.Controller) *MockRestAPIDiscoveryReconciler {
	mock := &MockRestAPIDiscoveryReconciler{ctrl: ctrl}
	mock.recorder = &MockRestAPIDiscoveryReconcilerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockRestAPIDiscoveryReconciler) EXPECT() *MockRestAPIDiscoveryReconcilerMockRecorder {
	return m.recorder
}

// Reconcile mocks base method.
func (m *MockRestAPIDiscoveryReconciler) Reconcile(ctx context.Context, creds *credentials.Credentials, region string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Reconcile", ctx, creds, region)
	ret0, _ := ret[0].(error)
	return ret0
}

// Reconcile indicates an expected call of Reconcile.
func (mr *MockRestAPIDiscoveryReconcilerMockRecorder) Reconcile(ctx, creds, region interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Reconcile", reflect.TypeOf((*MockRestAPIDiscoveryReconciler)(nil).Reconcile), ctx, creds, region)
}

// GetName mocks base method.
func (m *MockRestAPIDiscoveryReconciler) GetName() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetName")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetName indicates an expected call of GetName.
func (mr *MockRestAPIDiscoveryReconcilerMockRecorder) GetName() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetName", reflect.TypeOf((*MockRestAPIDiscoveryReconciler)(nil).GetName))
}
