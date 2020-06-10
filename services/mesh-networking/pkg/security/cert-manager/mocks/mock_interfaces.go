// Code generated by MockGen. DO NOT EDIT.
// Source: ./interfaces.go

// Package mock_cert_manager is a generated GoMock package.
package mock_cert_manager

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	v1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	v1alpha10 "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	types "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1/types"
	types0 "github.com/solo-io/service-mesh-hub/pkg/api/security.smh.solo.io/v1alpha1/types"
)

// MockCertConfigProducer is a mock of CertConfigProducer interface
type MockCertConfigProducer struct {
	ctrl     *gomock.Controller
	recorder *MockCertConfigProducerMockRecorder
}

// MockCertConfigProducerMockRecorder is the mock recorder for MockCertConfigProducer
type MockCertConfigProducerMockRecorder struct {
	mock *MockCertConfigProducer
}

// NewMockCertConfigProducer creates a new mock instance
func NewMockCertConfigProducer(ctrl *gomock.Controller) *MockCertConfigProducer {
	mock := &MockCertConfigProducer{ctrl: ctrl}
	mock.recorder = &MockCertConfigProducerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockCertConfigProducer) EXPECT() *MockCertConfigProducerMockRecorder {
	return m.recorder
}

// ConfigureCertificateInfo mocks base method
func (m *MockCertConfigProducer) ConfigureCertificateInfo(vm *v1alpha10.VirtualMesh, mesh *v1alpha1.Mesh) (*types0.VirtualMeshCertificateSigningRequestSpec_CertConfig, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ConfigureCertificateInfo", vm, mesh)
	ret0, _ := ret[0].(*types0.VirtualMeshCertificateSigningRequestSpec_CertConfig)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ConfigureCertificateInfo indicates an expected call of ConfigureCertificateInfo
func (mr *MockCertConfigProducerMockRecorder) ConfigureCertificateInfo(vm, mesh interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ConfigureCertificateInfo", reflect.TypeOf((*MockCertConfigProducer)(nil).ConfigureCertificateInfo), vm, mesh)
}

// MockVirtualMeshCertificateManager is a mock of VirtualMeshCertificateManager interface
type MockVirtualMeshCertificateManager struct {
	ctrl     *gomock.Controller
	recorder *MockVirtualMeshCertificateManagerMockRecorder
}

// MockVirtualMeshCertificateManagerMockRecorder is the mock recorder for MockVirtualMeshCertificateManager
type MockVirtualMeshCertificateManagerMockRecorder struct {
	mock *MockVirtualMeshCertificateManager
}

// NewMockVirtualMeshCertificateManager creates a new mock instance
func NewMockVirtualMeshCertificateManager(ctrl *gomock.Controller) *MockVirtualMeshCertificateManager {
	mock := &MockVirtualMeshCertificateManager{ctrl: ctrl}
	mock.recorder = &MockVirtualMeshCertificateManagerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockVirtualMeshCertificateManager) EXPECT() *MockVirtualMeshCertificateManagerMockRecorder {
	return m.recorder
}

// InitializeCertificateForVirtualMesh mocks base method
func (m *MockVirtualMeshCertificateManager) InitializeCertificateForVirtualMesh(ctx context.Context, new *v1alpha10.VirtualMesh) types.VirtualMeshStatus {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "InitializeCertificateForVirtualMesh", ctx, new)
	ret0, _ := ret[0].(types.VirtualMeshStatus)
	return ret0
}

// InitializeCertificateForVirtualMesh indicates an expected call of InitializeCertificateForVirtualMesh
func (mr *MockVirtualMeshCertificateManagerMockRecorder) InitializeCertificateForVirtualMesh(ctx, new interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InitializeCertificateForVirtualMesh", reflect.TypeOf((*MockVirtualMeshCertificateManager)(nil).InitializeCertificateForVirtualMesh), ctx, new)
}
