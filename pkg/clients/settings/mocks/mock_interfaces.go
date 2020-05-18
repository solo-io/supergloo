// Code generated by MockGen. DO NOT EDIT.
// Source: ./interfaces.go

// Package mock_settings is a generated GoMock package.
package mock_settings

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
)

// MockSettingsHelperClient is a mock of SettingsHelperClient interface.
type MockSettingsHelperClient struct {
	ctrl     *gomock.Controller
	recorder *MockSettingsHelperClientMockRecorder
}

// MockSettingsHelperClientMockRecorder is the mock recorder for MockSettingsHelperClient.
type MockSettingsHelperClientMockRecorder struct {
	mock *MockSettingsHelperClient
}

// NewMockSettingsHelperClient creates a new mock instance.
func NewMockSettingsHelperClient(ctrl *gomock.Controller) *MockSettingsHelperClient {
	mock := &MockSettingsHelperClient{ctrl: ctrl}
	mock.recorder = &MockSettingsHelperClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockSettingsHelperClient) EXPECT() *MockSettingsHelperClientMockRecorder {
	return m.recorder
}

// GetAWSSettingsForAccount mocks base method.
func (m *MockSettingsHelperClient) GetAWSSettingsForAccount(ctx context.Context, accountId string) (*types.SettingsSpec_AwsAccount, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAWSSettingsForAccount", ctx, accountId)
	ret0, _ := ret[0].(*types.SettingsSpec_AwsAccount)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAWSSettingsForAccount indicates an expected call of GetAWSSettingsForAccount.
func (mr *MockSettingsHelperClientMockRecorder) GetAWSSettingsForAccount(ctx, accountId interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAWSSettingsForAccount", reflect.TypeOf((*MockSettingsHelperClient)(nil).GetAWSSettingsForAccount), ctx, accountId)
}
