// Code generated by MockGen. DO NOT EDIT.
// Source: ./translation_processor.go

// Package mock_translation_framework is a generated GoMock package.
package mock_translation_framework

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	v1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	snapshot "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/traffic-policy-temp/translation/framework/snapshot"
)

// MockTranslationProcessor is a mock of TranslationProcessor interface.
type MockTranslationProcessor struct {
	ctrl     *gomock.Controller
	recorder *MockTranslationProcessorMockRecorder
}

// MockTranslationProcessorMockRecorder is the mock recorder for MockTranslationProcessor.
type MockTranslationProcessorMockRecorder struct {
	mock *MockTranslationProcessor
}

// NewMockTranslationProcessor creates a new mock instance.
func NewMockTranslationProcessor(ctrl *gomock.Controller) *MockTranslationProcessor {
	mock := &MockTranslationProcessor{ctrl: ctrl}
	mock.recorder = &MockTranslationProcessorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockTranslationProcessor) EXPECT() *MockTranslationProcessorMockRecorder {
	return m.recorder
}

// Process mocks base method.
func (m *MockTranslationProcessor) Process(ctx context.Context, allMeshServices []*v1alpha1.MeshService) (snapshot.ClusterNameToSnapshot, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Process", ctx, allMeshServices)
	ret0, _ := ret[0].(snapshot.ClusterNameToSnapshot)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Process indicates an expected call of Process.
func (mr *MockTranslationProcessorMockRecorder) Process(ctx, allMeshServices interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Process", reflect.TypeOf((*MockTranslationProcessor)(nil).Process), ctx, allMeshServices)
}
