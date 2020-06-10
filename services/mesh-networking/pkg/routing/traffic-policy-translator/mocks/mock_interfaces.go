// Code generated by MockGen. DO NOT EDIT.
// Source: ./interfaces.go

// Package mock_traffic_policy_translator is a generated GoMock package.
package mock_traffic_policy_translator

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	v1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	v1alpha10 "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	types "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1/types"
)

// MockTrafficPolicyMeshTranslator is a mock of TrafficPolicyMeshTranslator interface
type MockTrafficPolicyMeshTranslator struct {
	ctrl     *gomock.Controller
	recorder *MockTrafficPolicyMeshTranslatorMockRecorder
}

// MockTrafficPolicyMeshTranslatorMockRecorder is the mock recorder for MockTrafficPolicyMeshTranslator
type MockTrafficPolicyMeshTranslatorMockRecorder struct {
	mock *MockTrafficPolicyMeshTranslator
}

// NewMockTrafficPolicyMeshTranslator creates a new mock instance
func NewMockTrafficPolicyMeshTranslator(ctrl *gomock.Controller) *MockTrafficPolicyMeshTranslator {
	mock := &MockTrafficPolicyMeshTranslator{ctrl: ctrl}
	mock.recorder = &MockTrafficPolicyMeshTranslatorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockTrafficPolicyMeshTranslator) EXPECT() *MockTrafficPolicyMeshTranslatorMockRecorder {
	return m.recorder
}

// Name mocks base method
func (m *MockTrafficPolicyMeshTranslator) Name() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Name")
	ret0, _ := ret[0].(string)
	return ret0
}

// Name indicates an expected call of Name
func (mr *MockTrafficPolicyMeshTranslatorMockRecorder) Name() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Name", reflect.TypeOf((*MockTrafficPolicyMeshTranslator)(nil).Name))
}

// TranslateTrafficPolicy mocks base method
func (m *MockTrafficPolicyMeshTranslator) TranslateTrafficPolicy(ctx context.Context, meshService *v1alpha1.MeshService, mesh *v1alpha1.Mesh, mergedTrafficPolicy []*v1alpha10.TrafficPolicy) *types.TrafficPolicyStatus_TranslatorError {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "TranslateTrafficPolicy", ctx, meshService, mesh, mergedTrafficPolicy)
	ret0, _ := ret[0].(*types.TrafficPolicyStatus_TranslatorError)
	return ret0
}

// TranslateTrafficPolicy indicates an expected call of TranslateTrafficPolicy
func (mr *MockTrafficPolicyMeshTranslatorMockRecorder) TranslateTrafficPolicy(ctx, meshService, mesh, mergedTrafficPolicy interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "TranslateTrafficPolicy", reflect.TypeOf((*MockTrafficPolicyMeshTranslator)(nil).TranslateTrafficPolicy), ctx, meshService, mesh, mergedTrafficPolicy)
}

// MockTrafficPolicyTranslatorLoop is a mock of TrafficPolicyTranslatorLoop interface
type MockTrafficPolicyTranslatorLoop struct {
	ctrl     *gomock.Controller
	recorder *MockTrafficPolicyTranslatorLoopMockRecorder
}

// MockTrafficPolicyTranslatorLoopMockRecorder is the mock recorder for MockTrafficPolicyTranslatorLoop
type MockTrafficPolicyTranslatorLoopMockRecorder struct {
	mock *MockTrafficPolicyTranslatorLoop
}

// NewMockTrafficPolicyTranslatorLoop creates a new mock instance
func NewMockTrafficPolicyTranslatorLoop(ctrl *gomock.Controller) *MockTrafficPolicyTranslatorLoop {
	mock := &MockTrafficPolicyTranslatorLoop{ctrl: ctrl}
	mock.recorder = &MockTrafficPolicyTranslatorLoopMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockTrafficPolicyTranslatorLoop) EXPECT() *MockTrafficPolicyTranslatorLoopMockRecorder {
	return m.recorder
}

// Start mocks base method
func (m *MockTrafficPolicyTranslatorLoop) Start(ctx context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Start", ctx)
	ret0, _ := ret[0].(error)
	return ret0
}

// Start indicates an expected call of Start
func (mr *MockTrafficPolicyTranslatorLoopMockRecorder) Start(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Start", reflect.TypeOf((*MockTrafficPolicyTranslatorLoop)(nil).Start), ctx)
}
