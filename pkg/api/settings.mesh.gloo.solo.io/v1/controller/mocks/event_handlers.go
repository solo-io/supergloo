// Code generated by MockGen. DO NOT EDIT.
// Source: ./event_handlers.go

// Package mock_controller is a generated GoMock package.
package mock_controller

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	v1 "github.com/solo-io/gloo-mesh/pkg/api/settings.mesh.gloo.solo.io/v1"
	controller "github.com/solo-io/gloo-mesh/pkg/api/settings.mesh.gloo.solo.io/v1/controller"
	predicate "sigs.k8s.io/controller-runtime/pkg/predicate"
)

// MockSettingsEventHandler is a mock of SettingsEventHandler interface.
type MockSettingsEventHandler struct {
	ctrl     *gomock.Controller
	recorder *MockSettingsEventHandlerMockRecorder
}

// MockSettingsEventHandlerMockRecorder is the mock recorder for MockSettingsEventHandler.
type MockSettingsEventHandlerMockRecorder struct {
	mock *MockSettingsEventHandler
}

// NewMockSettingsEventHandler creates a new mock instance.
func NewMockSettingsEventHandler(ctrl *gomock.Controller) *MockSettingsEventHandler {
	mock := &MockSettingsEventHandler{ctrl: ctrl}
	mock.recorder = &MockSettingsEventHandlerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockSettingsEventHandler) EXPECT() *MockSettingsEventHandlerMockRecorder {
	return m.recorder
}

// CreateSettings mocks base method.
func (m *MockSettingsEventHandler) CreateSettings(obj *v1.Settings) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateSettings", obj)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateSettings indicates an expected call of CreateSettings.
func (mr *MockSettingsEventHandlerMockRecorder) CreateSettings(obj interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateSettings", reflect.TypeOf((*MockSettingsEventHandler)(nil).CreateSettings), obj)
}

// DeleteSettings mocks base method.
func (m *MockSettingsEventHandler) DeleteSettings(obj *v1.Settings) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteSettings", obj)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteSettings indicates an expected call of DeleteSettings.
func (mr *MockSettingsEventHandlerMockRecorder) DeleteSettings(obj interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteSettings", reflect.TypeOf((*MockSettingsEventHandler)(nil).DeleteSettings), obj)
}

// GenericSettings mocks base method.
func (m *MockSettingsEventHandler) GenericSettings(obj *v1.Settings) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GenericSettings", obj)
	ret0, _ := ret[0].(error)
	return ret0
}

// GenericSettings indicates an expected call of GenericSettings.
func (mr *MockSettingsEventHandlerMockRecorder) GenericSettings(obj interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GenericSettings", reflect.TypeOf((*MockSettingsEventHandler)(nil).GenericSettings), obj)
}

// UpdateSettings mocks base method.
func (m *MockSettingsEventHandler) UpdateSettings(old, new *v1.Settings) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateSettings", old, new)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateSettings indicates an expected call of UpdateSettings.
func (mr *MockSettingsEventHandlerMockRecorder) UpdateSettings(old, new interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateSettings", reflect.TypeOf((*MockSettingsEventHandler)(nil).UpdateSettings), old, new)
}

// MockSettingsEventWatcher is a mock of SettingsEventWatcher interface.
type MockSettingsEventWatcher struct {
	ctrl     *gomock.Controller
	recorder *MockSettingsEventWatcherMockRecorder
}

// MockSettingsEventWatcherMockRecorder is the mock recorder for MockSettingsEventWatcher.
type MockSettingsEventWatcherMockRecorder struct {
	mock *MockSettingsEventWatcher
}

// NewMockSettingsEventWatcher creates a new mock instance.
func NewMockSettingsEventWatcher(ctrl *gomock.Controller) *MockSettingsEventWatcher {
	mock := &MockSettingsEventWatcher{ctrl: ctrl}
	mock.recorder = &MockSettingsEventWatcherMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockSettingsEventWatcher) EXPECT() *MockSettingsEventWatcherMockRecorder {
	return m.recorder
}

// AddEventHandler mocks base method.
func (m *MockSettingsEventWatcher) AddEventHandler(ctx context.Context, h controller.SettingsEventHandler, predicates ...predicate.Predicate) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, h}
	for _, a := range predicates {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "AddEventHandler", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// AddEventHandler indicates an expected call of AddEventHandler.
func (mr *MockSettingsEventWatcherMockRecorder) AddEventHandler(ctx, h interface{}, predicates ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, h}, predicates...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddEventHandler", reflect.TypeOf((*MockSettingsEventWatcher)(nil).AddEventHandler), varargs...)
}

// MockDashboardEventHandler is a mock of DashboardEventHandler interface.
type MockDashboardEventHandler struct {
	ctrl     *gomock.Controller
	recorder *MockDashboardEventHandlerMockRecorder
}

// MockDashboardEventHandlerMockRecorder is the mock recorder for MockDashboardEventHandler.
type MockDashboardEventHandlerMockRecorder struct {
	mock *MockDashboardEventHandler
}

// NewMockDashboardEventHandler creates a new mock instance.
func NewMockDashboardEventHandler(ctrl *gomock.Controller) *MockDashboardEventHandler {
	mock := &MockDashboardEventHandler{ctrl: ctrl}
	mock.recorder = &MockDashboardEventHandlerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockDashboardEventHandler) EXPECT() *MockDashboardEventHandlerMockRecorder {
	return m.recorder
}

// CreateDashboard mocks base method.
func (m *MockDashboardEventHandler) CreateDashboard(obj *v1.Dashboard) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateDashboard", obj)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateDashboard indicates an expected call of CreateDashboard.
func (mr *MockDashboardEventHandlerMockRecorder) CreateDashboard(obj interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateDashboard", reflect.TypeOf((*MockDashboardEventHandler)(nil).CreateDashboard), obj)
}

// DeleteDashboard mocks base method.
func (m *MockDashboardEventHandler) DeleteDashboard(obj *v1.Dashboard) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteDashboard", obj)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteDashboard indicates an expected call of DeleteDashboard.
func (mr *MockDashboardEventHandlerMockRecorder) DeleteDashboard(obj interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteDashboard", reflect.TypeOf((*MockDashboardEventHandler)(nil).DeleteDashboard), obj)
}

// GenericDashboard mocks base method.
func (m *MockDashboardEventHandler) GenericDashboard(obj *v1.Dashboard) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GenericDashboard", obj)
	ret0, _ := ret[0].(error)
	return ret0
}

// GenericDashboard indicates an expected call of GenericDashboard.
func (mr *MockDashboardEventHandlerMockRecorder) GenericDashboard(obj interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GenericDashboard", reflect.TypeOf((*MockDashboardEventHandler)(nil).GenericDashboard), obj)
}

// UpdateDashboard mocks base method.
func (m *MockDashboardEventHandler) UpdateDashboard(old, new *v1.Dashboard) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateDashboard", old, new)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateDashboard indicates an expected call of UpdateDashboard.
func (mr *MockDashboardEventHandlerMockRecorder) UpdateDashboard(old, new interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateDashboard", reflect.TypeOf((*MockDashboardEventHandler)(nil).UpdateDashboard), old, new)
}

// MockDashboardEventWatcher is a mock of DashboardEventWatcher interface.
type MockDashboardEventWatcher struct {
	ctrl     *gomock.Controller
	recorder *MockDashboardEventWatcherMockRecorder
}

// MockDashboardEventWatcherMockRecorder is the mock recorder for MockDashboardEventWatcher.
type MockDashboardEventWatcherMockRecorder struct {
	mock *MockDashboardEventWatcher
}

// NewMockDashboardEventWatcher creates a new mock instance.
func NewMockDashboardEventWatcher(ctrl *gomock.Controller) *MockDashboardEventWatcher {
	mock := &MockDashboardEventWatcher{ctrl: ctrl}
	mock.recorder = &MockDashboardEventWatcherMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockDashboardEventWatcher) EXPECT() *MockDashboardEventWatcherMockRecorder {
	return m.recorder
}

// AddEventHandler mocks base method.
func (m *MockDashboardEventWatcher) AddEventHandler(ctx context.Context, h controller.DashboardEventHandler, predicates ...predicate.Predicate) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, h}
	for _, a := range predicates {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "AddEventHandler", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// AddEventHandler indicates an expected call of AddEventHandler.
func (mr *MockDashboardEventWatcherMockRecorder) AddEventHandler(ctx, h interface{}, predicates ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, h}, predicates...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddEventHandler", reflect.TypeOf((*MockDashboardEventWatcher)(nil).AddEventHandler), varargs...)
}
