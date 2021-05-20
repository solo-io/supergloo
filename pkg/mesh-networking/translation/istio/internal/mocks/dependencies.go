// Code generated by MockGen. DO NOT EDIT.
// Source: ./dependencies.go

// Package mock_internal is a generated GoMock package.
package mock_internal

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	v1sets "github.com/solo-io/external-apis/pkg/api/k8s/core/v1/sets"
	v1sets0 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1/sets"
	input "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/input"
	destination "github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/destination"
	mesh "github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/mesh"
	v1alpha1sets "github.com/solo-io/skv2/pkg/api/multicluster.solo.io/v1alpha1/sets"
)

// MockDependencyFactory is a mock of DependencyFactory interface.
type MockDependencyFactory struct {
	ctrl     *gomock.Controller
	recorder *MockDependencyFactoryMockRecorder
}

// MockDependencyFactoryMockRecorder is the mock recorder for MockDependencyFactory.
type MockDependencyFactoryMockRecorder struct {
	mock *MockDependencyFactory
}

// NewMockDependencyFactory creates a new mock instance.
func NewMockDependencyFactory(ctrl *gomock.Controller) *MockDependencyFactory {
	mock := &MockDependencyFactory{ctrl: ctrl}
	mock.recorder = &MockDependencyFactoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockDependencyFactory) EXPECT() *MockDependencyFactoryMockRecorder {
	return m.recorder
}

// MakeDestinationTranslator mocks base method.
func (m *MockDependencyFactory) MakeDestinationTranslator(ctx context.Context, userSupplied input.RemoteSnapshot, clusters v1alpha1sets.KubernetesClusterSet, destinations v1sets0.DestinationSet) destination.Translator {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MakeDestinationTranslator", ctx, userSupplied, clusters, destinations)
	ret0, _ := ret[0].(destination.Translator)
	return ret0
}

// MakeDestinationTranslator indicates an expected call of MakeDestinationTranslator.
func (mr *MockDependencyFactoryMockRecorder) MakeDestinationTranslator(ctx, userSupplied, clusters, destinations interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MakeDestinationTranslator", reflect.TypeOf((*MockDependencyFactory)(nil).MakeDestinationTranslator), ctx, userSupplied, clusters, destinations)
}

// MakeMeshTranslator mocks base method.
func (m *MockDependencyFactory) MakeMeshTranslator(ctx context.Context, secrets v1sets.SecretSet, workloads v1sets0.WorkloadSet) mesh.Translator {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MakeMeshTranslator", ctx, secrets, workloads)
	ret0, _ := ret[0].(mesh.Translator)
	return ret0
}

// MakeMeshTranslator indicates an expected call of MakeMeshTranslator.
func (mr *MockDependencyFactoryMockRecorder) MakeMeshTranslator(ctx, secrets, workloads interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MakeMeshTranslator", reflect.TypeOf((*MockDependencyFactory)(nil).MakeMeshTranslator), ctx, secrets, workloads)
}
