// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/solo-io/service-mesh-hub/pkg/api/linkerd/v1alpha2 (interfaces: ServiceProfileClient)

// Package mock_linkerd_clients is a generated GoMock package.
package mock_linkerd_clients

import (
	context "context"
	gomock "github.com/golang/mock/gomock"
	v1alpha2 "github.com/linkerd/linkerd2/controller/gen/apis/serviceprofile/v1alpha2"
	types "k8s.io/apimachinery/pkg/types"
	reflect "reflect"
	client "sigs.k8s.io/controller-runtime/pkg/client"
)

// MockServiceProfileClient is a mock of ServiceProfileClient interface.
type MockServiceProfileClient struct {
	ctrl     *gomock.Controller
	recorder *MockServiceProfileClientMockRecorder
}

// MockServiceProfileClientMockRecorder is the mock recorder for MockServiceProfileClient.
type MockServiceProfileClientMockRecorder struct {
	mock *MockServiceProfileClient
}

// NewMockServiceProfileClient creates a new mock instance.
func NewMockServiceProfileClient(ctrl *gomock.Controller) *MockServiceProfileClient {
	mock := &MockServiceProfileClient{ctrl: ctrl}
	mock.recorder = &MockServiceProfileClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockServiceProfileClient) EXPECT() *MockServiceProfileClientMockRecorder {
	return m.recorder
}

// CreateServiceProfile mocks base method.
func (m *MockServiceProfileClient) CreateServiceProfile(arg0 context.Context, arg1 *v1alpha2.ServiceProfile, arg2 ...client.CreateOption) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "CreateServiceProfile", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateServiceProfile indicates an expected call of CreateServiceProfile.
func (mr *MockServiceProfileClientMockRecorder) CreateServiceProfile(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateServiceProfile", reflect.TypeOf((*MockServiceProfileClient)(nil).CreateServiceProfile), varargs...)
}

// DeleteAllOfServiceProfile mocks base method.
func (m *MockServiceProfileClient) DeleteAllOfServiceProfile(arg0 context.Context, arg1 ...client.DeleteAllOfOption) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0}
	for _, a := range arg1 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "DeleteAllOfServiceProfile", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteAllOfServiceProfile indicates an expected call of DeleteAllOfServiceProfile.
func (mr *MockServiceProfileClientMockRecorder) DeleteAllOfServiceProfile(arg0 interface{}, arg1 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0}, arg1...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteAllOfServiceProfile", reflect.TypeOf((*MockServiceProfileClient)(nil).DeleteAllOfServiceProfile), varargs...)
}

// DeleteServiceProfile mocks base method.
func (m *MockServiceProfileClient) DeleteServiceProfile(arg0 context.Context, arg1 types.NamespacedName, arg2 ...client.DeleteOption) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "DeleteServiceProfile", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteServiceProfile indicates an expected call of DeleteServiceProfile.
func (mr *MockServiceProfileClientMockRecorder) DeleteServiceProfile(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteServiceProfile", reflect.TypeOf((*MockServiceProfileClient)(nil).DeleteServiceProfile), varargs...)
}

// GetServiceProfile mocks base method.
func (m *MockServiceProfileClient) GetServiceProfile(arg0 context.Context, arg1 types.NamespacedName) (*v1alpha2.ServiceProfile, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetServiceProfile", arg0, arg1)
	ret0, _ := ret[0].(*v1alpha2.ServiceProfile)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetServiceProfile indicates an expected call of GetServiceProfile.
func (mr *MockServiceProfileClientMockRecorder) GetServiceProfile(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetServiceProfile", reflect.TypeOf((*MockServiceProfileClient)(nil).GetServiceProfile), arg0, arg1)
}

// ListServiceProfile mocks base method.
func (m *MockServiceProfileClient) ListServiceProfile(arg0 context.Context, arg1 ...client.ListOption) (*v1alpha2.ServiceProfileList, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0}
	for _, a := range arg1 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "ListServiceProfile", varargs...)
	ret0, _ := ret[0].(*v1alpha2.ServiceProfileList)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListServiceProfile indicates an expected call of ListServiceProfile.
func (mr *MockServiceProfileClientMockRecorder) ListServiceProfile(arg0 interface{}, arg1 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0}, arg1...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListServiceProfile", reflect.TypeOf((*MockServiceProfileClient)(nil).ListServiceProfile), varargs...)
}

// PatchServiceProfile mocks base method.
func (m *MockServiceProfileClient) PatchServiceProfile(arg0 context.Context, arg1 *v1alpha2.ServiceProfile, arg2 client.Patch, arg3 ...client.PatchOption) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1, arg2}
	for _, a := range arg3 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "PatchServiceProfile", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// PatchServiceProfile indicates an expected call of PatchServiceProfile.
func (mr *MockServiceProfileClientMockRecorder) PatchServiceProfile(arg0, arg1, arg2 interface{}, arg3 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1, arg2}, arg3...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PatchServiceProfile", reflect.TypeOf((*MockServiceProfileClient)(nil).PatchServiceProfile), varargs...)
}

// PatchServiceProfileStatus mocks base method.
func (m *MockServiceProfileClient) PatchServiceProfileStatus(arg0 context.Context, arg1 *v1alpha2.ServiceProfile, arg2 client.Patch, arg3 ...client.PatchOption) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1, arg2}
	for _, a := range arg3 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "PatchServiceProfileStatus", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// PatchServiceProfileStatus indicates an expected call of PatchServiceProfileStatus.
func (mr *MockServiceProfileClientMockRecorder) PatchServiceProfileStatus(arg0, arg1, arg2 interface{}, arg3 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1, arg2}, arg3...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PatchServiceProfileStatus", reflect.TypeOf((*MockServiceProfileClient)(nil).PatchServiceProfileStatus), varargs...)
}

// UpdateServiceProfile mocks base method.
func (m *MockServiceProfileClient) UpdateServiceProfile(arg0 context.Context, arg1 *v1alpha2.ServiceProfile, arg2 ...client.UpdateOption) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "UpdateServiceProfile", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateServiceProfile indicates an expected call of UpdateServiceProfile.
func (mr *MockServiceProfileClientMockRecorder) UpdateServiceProfile(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateServiceProfile", reflect.TypeOf((*MockServiceProfileClient)(nil).UpdateServiceProfile), varargs...)
}

// UpdateServiceProfileStatus mocks base method.
func (m *MockServiceProfileClient) UpdateServiceProfileStatus(arg0 context.Context, arg1 *v1alpha2.ServiceProfile, arg2 ...client.UpdateOption) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "UpdateServiceProfileStatus", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateServiceProfileStatus indicates an expected call of UpdateServiceProfileStatus.
func (mr *MockServiceProfileClientMockRecorder) UpdateServiceProfileStatus(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateServiceProfileStatus", reflect.TypeOf((*MockServiceProfileClient)(nil).UpdateServiceProfileStatus), varargs...)
}

// UpsertServiceProfileSpec mocks base method.
func (m *MockServiceProfileClient) UpsertServiceProfileSpec(arg0 context.Context, arg1 *v1alpha2.ServiceProfile, arg2 ...client.UpdateOption) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "UpsertServiceProfileSpec", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpsertServiceProfileSpec indicates an expected call of UpsertServiceProfileSpec.
func (mr *MockServiceProfileClientMockRecorder) UpsertServiceProfileSpec(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpsertServiceProfileSpec", reflect.TypeOf((*MockServiceProfileClient)(nil).UpsertServiceProfileSpec), varargs...)
}
