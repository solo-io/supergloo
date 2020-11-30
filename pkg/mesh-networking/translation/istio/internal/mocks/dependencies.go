// Code generated by MockGen. DO NOT EDIT.
// Source: ./dependencies.go

// Package mock_internal is a generated GoMock package.
package mock_internal

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	v1sets "github.com/solo-io/external-apis/pkg/api/k8s/core/v1/sets"
	v1alpha2sets "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1alpha2/sets"
	user "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/input/user"
	v1alpha2sets0 "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1alpha2/sets"
	mesh "github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/mesh"
	traffictarget "github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/traffictarget"
	v1alpha1sets "github.com/solo-io/skv2/pkg/api/multicluster.solo.io/v1alpha1/sets"
)

// MockDependencyFactory is a mock of DependencyFactory interface
type MockDependencyFactory struct {
	ctrl     *gomock.Controller
	recorder *MockDependencyFactoryMockRecorder
}

// MockDependencyFactoryMockRecorder is the mock recorder for MockDependencyFactory
type MockDependencyFactoryMockRecorder struct {
	mock *MockDependencyFactory
}

// NewMockDependencyFactory creates a new mock instance
func NewMockDependencyFactory(ctrl *gomock.Controller) *MockDependencyFactory {
	mock := &MockDependencyFactory{ctrl: ctrl}
	mock.recorder = &MockDependencyFactoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockDependencyFactory) EXPECT() *MockDependencyFactoryMockRecorder {
	return m.recorder
}

// MakeTrafficTargetTranslator mocks base method
func (m *MockDependencyFactory) MakeTrafficTargetTranslator(ctx context.Context, userInputSnap user.Snapshot, clusters v1alpha1sets.KubernetesClusterSet, trafficTargets v1alpha2sets.TrafficTargetSet, failoverServices v1alpha2sets0.FailoverServiceSet) traffictarget.Translator {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MakeTrafficTargetTranslator", ctx, userInputSnap, clusters, trafficTargets, failoverServices)
	ret0, _ := ret[0].(traffictarget.Translator)
	return ret0
}

// MakeTrafficTargetTranslator indicates an expected call of MakeTrafficTargetTranslator
func (mr *MockDependencyFactoryMockRecorder) MakeTrafficTargetTranslator(ctx, userInputSnap, clusters, trafficTargets, failoverServices interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MakeTrafficTargetTranslator", reflect.TypeOf((*MockDependencyFactory)(nil).MakeTrafficTargetTranslator), ctx, userInputSnap, clusters, trafficTargets, failoverServices)
}

// MakeMeshTranslator mocks base method
func (m *MockDependencyFactory) MakeMeshTranslator(ctx context.Context, userInputSnap user.Snapshot, clusters v1alpha1sets.KubernetesClusterSet, secrets v1sets.SecretSet, workloads v1alpha2sets.WorkloadSet, trafficTargets v1alpha2sets.TrafficTargetSet, failoverServices v1alpha2sets0.FailoverServiceSet) mesh.Translator {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MakeMeshTranslator", ctx, userInputSnap, clusters, secrets, workloads, trafficTargets, failoverServices)
	ret0, _ := ret[0].(mesh.Translator)
	return ret0
}

// MakeMeshTranslator indicates an expected call of MakeMeshTranslator
func (mr *MockDependencyFactoryMockRecorder) MakeMeshTranslator(ctx, userInputSnap, clusters, secrets, workloads, trafficTargets, failoverServices interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MakeMeshTranslator", reflect.TypeOf((*MockDependencyFactory)(nil).MakeMeshTranslator), ctx, userInputSnap, clusters, secrets, workloads, trafficTargets, failoverServices)
}
