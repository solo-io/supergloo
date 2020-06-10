// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/core/v1/controller (interfaces: PodEventWatcher)

// Package mock_controllers is a generated GoMock package.
package mock_controllers

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	controller "github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/core/v1/controller"
	predicate "sigs.k8s.io/controller-runtime/pkg/predicate"
)

// MockPodEventWatcher is a mock of PodEventWatcher interface
type MockPodEventWatcher struct {
	ctrl     *gomock.Controller
	recorder *MockPodEventWatcherMockRecorder
}

// MockPodEventWatcherMockRecorder is the mock recorder for MockPodEventWatcher
type MockPodEventWatcherMockRecorder struct {
	mock *MockPodEventWatcher
}

// NewMockPodEventWatcher creates a new mock instance
func NewMockPodEventWatcher(ctrl *gomock.Controller) *MockPodEventWatcher {
	mock := &MockPodEventWatcher{ctrl: ctrl}
	mock.recorder = &MockPodEventWatcherMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockPodEventWatcher) EXPECT() *MockPodEventWatcherMockRecorder {
	return m.recorder
}

// AddEventHandler mocks base method
func (m *MockPodEventWatcher) AddEventHandler(arg0 context.Context, arg1 controller.PodEventHandler, arg2 ...predicate.Predicate) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "AddEventHandler", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// AddEventHandler indicates an expected call of AddEventHandler
func (mr *MockPodEventWatcherMockRecorder) AddEventHandler(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddEventHandler", reflect.TypeOf((*MockPodEventWatcher)(nil).AddEventHandler), varargs...)
}
