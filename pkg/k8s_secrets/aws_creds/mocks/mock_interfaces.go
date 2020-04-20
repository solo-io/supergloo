// Code generated by MockGen. DO NOT EDIT.
// Source: ./interfaces.go

// Package mock_aws_creds is a generated GoMock package.
package mock_aws_creds

import (
	credentials "github.com/aws/aws-sdk-go/aws/credentials"
	gomock "github.com/golang/mock/gomock"
	v1 "k8s.io/api/core/v1"
	reflect "reflect"
)

// MockSecretAwsCredsConverter is a mock of SecretAwsCredsConverter interface.
type MockSecretAwsCredsConverter struct {
	ctrl     *gomock.Controller
	recorder *MockSecretAwsCredsConverterMockRecorder
}

// MockSecretAwsCredsConverterMockRecorder is the mock recorder for MockSecretAwsCredsConverter.
type MockSecretAwsCredsConverterMockRecorder struct {
	mock *MockSecretAwsCredsConverter
}

// NewMockSecretAwsCredsConverter creates a new mock instance.
func NewMockSecretAwsCredsConverter(ctrl *gomock.Controller) *MockSecretAwsCredsConverter {
	mock := &MockSecretAwsCredsConverter{ctrl: ctrl}
	mock.recorder = &MockSecretAwsCredsConverterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockSecretAwsCredsConverter) EXPECT() *MockSecretAwsCredsConverterMockRecorder {
	return m.recorder
}

// CredsFileToSecret mocks base method.
func (m *MockSecretAwsCredsConverter) CredsFileToSecret(secretName, secretNamespace, credsFilename, credsProfile string) (*v1.Secret, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CredsFileToSecret", secretName, secretNamespace, credsFilename, credsProfile)
	ret0, _ := ret[0].(*v1.Secret)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CredsFileToSecret indicates an expected call of CredsFileToSecret.
func (mr *MockSecretAwsCredsConverterMockRecorder) CredsFileToSecret(secretName, secretNamespace, credsFilename, credsProfile interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CredsFileToSecret", reflect.TypeOf((*MockSecretAwsCredsConverter)(nil).CredsFileToSecret), secretName, secretNamespace, credsFilename, credsProfile)
}

// SecretToCreds mocks base method.
func (m *MockSecretAwsCredsConverter) SecretToCreds(secret *v1.Secret) (*credentials.Value, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SecretToCreds", secret)
	ret0, _ := ret[0].(*credentials.Value)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SecretToCreds indicates an expected call of SecretToCreds.
func (mr *MockSecretAwsCredsConverterMockRecorder) SecretToCreds(secret interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SecretToCreds", reflect.TypeOf((*MockSecretAwsCredsConverter)(nil).SecretToCreds), secret)
}
