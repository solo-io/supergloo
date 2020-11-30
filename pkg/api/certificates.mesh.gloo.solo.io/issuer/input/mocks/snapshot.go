// Code generated by MockGen. DO NOT EDIT.
// Source: ./snapshot.go

// Package mock_input is a generated GoMock package.
package mock_input

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	input "github.com/solo-io/gloo-mesh/pkg/api/certificates.mesh.gloo.solo.io/issuer/input/issuer"
	v1alpha2sets "github.com/solo-io/gloo-mesh/pkg/api/certificates.mesh.gloo.solo.io/v1alpha2/sets"
	multicluster "github.com/solo-io/skv2/pkg/multicluster"
	client "sigs.k8s.io/controller-runtime/pkg/client"
)

// MockSnapshot is a mock of Snapshot interface
type MockSnapshot struct {
	ctrl     *gomock.Controller
	recorder *MockSnapshotMockRecorder
}

// MockSnapshotMockRecorder is the mock recorder for MockSnapshot
type MockSnapshotMockRecorder struct {
	mock *MockSnapshot
}

// NewMockSnapshot creates a new mock instance
func NewMockSnapshot(ctrl *gomock.Controller) *MockSnapshot {
	mock := &MockSnapshot{ctrl: ctrl}
	mock.recorder = &MockSnapshotMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockSnapshot) EXPECT() *MockSnapshotMockRecorder {
	return m.recorder
}

// IssuedCertificates mocks base method
func (m *MockSnapshot) IssuedCertificates() v1alpha2sets.IssuedCertificateSet {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IssuedCertificates")
	ret0, _ := ret[0].(v1alpha2sets.IssuedCertificateSet)
	return ret0
}

// IssuedCertificates indicates an expected call of IssuedCertificates
func (mr *MockSnapshotMockRecorder) IssuedCertificates() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IssuedCertificates", reflect.TypeOf((*MockSnapshot)(nil).IssuedCertificates))
}

// CertificateRequests mocks base method
func (m *MockSnapshot) CertificateRequests() v1alpha2sets.CertificateRequestSet {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CertificateRequests")
	ret0, _ := ret[0].(v1alpha2sets.CertificateRequestSet)
	return ret0
}

// CertificateRequests indicates an expected call of CertificateRequests
func (mr *MockSnapshotMockRecorder) CertificateRequests() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CertificateRequests", reflect.TypeOf((*MockSnapshot)(nil).CertificateRequests))
}

// SyncStatuses mocks base method
func (m *MockSnapshot) SyncStatuses(ctx context.Context, c client.Client) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SyncStatuses", ctx, c)
	ret0, _ := ret[0].(error)
	return ret0
}

// SyncStatuses indicates an expected call of SyncStatuses
func (mr *MockSnapshotMockRecorder) SyncStatuses(ctx, c interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SyncStatuses", reflect.TypeOf((*MockSnapshot)(nil).SyncStatuses), ctx, c)
}

// SyncStatusesMultiCluster mocks base method
func (m *MockSnapshot) SyncStatusesMultiCluster(ctx context.Context, mcClient multicluster.Client) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SyncStatusesMultiCluster", ctx, mcClient)
	ret0, _ := ret[0].(error)
	return ret0
}

// SyncStatusesMultiCluster indicates an expected call of SyncStatusesMultiCluster
func (mr *MockSnapshotMockRecorder) SyncStatusesMultiCluster(ctx, mcClient interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SyncStatusesMultiCluster", reflect.TypeOf((*MockSnapshot)(nil).SyncStatusesMultiCluster), ctx, mcClient)
}

// MarshalJSON mocks base method
func (m *MockSnapshot) MarshalJSON() ([]byte, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MarshalJSON")
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// MarshalJSON indicates an expected call of MarshalJSON
func (mr *MockSnapshotMockRecorder) MarshalJSON() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MarshalJSON", reflect.TypeOf((*MockSnapshot)(nil).MarshalJSON))
}

// MockBuilder is a mock of Builder interface
type MockBuilder struct {
	ctrl     *gomock.Controller
	recorder *MockBuilderMockRecorder
}

// MockBuilderMockRecorder is the mock recorder for MockBuilder
type MockBuilderMockRecorder struct {
	mock *MockBuilder
}

// NewMockBuilder creates a new mock instance
func NewMockBuilder(ctrl *gomock.Controller) *MockBuilder {
	mock := &MockBuilder{ctrl: ctrl}
	mock.recorder = &MockBuilderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockBuilder) EXPECT() *MockBuilderMockRecorder {
	return m.recorder
}

// BuildSnapshot mocks base method
func (m *MockBuilder) BuildSnapshot(ctx context.Context, name string, opts input.BuildOptions) (input.Snapshot, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BuildSnapshot", ctx, name, opts)
	ret0, _ := ret[0].(input.Snapshot)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// BuildSnapshot indicates an expected call of BuildSnapshot
func (mr *MockBuilderMockRecorder) BuildSnapshot(ctx, name, opts interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BuildSnapshot", reflect.TypeOf((*MockBuilder)(nil).BuildSnapshot), ctx, name, opts)
}
