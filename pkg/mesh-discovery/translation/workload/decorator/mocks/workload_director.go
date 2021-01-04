// Code generated by MockGen. DO NOT EDIT.
// Source: ./workload_decorator.go

// Package mock_decorator is a generated GoMock package.
package mock_decorator

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	v1sets "github.com/solo-io/external-apis/pkg/api/k8s/core/v1/sets"
	v1alpha2 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1alpha2"
	types "github.com/solo-io/gloo-mesh/pkg/mesh-discovery/translation/workload/types"
)

// MockWorkloadDecorator is a mock of WorkloadDecorator interface
type MockWorkloadDecorator struct {
	ctrl     *gomock.Controller
	recorder *MockWorkloadDecoratorMockRecorder
}

// MockWorkloadDecoratorMockRecorder is the mock recorder for MockWorkloadDecorator
type MockWorkloadDecoratorMockRecorder struct {
	mock *MockWorkloadDecorator
}

// NewMockWorkloadDecorator creates a new mock instance
func NewMockWorkloadDecorator(ctrl *gomock.Controller) *MockWorkloadDecorator {
	mock := &MockWorkloadDecorator{ctrl: ctrl}
	mock.recorder = &MockWorkloadDecoratorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockWorkloadDecorator) EXPECT() *MockWorkloadDecoratorMockRecorder {
	return m.recorder
}

// DecorateWorkload mocks base method
func (m *MockWorkloadDecorator) DecorateWorkload(discoveredWorkload *v1alpha2.Workload, kubeWorkload types.Workload, mesh *v1alpha2.Mesh, pods v1sets.PodSet) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "DecorateWorkload", discoveredWorkload, kubeWorkload, mesh, pods)
}

// DecorateWorkload indicates an expected call of DecorateWorkload
func (mr *MockWorkloadDecoratorMockRecorder) DecorateWorkload(discoveredWorkload, kubeWorkload, mesh, pods interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DecorateWorkload", reflect.TypeOf((*MockWorkloadDecorator)(nil).DecorateWorkload), discoveredWorkload, kubeWorkload, mesh, pods)
}