// Code generated by MockGen. DO NOT EDIT.
// Source: ./interfaces.go

// Package mock_kubernetes_core is a generated GoMock package.
package mock_kubernetes_core

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	v1 "k8s.io/api/core/v1"
	client "sigs.k8s.io/controller-runtime/pkg/client"
)

// MockServiceClient is a mock of ServiceClient interface.
type MockServiceClient struct {
	ctrl     *gomock.Controller
	recorder *MockServiceClientMockRecorder
}

// MockServiceClientMockRecorder is the mock recorder for MockServiceClient.
type MockServiceClientMockRecorder struct {
	mock *MockServiceClient
}

// NewMockServiceClient creates a new mock instance.
func NewMockServiceClient(ctrl *gomock.Controller) *MockServiceClient {
	mock := &MockServiceClient{ctrl: ctrl}
	mock.recorder = &MockServiceClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockServiceClient) EXPECT() *MockServiceClientMockRecorder {
	return m.recorder
}

// Get mocks base method.
func (m *MockServiceClient) Get(ctx context.Context, name, namespace string) (*v1.Service, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", ctx, name, namespace)
	ret0, _ := ret[0].(*v1.Service)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get.
func (mr *MockServiceClientMockRecorder) Get(ctx, name, namespace interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockServiceClient)(nil).Get), ctx, name, namespace)
}

// List mocks base method.
func (m *MockServiceClient) List(ctx context.Context, options ...client.ListOption) (*v1.ServiceList, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx}
	for _, a := range options {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "List", varargs...)
	ret0, _ := ret[0].(*v1.ServiceList)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// List indicates an expected call of List.
func (mr *MockServiceClientMockRecorder) List(ctx interface{}, options ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx}, options...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "List", reflect.TypeOf((*MockServiceClient)(nil).List), varargs...)
}

// MockPodClient is a mock of PodClient interface.
type MockPodClient struct {
	ctrl     *gomock.Controller
	recorder *MockPodClientMockRecorder
}

// MockPodClientMockRecorder is the mock recorder for MockPodClient.
type MockPodClientMockRecorder struct {
	mock *MockPodClient
}

// NewMockPodClient creates a new mock instance.
func NewMockPodClient(ctrl *gomock.Controller) *MockPodClient {
	mock := &MockPodClient{ctrl: ctrl}
	mock.recorder = &MockPodClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockPodClient) EXPECT() *MockPodClientMockRecorder {
	return m.recorder
}

// Get mocks base method.
func (m *MockPodClient) Get(ctx context.Context, name, namespace string) (*v1.Pod, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", ctx, name, namespace)
	ret0, _ := ret[0].(*v1.Pod)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get.
func (mr *MockPodClientMockRecorder) Get(ctx, name, namespace interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockPodClient)(nil).Get), ctx, name, namespace)
}

// List mocks base method.
func (m *MockPodClient) List(ctx context.Context, options ...client.ListOption) (*v1.PodList, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx}
	for _, a := range options {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "List", varargs...)
	ret0, _ := ret[0].(*v1.PodList)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// List indicates an expected call of List.
func (mr *MockPodClientMockRecorder) List(ctx interface{}, options ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx}, options...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "List", reflect.TypeOf((*MockPodClient)(nil).List), varargs...)
}

// MockNodeClient is a mock of NodeClient interface.
type MockNodeClient struct {
	ctrl     *gomock.Controller
	recorder *MockNodeClientMockRecorder
}

// MockNodeClientMockRecorder is the mock recorder for MockNodeClient.
type MockNodeClientMockRecorder struct {
	mock *MockNodeClient
}

// NewMockNodeClient creates a new mock instance.
func NewMockNodeClient(ctrl *gomock.Controller) *MockNodeClient {
	mock := &MockNodeClient{ctrl: ctrl}
	mock.recorder = &MockNodeClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockNodeClient) EXPECT() *MockNodeClientMockRecorder {
	return m.recorder
}

// Get mocks base method.
func (m *MockNodeClient) Get(ctx context.Context, name string) (*v1.Node, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", ctx, name)
	ret0, _ := ret[0].(*v1.Node)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get.
func (mr *MockNodeClientMockRecorder) Get(ctx, name interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockNodeClient)(nil).Get), ctx, name)
}

// List mocks base method.
func (m *MockNodeClient) List(ctx context.Context, options ...client.ListOption) (*v1.NodeList, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx}
	for _, a := range options {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "List", varargs...)
	ret0, _ := ret[0].(*v1.NodeList)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// List indicates an expected call of List.
func (mr *MockNodeClientMockRecorder) List(ctx interface{}, options ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx}, options...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "List", reflect.TypeOf((*MockNodeClient)(nil).List), varargs...)
}

// MockSecretsClient is a mock of SecretsClient interface.
type MockSecretsClient struct {
	ctrl     *gomock.Controller
	recorder *MockSecretsClientMockRecorder
}

// MockSecretsClientMockRecorder is the mock recorder for MockSecretsClient.
type MockSecretsClientMockRecorder struct {
	mock *MockSecretsClient
}

// NewMockSecretsClient creates a new mock instance.
func NewMockSecretsClient(ctrl *gomock.Controller) *MockSecretsClient {
	mock := &MockSecretsClient{ctrl: ctrl}
	mock.recorder = &MockSecretsClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockSecretsClient) EXPECT() *MockSecretsClientMockRecorder {
	return m.recorder
}

// Create mocks base method.
func (m *MockSecretsClient) Create(ctx context.Context, secret *v1.Secret, opts ...client.CreateOption) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, secret}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Create", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// Create indicates an expected call of Create.
func (mr *MockSecretsClientMockRecorder) Create(ctx, secret interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, secret}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockSecretsClient)(nil).Create), varargs...)
}

// Update mocks base method.
func (m *MockSecretsClient) Update(ctx context.Context, secret *v1.Secret, opts ...client.UpdateOption) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, secret}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Update", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// Update indicates an expected call of Update.
func (mr *MockSecretsClientMockRecorder) Update(ctx, secret interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, secret}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Update", reflect.TypeOf((*MockSecretsClient)(nil).Update), varargs...)
}

// UpsertData mocks base method.
func (m *MockSecretsClient) UpsertData(ctx context.Context, secret *v1.Secret) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpsertData", ctx, secret)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpsertData indicates an expected call of UpsertData.
func (mr *MockSecretsClientMockRecorder) UpsertData(ctx, secret interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpsertData", reflect.TypeOf((*MockSecretsClient)(nil).UpsertData), ctx, secret)
}

// Get mocks base method.
func (m *MockSecretsClient) Get(ctx context.Context, name, namespace string) (*v1.Secret, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", ctx, name, namespace)
	ret0, _ := ret[0].(*v1.Secret)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get.
func (mr *MockSecretsClientMockRecorder) Get(ctx, name, namespace interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockSecretsClient)(nil).Get), ctx, name, namespace)
}

// List mocks base method.
func (m *MockSecretsClient) List(ctx context.Context, namespace string, labels map[string]string) (*v1.SecretList, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "List", ctx, namespace, labels)
	ret0, _ := ret[0].(*v1.SecretList)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// List indicates an expected call of List.
func (mr *MockSecretsClientMockRecorder) List(ctx, namespace, labels interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "List", reflect.TypeOf((*MockSecretsClient)(nil).List), ctx, namespace, labels)
}

// Delete mocks base method.
func (m *MockSecretsClient) Delete(ctx context.Context, secret *v1.Secret) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Delete", ctx, secret)
	ret0, _ := ret[0].(error)
	return ret0
}

// Delete indicates an expected call of Delete.
func (mr *MockSecretsClientMockRecorder) Delete(ctx, secret interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockSecretsClient)(nil).Delete), ctx, secret)
}

// MockServiceAccountClient is a mock of ServiceAccountClient interface.
type MockServiceAccountClient struct {
	ctrl     *gomock.Controller
	recorder *MockServiceAccountClientMockRecorder
}

// MockServiceAccountClientMockRecorder is the mock recorder for MockServiceAccountClient.
type MockServiceAccountClientMockRecorder struct {
	mock *MockServiceAccountClient
}

// NewMockServiceAccountClient creates a new mock instance.
func NewMockServiceAccountClient(ctrl *gomock.Controller) *MockServiceAccountClient {
	mock := &MockServiceAccountClient{ctrl: ctrl}
	mock.recorder = &MockServiceAccountClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockServiceAccountClient) EXPECT() *MockServiceAccountClientMockRecorder {
	return m.recorder
}

// Create mocks base method.
func (m *MockServiceAccountClient) Create(ctx context.Context, serviceAccount *v1.ServiceAccount) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Create", ctx, serviceAccount)
	ret0, _ := ret[0].(error)
	return ret0
}

// Create indicates an expected call of Create.
func (mr *MockServiceAccountClientMockRecorder) Create(ctx, serviceAccount interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockServiceAccountClient)(nil).Create), ctx, serviceAccount)
}

// Get mocks base method.
func (m *MockServiceAccountClient) Get(ctx context.Context, name, namespace string) (*v1.ServiceAccount, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", ctx, name, namespace)
	ret0, _ := ret[0].(*v1.ServiceAccount)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get.
func (mr *MockServiceAccountClientMockRecorder) Get(ctx, name, namespace interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockServiceAccountClient)(nil).Get), ctx, name, namespace)
}

// Update mocks base method.
func (m *MockServiceAccountClient) Update(ctx context.Context, serviceAccount *v1.ServiceAccount) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Update", ctx, serviceAccount)
	ret0, _ := ret[0].(error)
	return ret0
}

// Update indicates an expected call of Update.
func (mr *MockServiceAccountClientMockRecorder) Update(ctx, serviceAccount interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Update", reflect.TypeOf((*MockServiceAccountClient)(nil).Update), ctx, serviceAccount)
}

// List mocks base method.
func (m *MockServiceAccountClient) List(ctx context.Context, options ...client.ListOption) (*v1.ServiceAccountList, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx}
	for _, a := range options {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "List", varargs...)
	ret0, _ := ret[0].(*v1.ServiceAccountList)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// List indicates an expected call of List.
func (mr *MockServiceAccountClientMockRecorder) List(ctx interface{}, options ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx}, options...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "List", reflect.TypeOf((*MockServiceAccountClient)(nil).List), varargs...)
}

// MockConfigMapClient is a mock of ConfigMapClient interface.
type MockConfigMapClient struct {
	ctrl     *gomock.Controller
	recorder *MockConfigMapClientMockRecorder
}

// MockConfigMapClientMockRecorder is the mock recorder for MockConfigMapClient.
type MockConfigMapClientMockRecorder struct {
	mock *MockConfigMapClient
}

// NewMockConfigMapClient creates a new mock instance.
func NewMockConfigMapClient(ctrl *gomock.Controller) *MockConfigMapClient {
	mock := &MockConfigMapClient{ctrl: ctrl}
	mock.recorder = &MockConfigMapClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockConfigMapClient) EXPECT() *MockConfigMapClientMockRecorder {
	return m.recorder
}

// Create mocks base method.
func (m *MockConfigMapClient) Create(ctx context.Context, configMap *v1.ConfigMap) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Create", ctx, configMap)
	ret0, _ := ret[0].(error)
	return ret0
}

// Create indicates an expected call of Create.
func (mr *MockConfigMapClientMockRecorder) Create(ctx, configMap interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockConfigMapClient)(nil).Create), ctx, configMap)
}

// Get mocks base method.
func (m *MockConfigMapClient) Get(ctx context.Context, objKey client.ObjectKey) (*v1.ConfigMap, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", ctx, objKey)
	ret0, _ := ret[0].(*v1.ConfigMap)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get.
func (mr *MockConfigMapClientMockRecorder) Get(ctx, objKey interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockConfigMapClient)(nil).Get), ctx, objKey)
}

// Update mocks base method.
func (m *MockConfigMapClient) Update(ctx context.Context, configMap *v1.ConfigMap) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Update", ctx, configMap)
	ret0, _ := ret[0].(error)
	return ret0
}

// Update indicates an expected call of Update.
func (mr *MockConfigMapClientMockRecorder) Update(ctx, configMap interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Update", reflect.TypeOf((*MockConfigMapClient)(nil).Update), ctx, configMap)
}

// MockNamespaceClient is a mock of NamespaceClient interface.
type MockNamespaceClient struct {
	ctrl     *gomock.Controller
	recorder *MockNamespaceClientMockRecorder
}

// MockNamespaceClientMockRecorder is the mock recorder for MockNamespaceClient.
type MockNamespaceClientMockRecorder struct {
	mock *MockNamespaceClient
}

// NewMockNamespaceClient creates a new mock instance.
func NewMockNamespaceClient(ctrl *gomock.Controller) *MockNamespaceClient {
	mock := &MockNamespaceClient{ctrl: ctrl}
	mock.recorder = &MockNamespaceClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockNamespaceClient) EXPECT() *MockNamespaceClientMockRecorder {
	return m.recorder
}

// Create mocks base method.
func (m *MockNamespaceClient) Create(ctx context.Context, ns *v1.Namespace) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Create", ctx, ns)
	ret0, _ := ret[0].(error)
	return ret0
}

// Create indicates an expected call of Create.
func (mr *MockNamespaceClientMockRecorder) Create(ctx, ns interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockNamespaceClient)(nil).Create), ctx, ns)
}

// Get mocks base method.
func (m *MockNamespaceClient) Get(ctx context.Context, name string) (*v1.Namespace, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", ctx, name)
	ret0, _ := ret[0].(*v1.Namespace)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get.
func (mr *MockNamespaceClientMockRecorder) Get(ctx, name interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockNamespaceClient)(nil).Get), ctx, name)
}

// Delete mocks base method.
func (m *MockNamespaceClient) Delete(ctx context.Context, name string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Delete", ctx, name)
	ret0, _ := ret[0].(error)
	return ret0
}

// Delete indicates an expected call of Delete.
func (mr *MockNamespaceClientMockRecorder) Delete(ctx, name interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockNamespaceClient)(nil).Delete), ctx, name)
}
