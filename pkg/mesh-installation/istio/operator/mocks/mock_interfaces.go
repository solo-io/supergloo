// Code generated by MockGen. DO NOT EDIT.
// Source: ./interfaces.go

// Package mock_operator is a generated GoMock package.
package mock_operator

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	operator "github.com/solo-io/service-mesh-hub/pkg/mesh-installation/istio/operator"
	v1 "k8s.io/api/apps/v1"
)

// MockOperatorManager is a mock of OperatorManager interface
type MockOperatorManager struct {
	ctrl     *gomock.Controller
	recorder *MockOperatorManagerMockRecorder
}

// MockOperatorManagerMockRecorder is the mock recorder for MockOperatorManager
type MockOperatorManagerMockRecorder struct {
	mock *MockOperatorManager
}

// NewMockOperatorManager creates a new mock instance
func NewMockOperatorManager(ctrl *gomock.Controller) *MockOperatorManager {
	mock := &MockOperatorManager{ctrl: ctrl}
	mock.recorder = &MockOperatorManagerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockOperatorManager) EXPECT() *MockOperatorManagerMockRecorder {
	return m.recorder
}

// InstallOperatorApplication mocks base method
func (m *MockOperatorManager) InstallOperatorApplication(installationOptions *operator.InstallationOptions) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "InstallOperatorApplication", installationOptions)
	ret0, _ := ret[0].(error)
	return ret0
}

// InstallOperatorApplication indicates an expected call of InstallOperatorApplication
func (mr *MockOperatorManagerMockRecorder) InstallOperatorApplication(installationOptions interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InstallOperatorApplication", reflect.TypeOf((*MockOperatorManager)(nil).InstallOperatorApplication), installationOptions)
}

// InstallDryRun mocks base method
func (m *MockOperatorManager) InstallDryRun(installationOptions *operator.InstallationOptions) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "InstallDryRun", installationOptions)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// InstallDryRun indicates an expected call of InstallDryRun
func (mr *MockOperatorManagerMockRecorder) InstallDryRun(installationOptions interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InstallDryRun", reflect.TypeOf((*MockOperatorManager)(nil).InstallDryRun), installationOptions)
}

// OperatorConfigDryRun mocks base method
func (m *MockOperatorManager) OperatorConfigDryRun(installationOptions *operator.InstallationOptions) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "OperatorConfigDryRun", installationOptions)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// OperatorConfigDryRun indicates an expected call of OperatorConfigDryRun
func (mr *MockOperatorManagerMockRecorder) OperatorConfigDryRun(installationOptions interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "OperatorConfigDryRun", reflect.TypeOf((*MockOperatorManager)(nil).OperatorConfigDryRun), installationOptions)
}

// MockInstallerManifestBuilder is a mock of InstallerManifestBuilder interface
type MockInstallerManifestBuilder struct {
	ctrl     *gomock.Controller
	recorder *MockInstallerManifestBuilderMockRecorder
}

// MockInstallerManifestBuilderMockRecorder is the mock recorder for MockInstallerManifestBuilder
type MockInstallerManifestBuilderMockRecorder struct {
	mock *MockInstallerManifestBuilder
}

// NewMockInstallerManifestBuilder creates a new mock instance
func NewMockInstallerManifestBuilder(ctrl *gomock.Controller) *MockInstallerManifestBuilder {
	mock := &MockInstallerManifestBuilder{ctrl: ctrl}
	mock.recorder = &MockInstallerManifestBuilderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockInstallerManifestBuilder) EXPECT() *MockInstallerManifestBuilderMockRecorder {
	return m.recorder
}

// BuildOperatorDeploymentManifest mocks base method
func (m *MockInstallerManifestBuilder) BuildOperatorDeploymentManifest(istioVersion operator.IstioVersion, installNamespace string, createNamespace bool) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BuildOperatorDeploymentManifest", istioVersion, installNamespace, createNamespace)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// BuildOperatorDeploymentManifest indicates an expected call of BuildOperatorDeploymentManifest
func (mr *MockInstallerManifestBuilderMockRecorder) BuildOperatorDeploymentManifest(istioVersion, installNamespace, createNamespace interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BuildOperatorDeploymentManifest", reflect.TypeOf((*MockInstallerManifestBuilder)(nil).BuildOperatorDeploymentManifest), istioVersion, installNamespace, createNamespace)
}

// BuildOperatorConfigurationWithProfile mocks base method
func (m *MockInstallerManifestBuilder) BuildOperatorConfigurationWithProfile(profile, installationNamespace string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BuildOperatorConfigurationWithProfile", profile, installationNamespace)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// BuildOperatorConfigurationWithProfile indicates an expected call of BuildOperatorConfigurationWithProfile
func (mr *MockInstallerManifestBuilderMockRecorder) BuildOperatorConfigurationWithProfile(profile, installationNamespace interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BuildOperatorConfigurationWithProfile", reflect.TypeOf((*MockInstallerManifestBuilder)(nil).BuildOperatorConfigurationWithProfile), profile, installationNamespace)
}

// MockOperatorDao is a mock of OperatorDao interface
type MockOperatorDao struct {
	ctrl     *gomock.Controller
	recorder *MockOperatorDaoMockRecorder
}

// MockOperatorDaoMockRecorder is the mock recorder for MockOperatorDao
type MockOperatorDaoMockRecorder struct {
	mock *MockOperatorDao
}

// NewMockOperatorDao creates a new mock instance
func NewMockOperatorDao(ctrl *gomock.Controller) *MockOperatorDao {
	mock := &MockOperatorDao{ctrl: ctrl}
	mock.recorder = &MockOperatorDaoMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockOperatorDao) EXPECT() *MockOperatorDaoMockRecorder {
	return m.recorder
}

// ApplyManifest mocks base method
func (m *MockOperatorDao) ApplyManifest(installationNamespace, manifest string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ApplyManifest", installationNamespace, manifest)
	ret0, _ := ret[0].(error)
	return ret0
}

// ApplyManifest indicates an expected call of ApplyManifest
func (mr *MockOperatorDaoMockRecorder) ApplyManifest(installationNamespace, manifest interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ApplyManifest", reflect.TypeOf((*MockOperatorDao)(nil).ApplyManifest), installationNamespace, manifest)
}

// FindOperatorDeployment mocks base method
func (m *MockOperatorDao) FindOperatorDeployment(name, namespace string) (*v1.Deployment, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FindOperatorDeployment", name, namespace)
	ret0, _ := ret[0].(*v1.Deployment)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FindOperatorDeployment indicates an expected call of FindOperatorDeployment
func (mr *MockOperatorDaoMockRecorder) FindOperatorDeployment(name, namespace interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FindOperatorDeployment", reflect.TypeOf((*MockOperatorDao)(nil).FindOperatorDeployment), name, namespace)
}
