// Code generated by MockGen. DO NOT EDIT.
// Source: ./interfaces.go

// Package mock_aws_utils is a generated GoMock package.
package mock_aws_utils

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	aws_utils "github.com/solo-io/service-mesh-hub/pkg/aws/parser"
	v1 "k8s.io/api/core/v1"
	client "sigs.k8s.io/controller-runtime/pkg/client"
)

// MockAppMeshScanner is a mock of AppMeshScanner interface
type MockAppMeshScanner struct {
	ctrl     *gomock.Controller
	recorder *MockAppMeshScannerMockRecorder
}

// MockAppMeshScannerMockRecorder is the mock recorder for MockAppMeshScanner
type MockAppMeshScannerMockRecorder struct {
	mock *MockAppMeshScanner
}

// NewMockAppMeshScanner creates a new mock instance
func NewMockAppMeshScanner(ctrl *gomock.Controller) *MockAppMeshScanner {
	mock := &MockAppMeshScanner{ctrl: ctrl}
	mock.recorder = &MockAppMeshScannerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockAppMeshScanner) EXPECT() *MockAppMeshScannerMockRecorder {
	return m.recorder
}

// ScanPodForAppMesh mocks base method
func (m *MockAppMeshScanner) ScanPodForAppMesh(pod *v1.Pod, awsAccountId aws_utils.AwsAccountId) (*aws_utils.AppMeshPod, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ScanPodForAppMesh", pod, awsAccountId)
	ret0, _ := ret[0].(*aws_utils.AppMeshPod)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ScanPodForAppMesh indicates an expected call of ScanPodForAppMesh
func (mr *MockAppMeshScannerMockRecorder) ScanPodForAppMesh(pod, awsAccountId interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ScanPodForAppMesh", reflect.TypeOf((*MockAppMeshScanner)(nil).ScanPodForAppMesh), pod, awsAccountId)
}

// MockArnParser is a mock of ArnParser interface
type MockArnParser struct {
	ctrl     *gomock.Controller
	recorder *MockArnParserMockRecorder
}

// MockArnParserMockRecorder is the mock recorder for MockArnParser
type MockArnParserMockRecorder struct {
	mock *MockArnParser
}

// NewMockArnParser creates a new mock instance
func NewMockArnParser(ctrl *gomock.Controller) *MockArnParser {
	mock := &MockArnParser{ctrl: ctrl}
	mock.recorder = &MockArnParserMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockArnParser) EXPECT() *MockArnParserMockRecorder {
	return m.recorder
}

// ParseAccountID mocks base method
func (m *MockArnParser) ParseAccountID(arn string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ParseAccountID", arn)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ParseAccountID indicates an expected call of ParseAccountID
func (mr *MockArnParserMockRecorder) ParseAccountID(arn interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ParseAccountID", reflect.TypeOf((*MockArnParser)(nil).ParseAccountID), arn)
}

// ParseRegion mocks base method
func (m *MockArnParser) ParseRegion(arn string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ParseRegion", arn)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ParseRegion indicates an expected call of ParseRegion
func (mr *MockArnParserMockRecorder) ParseRegion(arn interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ParseRegion", reflect.TypeOf((*MockArnParser)(nil).ParseRegion), arn)
}

// MockAwsAccountIdFetcher is a mock of AwsAccountIdFetcher interface
type MockAwsAccountIdFetcher struct {
	ctrl     *gomock.Controller
	recorder *MockAwsAccountIdFetcherMockRecorder
}

// MockAwsAccountIdFetcherMockRecorder is the mock recorder for MockAwsAccountIdFetcher
type MockAwsAccountIdFetcherMockRecorder struct {
	mock *MockAwsAccountIdFetcher
}

// NewMockAwsAccountIdFetcher creates a new mock instance
func NewMockAwsAccountIdFetcher(ctrl *gomock.Controller) *MockAwsAccountIdFetcher {
	mock := &MockAwsAccountIdFetcher{ctrl: ctrl}
	mock.recorder = &MockAwsAccountIdFetcherMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockAwsAccountIdFetcher) EXPECT() *MockAwsAccountIdFetcherMockRecorder {
	return m.recorder
}

// GetEksAccountId mocks base method
func (m *MockAwsAccountIdFetcher) GetEksAccountId(ctx context.Context, clusterScopedClient client.Client) (aws_utils.AwsAccountId, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetEksAccountId", ctx, clusterScopedClient)
	ret0, _ := ret[0].(aws_utils.AwsAccountId)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetEksAccountId indicates an expected call of GetEksAccountId
func (mr *MockAwsAccountIdFetcherMockRecorder) GetEksAccountId(ctx, clusterScopedClient interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetEksAccountId", reflect.TypeOf((*MockAwsAccountIdFetcher)(nil).GetEksAccountId), ctx, clusterScopedClient)
}
