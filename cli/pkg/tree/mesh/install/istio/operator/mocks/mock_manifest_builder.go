// Code generated by MockGen. DO NOT EDIT.
// Source: ./manifest.go

// Package mock_operator is a generated GoMock package.
package mock_operator

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	options "github.com/solo-io/service-mesh-hub/cli/pkg/options"
	operator "github.com/solo-io/service-mesh-hub/cli/pkg/tree/mesh/install/istio/operator"
)

// MockInstallerManifestBuilder is a mock of InstallerManifestBuilder interface.
type MockInstallerManifestBuilder struct {
	ctrl     *gomock.Controller
	recorder *MockInstallerManifestBuilderMockRecorder
}

// MockInstallerManifestBuilderMockRecorder is the mock recorder for MockInstallerManifestBuilder.
type MockInstallerManifestBuilderMockRecorder struct {
	mock *MockInstallerManifestBuilder
}

// NewMockInstallerManifestBuilder creates a new mock instance.
func NewMockInstallerManifestBuilder(ctrl *gomock.Controller) *MockInstallerManifestBuilder {
	mock := &MockInstallerManifestBuilder{ctrl: ctrl}
	mock.recorder = &MockInstallerManifestBuilderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockInstallerManifestBuilder) EXPECT() *MockInstallerManifestBuilderMockRecorder {
	return m.recorder
}

// Build mocks base method.
func (m *MockInstallerManifestBuilder) Build(istioVersion operator.IstioVersion, options *options.MeshInstallationConfig) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Build", istioVersion, options)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Build indicates an expected call of Build.
func (mr *MockInstallerManifestBuilderMockRecorder) Build(istioVersion, options interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Build", reflect.TypeOf((*MockInstallerManifestBuilder)(nil).Build), istioVersion, options)
}

// GetOperatorSpecWithProfile mocks base method.
func (m *MockInstallerManifestBuilder) GetOperatorSpecWithProfile(profile, installationNamespace string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetOperatorSpecWithProfile", profile, installationNamespace)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetOperatorSpecWithProfile indicates an expected call of GetOperatorSpecWithProfile.
func (mr *MockInstallerManifestBuilderMockRecorder) GetOperatorSpecWithProfile(profile, installationNamespace interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetOperatorSpecWithProfile", reflect.TypeOf((*MockInstallerManifestBuilder)(nil).GetOperatorSpecWithProfile), profile, installationNamespace)
}
