// Code generated by MockGen. DO NOT EDIT.
// Source: sigs.k8s.io/controller-runtime/pkg/manager (interfaces: Manager)

// Package mock_controller_runtime is a generated GoMock package.
package mock_controller_runtime

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	meta "k8s.io/apimachinery/pkg/api/meta"
	runtime "k8s.io/apimachinery/pkg/runtime"
	rest "k8s.io/client-go/rest"
	record "k8s.io/client-go/tools/record"
	cache "sigs.k8s.io/controller-runtime/pkg/cache"
	client "sigs.k8s.io/controller-runtime/pkg/client"
	healthz "sigs.k8s.io/controller-runtime/pkg/healthz"
	manager "sigs.k8s.io/controller-runtime/pkg/manager"
	webhook "sigs.k8s.io/controller-runtime/pkg/webhook"
)

// MockManager is a mock of Manager interface
type MockManager struct {
	ctrl     *gomock.Controller
	recorder *MockManagerMockRecorder
}

// MockManagerMockRecorder is the mock recorder for MockManager
type MockManagerMockRecorder struct {
	mock *MockManager
}

// NewMockManager creates a new mock instance
func NewMockManager(ctrl *gomock.Controller) *MockManager {
	mock := &MockManager{ctrl: ctrl}
	mock.recorder = &MockManagerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockManager) EXPECT() *MockManagerMockRecorder {
	return m.recorder
}

// Add mocks base method
func (m *MockManager) Add(arg0 manager.Runnable) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Add", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Add indicates an expected call of Add
func (mr *MockManagerMockRecorder) Add(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Add", reflect.TypeOf((*MockManager)(nil).Add), arg0)
}

// AddHealthzCheck mocks base method
func (m *MockManager) AddHealthzCheck(arg0 string, arg1 healthz.Checker) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddHealthzCheck", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// AddHealthzCheck indicates an expected call of AddHealthzCheck
func (mr *MockManagerMockRecorder) AddHealthzCheck(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddHealthzCheck", reflect.TypeOf((*MockManager)(nil).AddHealthzCheck), arg0, arg1)
}

// AddReadyzCheck mocks base method
func (m *MockManager) AddReadyzCheck(arg0 string, arg1 healthz.Checker) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddReadyzCheck", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// AddReadyzCheck indicates an expected call of AddReadyzCheck
func (mr *MockManagerMockRecorder) AddReadyzCheck(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddReadyzCheck", reflect.TypeOf((*MockManager)(nil).AddReadyzCheck), arg0, arg1)
}

// GetAPIReader mocks base method
func (m *MockManager) GetAPIReader() client.Reader {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAPIReader")
	ret0, _ := ret[0].(client.Reader)
	return ret0
}

// GetAPIReader indicates an expected call of GetAPIReader
func (mr *MockManagerMockRecorder) GetAPIReader() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAPIReader", reflect.TypeOf((*MockManager)(nil).GetAPIReader))
}

// GetCache mocks base method
func (m *MockManager) GetCache() cache.Cache {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetCache")
	ret0, _ := ret[0].(cache.Cache)
	return ret0
}

// GetCache indicates an expected call of GetCache
func (mr *MockManagerMockRecorder) GetCache() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetCache", reflect.TypeOf((*MockManager)(nil).GetCache))
}

// GetClient mocks base method
func (m *MockManager) GetClient() client.Client {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetClient")
	ret0, _ := ret[0].(client.Client)
	return ret0
}

// GetClient indicates an expected call of GetClient
func (mr *MockManagerMockRecorder) GetClient() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetClient", reflect.TypeOf((*MockManager)(nil).GetClient))
}

// GetConfig mocks base method
func (m *MockManager) GetConfig() *rest.Config {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetConfig")
	ret0, _ := ret[0].(*rest.Config)
	return ret0
}

// GetConfig indicates an expected call of GetConfig
func (mr *MockManagerMockRecorder) GetConfig() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetConfig", reflect.TypeOf((*MockManager)(nil).GetConfig))
}

// GetEventRecorderFor mocks base method
func (m *MockManager) GetEventRecorderFor(arg0 string) record.EventRecorder {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetEventRecorderFor", arg0)
	ret0, _ := ret[0].(record.EventRecorder)
	return ret0
}

// GetEventRecorderFor indicates an expected call of GetEventRecorderFor
func (mr *MockManagerMockRecorder) GetEventRecorderFor(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetEventRecorderFor", reflect.TypeOf((*MockManager)(nil).GetEventRecorderFor), arg0)
}

// GetFieldIndexer mocks base method
func (m *MockManager) GetFieldIndexer() client.FieldIndexer {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetFieldIndexer")
	ret0, _ := ret[0].(client.FieldIndexer)
	return ret0
}

// GetFieldIndexer indicates an expected call of GetFieldIndexer
func (mr *MockManagerMockRecorder) GetFieldIndexer() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetFieldIndexer", reflect.TypeOf((*MockManager)(nil).GetFieldIndexer))
}

// GetRESTMapper mocks base method
func (m *MockManager) GetRESTMapper() meta.RESTMapper {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetRESTMapper")
	ret0, _ := ret[0].(meta.RESTMapper)
	return ret0
}

// GetRESTMapper indicates an expected call of GetRESTMapper
func (mr *MockManagerMockRecorder) GetRESTMapper() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetRESTMapper", reflect.TypeOf((*MockManager)(nil).GetRESTMapper))
}

// GetScheme mocks base method
func (m *MockManager) GetScheme() *runtime.Scheme {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetScheme")
	ret0, _ := ret[0].(*runtime.Scheme)
	return ret0
}

// GetScheme indicates an expected call of GetScheme
func (mr *MockManagerMockRecorder) GetScheme() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetScheme", reflect.TypeOf((*MockManager)(nil).GetScheme))
}

// GetWebhookServer mocks base method
func (m *MockManager) GetWebhookServer() *webhook.Server {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetWebhookServer")
	ret0, _ := ret[0].(*webhook.Server)
	return ret0
}

// GetWebhookServer indicates an expected call of GetWebhookServer
func (mr *MockManagerMockRecorder) GetWebhookServer() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetWebhookServer", reflect.TypeOf((*MockManager)(nil).GetWebhookServer))
}

// SetFields mocks base method
func (m *MockManager) SetFields(arg0 interface{}) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetFields", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetFields indicates an expected call of SetFields
func (mr *MockManagerMockRecorder) SetFields(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetFields", reflect.TypeOf((*MockManager)(nil).SetFields), arg0)
}

// Start mocks base method
func (m *MockManager) Start(arg0 <-chan struct{}) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Start", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Start indicates an expected call of Start
func (mr *MockManagerMockRecorder) Start(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Start", reflect.TypeOf((*MockManager)(nil).Start), arg0)
}
