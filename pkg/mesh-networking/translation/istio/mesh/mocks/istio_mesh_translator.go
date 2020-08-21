// Code generated by MockGen. DO NOT EDIT.
// Source: ./istio_mesh_translator.go

// Package mock_mesh is a generated GoMock package.
package mock_mesh

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	v1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	input "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/istio/input"
	output "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/istio/output"
	reporting "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/reporting"
)

// MockTranslator is a mock of Translator interface
type MockTranslator struct {
	ctrl     *gomock.Controller
	recorder *MockTranslatorMockRecorder
}

// MockTranslatorMockRecorder is the mock recorder for MockTranslator
type MockTranslatorMockRecorder struct {
	mock *MockTranslator
}

// NewMockTranslator creates a new mock instance
func NewMockTranslator(ctrl *gomock.Controller) *MockTranslator {
	mock := &MockTranslator{ctrl: ctrl}
	mock.recorder = &MockTranslatorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockTranslator) EXPECT() *MockTranslatorMockRecorder {
	return m.recorder
}

// Translate mocks base method
func (m *MockTranslator) Translate(in input.Snapshot, mesh *v1alpha2.Mesh, outputs output.Builder, reporter reporting.Reporter) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Translate", in, mesh, outputs, reporter)
}

// Translate indicates an expected call of Translate
func (mr *MockTranslatorMockRecorder) Translate(in, mesh, outputs, reporter interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Translate", reflect.TypeOf((*MockTranslator)(nil).Translate), in, mesh, outputs, reporter)
}
