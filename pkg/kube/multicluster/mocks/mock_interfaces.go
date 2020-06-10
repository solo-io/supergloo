// Code generated by MockGen. DO NOT EDIT.
// Source: ./interfaces.go

// Package mock_multicluster is a generated GoMock package.
package mock_multicluster

import (
	context "context"
	reflect "reflect"

	retry "github.com/avast/retry-go"
	gomock "github.com/golang/mock/gomock"
	client "sigs.k8s.io/controller-runtime/pkg/client"
)

// MockDynamicClientGetter is a mock of DynamicClientGetter interface
type MockDynamicClientGetter struct {
	ctrl     *gomock.Controller
	recorder *MockDynamicClientGetterMockRecorder
}

// MockDynamicClientGetterMockRecorder is the mock recorder for MockDynamicClientGetter
type MockDynamicClientGetterMockRecorder struct {
	mock *MockDynamicClientGetter
}

// NewMockDynamicClientGetter creates a new mock instance
func NewMockDynamicClientGetter(ctrl *gomock.Controller) *MockDynamicClientGetter {
	mock := &MockDynamicClientGetter{ctrl: ctrl}
	mock.recorder = &MockDynamicClientGetterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockDynamicClientGetter) EXPECT() *MockDynamicClientGetterMockRecorder {
	return m.recorder
}

// GetClientForCluster mocks base method
func (m *MockDynamicClientGetter) GetClientForCluster(ctx context.Context, clusterName string, opts ...retry.Option) (client.Client, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, clusterName}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetClientForCluster", varargs...)
	ret0, _ := ret[0].(client.Client)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetClientForCluster indicates an expected call of GetClientForCluster
func (mr *MockDynamicClientGetterMockRecorder) GetClientForCluster(ctx, clusterName interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, clusterName}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetClientForCluster", reflect.TypeOf((*MockDynamicClientGetter)(nil).GetClientForCluster), varargs...)
}
