// Code generated by MockGen. DO NOT EDIT.
// Source: ./local_snapshot.go

// Package mock_input is a generated GoMock package.
package mock_input

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	v1sets "github.com/solo-io/external-apis/pkg/api/k8s/core/v1/sets"
	v1alpha2sets "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1alpha2/sets"
	v1alpha1sets "github.com/solo-io/gloo-mesh/pkg/api/networking.enterprise.mesh.gloo.solo.io/v1alpha1/sets"
	input "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/input"
	v1alpha2sets0 "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1alpha2/sets"
	v1alpha1sets0 "github.com/solo-io/gloo-mesh/pkg/api/observability.enterprise.mesh.gloo.solo.io/v1alpha1/sets"
	v1alpha2sets1 "github.com/solo-io/gloo-mesh/pkg/api/settings.mesh.gloo.solo.io/v1alpha2/sets"
	v1alpha1sets1 "github.com/solo-io/skv2/pkg/api/multicluster.solo.io/v1alpha1/sets"
	multicluster "github.com/solo-io/skv2/pkg/multicluster"
	client "sigs.k8s.io/controller-runtime/pkg/client"
)

// MockLocalSnapshot is a mock of LocalSnapshot interface
type MockLocalSnapshot struct {
	ctrl     *gomock.Controller
	recorder *MockLocalSnapshotMockRecorder
}

// MockLocalSnapshotMockRecorder is the mock recorder for MockLocalSnapshot
type MockLocalSnapshotMockRecorder struct {
	mock *MockLocalSnapshot
}

// NewMockLocalSnapshot creates a new mock instance
func NewMockLocalSnapshot(ctrl *gomock.Controller) *MockLocalSnapshot {
	mock := &MockLocalSnapshot{ctrl: ctrl}
	mock.recorder = &MockLocalSnapshotMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockLocalSnapshot) EXPECT() *MockLocalSnapshotMockRecorder {
	return m.recorder
}

// SettingsMeshGlooSoloIov1Alpha2Settings mocks base method
func (m *MockLocalSnapshot) SettingsMeshGlooSoloIov1Alpha2Settings() v1alpha2sets1.SettingsSet {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SettingsMeshGlooSoloIov1Alpha2Settings")
	ret0, _ := ret[0].(v1alpha2sets1.SettingsSet)
	return ret0
}

// SettingsMeshGlooSoloIov1Alpha2Settings indicates an expected call of SettingsMeshGlooSoloIov1Alpha2Settings
func (mr *MockLocalSnapshotMockRecorder) SettingsMeshGlooSoloIov1Alpha2Settings() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SettingsMeshGlooSoloIov1Alpha2Settings", reflect.TypeOf((*MockLocalSnapshot)(nil).SettingsMeshGlooSoloIov1Alpha2Settings))
}

// DiscoveryMeshGlooSoloIov1Alpha2TrafficTargets mocks base method
func (m *MockLocalSnapshot) DiscoveryMeshGlooSoloIov1Alpha2TrafficTargets() v1alpha2sets.TrafficTargetSet {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DiscoveryMeshGlooSoloIov1Alpha2TrafficTargets")
	ret0, _ := ret[0].(v1alpha2sets.TrafficTargetSet)
	return ret0
}

// DiscoveryMeshGlooSoloIov1Alpha2TrafficTargets indicates an expected call of DiscoveryMeshGlooSoloIov1Alpha2TrafficTargets
func (mr *MockLocalSnapshotMockRecorder) DiscoveryMeshGlooSoloIov1Alpha2TrafficTargets() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DiscoveryMeshGlooSoloIov1Alpha2TrafficTargets", reflect.TypeOf((*MockLocalSnapshot)(nil).DiscoveryMeshGlooSoloIov1Alpha2TrafficTargets))
}

// DiscoveryMeshGlooSoloIov1Alpha2Workloads mocks base method
func (m *MockLocalSnapshot) DiscoveryMeshGlooSoloIov1Alpha2Workloads() v1alpha2sets.WorkloadSet {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DiscoveryMeshGlooSoloIov1Alpha2Workloads")
	ret0, _ := ret[0].(v1alpha2sets.WorkloadSet)
	return ret0
}

// DiscoveryMeshGlooSoloIov1Alpha2Workloads indicates an expected call of DiscoveryMeshGlooSoloIov1Alpha2Workloads
func (mr *MockLocalSnapshotMockRecorder) DiscoveryMeshGlooSoloIov1Alpha2Workloads() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DiscoveryMeshGlooSoloIov1Alpha2Workloads", reflect.TypeOf((*MockLocalSnapshot)(nil).DiscoveryMeshGlooSoloIov1Alpha2Workloads))
}

// DiscoveryMeshGlooSoloIov1Alpha2Meshes mocks base method
func (m *MockLocalSnapshot) DiscoveryMeshGlooSoloIov1Alpha2Meshes() v1alpha2sets.MeshSet {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DiscoveryMeshGlooSoloIov1Alpha2Meshes")
	ret0, _ := ret[0].(v1alpha2sets.MeshSet)
	return ret0
}

// DiscoveryMeshGlooSoloIov1Alpha2Meshes indicates an expected call of DiscoveryMeshGlooSoloIov1Alpha2Meshes
func (mr *MockLocalSnapshotMockRecorder) DiscoveryMeshGlooSoloIov1Alpha2Meshes() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DiscoveryMeshGlooSoloIov1Alpha2Meshes", reflect.TypeOf((*MockLocalSnapshot)(nil).DiscoveryMeshGlooSoloIov1Alpha2Meshes))
}

// NetworkingMeshGlooSoloIov1Alpha2TrafficPolicies mocks base method
func (m *MockLocalSnapshot) NetworkingMeshGlooSoloIov1Alpha2TrafficPolicies() v1alpha2sets0.TrafficPolicySet {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "NetworkingMeshGlooSoloIov1Alpha2TrafficPolicies")
	ret0, _ := ret[0].(v1alpha2sets0.TrafficPolicySet)
	return ret0
}

// NetworkingMeshGlooSoloIov1Alpha2TrafficPolicies indicates an expected call of NetworkingMeshGlooSoloIov1Alpha2TrafficPolicies
func (mr *MockLocalSnapshotMockRecorder) NetworkingMeshGlooSoloIov1Alpha2TrafficPolicies() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NetworkingMeshGlooSoloIov1Alpha2TrafficPolicies", reflect.TypeOf((*MockLocalSnapshot)(nil).NetworkingMeshGlooSoloIov1Alpha2TrafficPolicies))
}

// NetworkingMeshGlooSoloIov1Alpha2AccessPolicies mocks base method
func (m *MockLocalSnapshot) NetworkingMeshGlooSoloIov1Alpha2AccessPolicies() v1alpha2sets0.AccessPolicySet {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "NetworkingMeshGlooSoloIov1Alpha2AccessPolicies")
	ret0, _ := ret[0].(v1alpha2sets0.AccessPolicySet)
	return ret0
}

// NetworkingMeshGlooSoloIov1Alpha2AccessPolicies indicates an expected call of NetworkingMeshGlooSoloIov1Alpha2AccessPolicies
func (mr *MockLocalSnapshotMockRecorder) NetworkingMeshGlooSoloIov1Alpha2AccessPolicies() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NetworkingMeshGlooSoloIov1Alpha2AccessPolicies", reflect.TypeOf((*MockLocalSnapshot)(nil).NetworkingMeshGlooSoloIov1Alpha2AccessPolicies))
}

// NetworkingMeshGlooSoloIov1Alpha2VirtualMeshes mocks base method
func (m *MockLocalSnapshot) NetworkingMeshGlooSoloIov1Alpha2VirtualMeshes() v1alpha2sets0.VirtualMeshSet {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "NetworkingMeshGlooSoloIov1Alpha2VirtualMeshes")
	ret0, _ := ret[0].(v1alpha2sets0.VirtualMeshSet)
	return ret0
}

// NetworkingMeshGlooSoloIov1Alpha2VirtualMeshes indicates an expected call of NetworkingMeshGlooSoloIov1Alpha2VirtualMeshes
func (mr *MockLocalSnapshotMockRecorder) NetworkingMeshGlooSoloIov1Alpha2VirtualMeshes() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NetworkingMeshGlooSoloIov1Alpha2VirtualMeshes", reflect.TypeOf((*MockLocalSnapshot)(nil).NetworkingMeshGlooSoloIov1Alpha2VirtualMeshes))
}

// NetworkingMeshGlooSoloIov1Alpha2FailoverServices mocks base method
func (m *MockLocalSnapshot) NetworkingMeshGlooSoloIov1Alpha2FailoverServices() v1alpha2sets0.FailoverServiceSet {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "NetworkingMeshGlooSoloIov1Alpha2FailoverServices")
	ret0, _ := ret[0].(v1alpha2sets0.FailoverServiceSet)
	return ret0
}

// NetworkingMeshGlooSoloIov1Alpha2FailoverServices indicates an expected call of NetworkingMeshGlooSoloIov1Alpha2FailoverServices
func (mr *MockLocalSnapshotMockRecorder) NetworkingMeshGlooSoloIov1Alpha2FailoverServices() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NetworkingMeshGlooSoloIov1Alpha2FailoverServices", reflect.TypeOf((*MockLocalSnapshot)(nil).NetworkingMeshGlooSoloIov1Alpha2FailoverServices))
}

// NetworkingEnterpriseMeshGlooSoloIov1Alpha1WasmDeployments mocks base method
func (m *MockLocalSnapshot) NetworkingEnterpriseMeshGlooSoloIov1Alpha1WasmDeployments() v1alpha1sets.WasmDeploymentSet {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "NetworkingEnterpriseMeshGlooSoloIov1Alpha1WasmDeployments")
	ret0, _ := ret[0].(v1alpha1sets.WasmDeploymentSet)
	return ret0
}

// NetworkingEnterpriseMeshGlooSoloIov1Alpha1WasmDeployments indicates an expected call of NetworkingEnterpriseMeshGlooSoloIov1Alpha1WasmDeployments
func (mr *MockLocalSnapshotMockRecorder) NetworkingEnterpriseMeshGlooSoloIov1Alpha1WasmDeployments() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NetworkingEnterpriseMeshGlooSoloIov1Alpha1WasmDeployments", reflect.TypeOf((*MockLocalSnapshot)(nil).NetworkingEnterpriseMeshGlooSoloIov1Alpha1WasmDeployments))
}

// ObservabilityEnterpriseMeshGlooSoloIov1Alpha1AccessLogRecords mocks base method
func (m *MockLocalSnapshot) ObservabilityEnterpriseMeshGlooSoloIov1Alpha1AccessLogRecords() v1alpha1sets0.AccessLogRecordSet {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ObservabilityEnterpriseMeshGlooSoloIov1Alpha1AccessLogRecords")
	ret0, _ := ret[0].(v1alpha1sets0.AccessLogRecordSet)
	return ret0
}

// ObservabilityEnterpriseMeshGlooSoloIov1Alpha1AccessLogRecords indicates an expected call of ObservabilityEnterpriseMeshGlooSoloIov1Alpha1AccessLogRecords
func (mr *MockLocalSnapshotMockRecorder) ObservabilityEnterpriseMeshGlooSoloIov1Alpha1AccessLogRecords() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ObservabilityEnterpriseMeshGlooSoloIov1Alpha1AccessLogRecords", reflect.TypeOf((*MockLocalSnapshot)(nil).ObservabilityEnterpriseMeshGlooSoloIov1Alpha1AccessLogRecords))
}

// V1Secrets mocks base method
func (m *MockLocalSnapshot) V1Secrets() v1sets.SecretSet {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "V1Secrets")
	ret0, _ := ret[0].(v1sets.SecretSet)
	return ret0
}

// V1Secrets indicates an expected call of V1Secrets
func (mr *MockLocalSnapshotMockRecorder) V1Secrets() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "V1Secrets", reflect.TypeOf((*MockLocalSnapshot)(nil).V1Secrets))
}

// MulticlusterSoloIov1Alpha1KubernetesClusters mocks base method
func (m *MockLocalSnapshot) MulticlusterSoloIov1Alpha1KubernetesClusters() v1alpha1sets1.KubernetesClusterSet {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MulticlusterSoloIov1Alpha1KubernetesClusters")
	ret0, _ := ret[0].(v1alpha1sets1.KubernetesClusterSet)
	return ret0
}

// MulticlusterSoloIov1Alpha1KubernetesClusters indicates an expected call of MulticlusterSoloIov1Alpha1KubernetesClusters
func (mr *MockLocalSnapshotMockRecorder) MulticlusterSoloIov1Alpha1KubernetesClusters() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MulticlusterSoloIov1Alpha1KubernetesClusters", reflect.TypeOf((*MockLocalSnapshot)(nil).MulticlusterSoloIov1Alpha1KubernetesClusters))
}

// SyncStatusesMultiCluster mocks base method
func (m *MockLocalSnapshot) SyncStatusesMultiCluster(ctx context.Context, mcClient multicluster.Client, opts input.LocalSyncStatusOptions) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SyncStatusesMultiCluster", ctx, mcClient, opts)
	ret0, _ := ret[0].(error)
	return ret0
}

// SyncStatusesMultiCluster indicates an expected call of SyncStatusesMultiCluster
func (mr *MockLocalSnapshotMockRecorder) SyncStatusesMultiCluster(ctx, mcClient, opts interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SyncStatusesMultiCluster", reflect.TypeOf((*MockLocalSnapshot)(nil).SyncStatusesMultiCluster), ctx, mcClient, opts)
}

// SyncStatuses mocks base method
func (m *MockLocalSnapshot) SyncStatuses(ctx context.Context, c client.Client, opts input.LocalSyncStatusOptions) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SyncStatuses", ctx, c, opts)
	ret0, _ := ret[0].(error)
	return ret0
}

// SyncStatuses indicates an expected call of SyncStatuses
func (mr *MockLocalSnapshotMockRecorder) SyncStatuses(ctx, c, opts interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SyncStatuses", reflect.TypeOf((*MockLocalSnapshot)(nil).SyncStatuses), ctx, c, opts)
}

// MarshalJSON mocks base method
func (m *MockLocalSnapshot) MarshalJSON() ([]byte, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MarshalJSON")
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// MarshalJSON indicates an expected call of MarshalJSON
func (mr *MockLocalSnapshotMockRecorder) MarshalJSON() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MarshalJSON", reflect.TypeOf((*MockLocalSnapshot)(nil).MarshalJSON))
}

// MockLocalBuilder is a mock of LocalBuilder interface
type MockLocalBuilder struct {
	ctrl     *gomock.Controller
	recorder *MockLocalBuilderMockRecorder
}

// MockLocalBuilderMockRecorder is the mock recorder for MockLocalBuilder
type MockLocalBuilderMockRecorder struct {
	mock *MockLocalBuilder
}

// NewMockLocalBuilder creates a new mock instance
func NewMockLocalBuilder(ctrl *gomock.Controller) *MockLocalBuilder {
	mock := &MockLocalBuilder{ctrl: ctrl}
	mock.recorder = &MockLocalBuilderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockLocalBuilder) EXPECT() *MockLocalBuilderMockRecorder {
	return m.recorder
}

// BuildSnapshot mocks base method
func (m *MockLocalBuilder) BuildSnapshot(ctx context.Context, name string, opts input.LocalBuildOptions) (input.LocalSnapshot, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BuildSnapshot", ctx, name, opts)
	ret0, _ := ret[0].(input.LocalSnapshot)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// BuildSnapshot indicates an expected call of BuildSnapshot
func (mr *MockLocalBuilderMockRecorder) BuildSnapshot(ctx, name, opts interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BuildSnapshot", reflect.TypeOf((*MockLocalBuilder)(nil).BuildSnapshot), ctx, name, opts)
}
