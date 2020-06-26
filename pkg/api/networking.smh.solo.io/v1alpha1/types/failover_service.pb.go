// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: github.com/solo-io/service-mesh-hub/api/networking/v1alpha1/failover_service.proto

package types

import (
	fmt "fmt"
	math "math"

	proto "github.com/gogo/protobuf/proto"
	types "github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1/types"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

//
//This configures an existing service with failover functionality, where in the case of
//an unhealthy service, requests will be shifted over to other services in priority
//order defined in the list of failover services,
//i.e. an unhealthy target_service will cause failover to workloads[0], etc.
//
//Currently this feature only supports Services backed by Istio.
type FailoverServiceSpec struct {
	// The service for which to add failover functionality.
	TargetService *types.ResourceRef `protobuf:"bytes,1,opt,name=target_service,json=targetService,proto3" json:"target_service,omitempty"`
	//
	//A list of services ordered by decreasing priority for failover.
	//All services must be controlled by service meshes that are grouped under a common VirtualMesh.
	FailoverServices     []*types.ResourceRef `protobuf:"bytes,2,rep,name=failover_services,json=failoverServices,proto3" json:"failover_services,omitempty"`
	XXX_NoUnkeyedLiteral struct{}             `json:"-"`
	XXX_unrecognized     []byte               `json:"-"`
	XXX_sizecache        int32                `json:"-"`
}

func (m *FailoverServiceSpec) Reset()         { *m = FailoverServiceSpec{} }
func (m *FailoverServiceSpec) String() string { return proto.CompactTextString(m) }
func (*FailoverServiceSpec) ProtoMessage()    {}
func (*FailoverServiceSpec) Descriptor() ([]byte, []int) {
	return fileDescriptor_05f0dd5940b7bfb1, []int{0}
}
func (m *FailoverServiceSpec) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_FailoverServiceSpec.Unmarshal(m, b)
}
func (m *FailoverServiceSpec) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_FailoverServiceSpec.Marshal(b, m, deterministic)
}
func (m *FailoverServiceSpec) XXX_Merge(src proto.Message) {
	xxx_messageInfo_FailoverServiceSpec.Merge(m, src)
}
func (m *FailoverServiceSpec) XXX_Size() int {
	return xxx_messageInfo_FailoverServiceSpec.Size(m)
}
func (m *FailoverServiceSpec) XXX_DiscardUnknown() {
	xxx_messageInfo_FailoverServiceSpec.DiscardUnknown(m)
}

var xxx_messageInfo_FailoverServiceSpec proto.InternalMessageInfo

func (m *FailoverServiceSpec) GetTargetService() *types.ResourceRef {
	if m != nil {
		return m.TargetService
	}
	return nil
}

func (m *FailoverServiceSpec) GetFailoverServices() []*types.ResourceRef {
	if m != nil {
		return m.FailoverServices
	}
	return nil
}

type FailoverServiceStatus struct {
	// The generation the validation_status was observed on.
	ObservedGeneration int64 `protobuf:"varint,1,opt,name=observed_generation,json=observedGeneration,proto3" json:"observed_generation,omitempty"`
	// Whether or not the resource has been successfully translated into concrete, mesh-specific routing configuration.
	TranslationStatus *types.Status `protobuf:"bytes,2,opt,name=translation_status,json=translationStatus,proto3" json:"translation_status,omitempty"`
	// Provides details on any translation errors that occurred. If any errors exist, this FailoverService has not been translated into mesh-specific config.
	TranslatorErrors []*FailoverServiceStatus_TranslatorError `protobuf:"bytes,3,rep,name=translator_errors,json=translatorErrors,proto3" json:"translator_errors,omitempty"`
	// Whether or not this resource has passed validation. This is a required step before it can be translated into concrete, mesh-specific failover configuration.
	ValidationStatus     *types.Status `protobuf:"bytes,4,opt,name=validation_status,json=validationStatus,proto3" json:"validation_status,omitempty"`
	XXX_NoUnkeyedLiteral struct{}      `json:"-"`
	XXX_unrecognized     []byte        `json:"-"`
	XXX_sizecache        int32         `json:"-"`
}

func (m *FailoverServiceStatus) Reset()         { *m = FailoverServiceStatus{} }
func (m *FailoverServiceStatus) String() string { return proto.CompactTextString(m) }
func (*FailoverServiceStatus) ProtoMessage()    {}
func (*FailoverServiceStatus) Descriptor() ([]byte, []int) {
	return fileDescriptor_05f0dd5940b7bfb1, []int{1}
}
func (m *FailoverServiceStatus) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_FailoverServiceStatus.Unmarshal(m, b)
}
func (m *FailoverServiceStatus) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_FailoverServiceStatus.Marshal(b, m, deterministic)
}
func (m *FailoverServiceStatus) XXX_Merge(src proto.Message) {
	xxx_messageInfo_FailoverServiceStatus.Merge(m, src)
}
func (m *FailoverServiceStatus) XXX_Size() int {
	return xxx_messageInfo_FailoverServiceStatus.Size(m)
}
func (m *FailoverServiceStatus) XXX_DiscardUnknown() {
	xxx_messageInfo_FailoverServiceStatus.DiscardUnknown(m)
}

var xxx_messageInfo_FailoverServiceStatus proto.InternalMessageInfo

func (m *FailoverServiceStatus) GetObservedGeneration() int64 {
	if m != nil {
		return m.ObservedGeneration
	}
	return 0
}

func (m *FailoverServiceStatus) GetTranslationStatus() *types.Status {
	if m != nil {
		return m.TranslationStatus
	}
	return nil
}

func (m *FailoverServiceStatus) GetTranslatorErrors() []*FailoverServiceStatus_TranslatorError {
	if m != nil {
		return m.TranslatorErrors
	}
	return nil
}

func (m *FailoverServiceStatus) GetValidationStatus() *types.Status {
	if m != nil {
		return m.ValidationStatus
	}
	return nil
}

type FailoverServiceStatus_TranslatorError struct {
	// ID representing a translator that translates FailoverService to Mesh-specific config.
	TranslatorId         string   `protobuf:"bytes,1,opt,name=translator_id,json=translatorId,proto3" json:"translator_id,omitempty"`
	ErrorMessage         string   `protobuf:"bytes,2,opt,name=error_message,json=errorMessage,proto3" json:"error_message,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *FailoverServiceStatus_TranslatorError) Reset()         { *m = FailoverServiceStatus_TranslatorError{} }
func (m *FailoverServiceStatus_TranslatorError) String() string { return proto.CompactTextString(m) }
func (*FailoverServiceStatus_TranslatorError) ProtoMessage()    {}
func (*FailoverServiceStatus_TranslatorError) Descriptor() ([]byte, []int) {
	return fileDescriptor_05f0dd5940b7bfb1, []int{1, 0}
}
func (m *FailoverServiceStatus_TranslatorError) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_FailoverServiceStatus_TranslatorError.Unmarshal(m, b)
}
func (m *FailoverServiceStatus_TranslatorError) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_FailoverServiceStatus_TranslatorError.Marshal(b, m, deterministic)
}
func (m *FailoverServiceStatus_TranslatorError) XXX_Merge(src proto.Message) {
	xxx_messageInfo_FailoverServiceStatus_TranslatorError.Merge(m, src)
}
func (m *FailoverServiceStatus_TranslatorError) XXX_Size() int {
	return xxx_messageInfo_FailoverServiceStatus_TranslatorError.Size(m)
}
func (m *FailoverServiceStatus_TranslatorError) XXX_DiscardUnknown() {
	xxx_messageInfo_FailoverServiceStatus_TranslatorError.DiscardUnknown(m)
}

var xxx_messageInfo_FailoverServiceStatus_TranslatorError proto.InternalMessageInfo

func (m *FailoverServiceStatus_TranslatorError) GetTranslatorId() string {
	if m != nil {
		return m.TranslatorId
	}
	return ""
}

func (m *FailoverServiceStatus_TranslatorError) GetErrorMessage() string {
	if m != nil {
		return m.ErrorMessage
	}
	return ""
}

func init() {
	proto.RegisterType((*FailoverServiceSpec)(nil), "networking.smh.solo.io.FailoverServiceSpec")
	proto.RegisterType((*FailoverServiceStatus)(nil), "networking.smh.solo.io.FailoverServiceStatus")
	proto.RegisterType((*FailoverServiceStatus_TranslatorError)(nil), "networking.smh.solo.io.FailoverServiceStatus.TranslatorError")
}

func init() {
	proto.RegisterFile("github.com/solo-io/service-mesh-hub/api/networking/v1alpha1/failover_service.proto", fileDescriptor_05f0dd5940b7bfb1)
}

var fileDescriptor_05f0dd5940b7bfb1 = []byte{
	// 401 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x93, 0x31, 0x8f, 0xd3, 0x30,
	0x1c, 0xc5, 0xd5, 0x0b, 0x42, 0x3a, 0x73, 0x07, 0xad, 0x4f, 0xa0, 0xa8, 0x12, 0xd2, 0xa9, 0x2c,
	0x0c, 0xd4, 0x56, 0x61, 0x66, 0x41, 0x94, 0x0a, 0x24, 0x06, 0x5c, 0x26, 0x18, 0x22, 0x27, 0xf9,
	0x27, 0x31, 0x4d, 0xe2, 0xc8, 0x76, 0x82, 0xf8, 0x38, 0x4c, 0x7c, 0x4d, 0x14, 0x3b, 0x6d, 0xd2,
	0x5c, 0xa5, 0x76, 0xcc, 0xb3, 0xdf, 0xf3, 0xef, 0xfd, 0x63, 0x23, 0x96, 0x0a, 0x93, 0xd5, 0x21,
	0x89, 0x64, 0x41, 0xb5, 0xcc, 0xe5, 0x52, 0x48, 0xaa, 0x41, 0x35, 0x22, 0x82, 0x65, 0x01, 0x3a,
	0x5b, 0x66, 0x75, 0x48, 0x79, 0x25, 0x68, 0x09, 0xe6, 0xb7, 0x54, 0x3b, 0x51, 0xa6, 0xb4, 0x59,
	0xf1, 0xbc, 0xca, 0xf8, 0x8a, 0x26, 0x5c, 0xe4, 0xb2, 0x01, 0x15, 0x74, 0x0e, 0x52, 0x29, 0x69,
	0x24, 0x7e, 0xd1, 0xef, 0x25, 0xba, 0xc8, 0x48, 0x9b, 0x4b, 0x84, 0x9c, 0xbf, 0x39, 0x19, 0x1c,
	0x49, 0x05, 0x7d, 0xa4, 0x82, 0xc4, 0xa5, 0xcc, 0xe9, 0x05, 0xbb, 0xb5, 0xe1, 0xa6, 0xd6, 0xce,
	0xb0, 0xf8, 0x37, 0x41, 0x77, 0x9f, 0x3a, 0xa2, 0xad, 0xf3, 0x6e, 0x2b, 0x88, 0xf0, 0x47, 0xf4,
	0xd4, 0x70, 0x95, 0x82, 0xd9, 0x63, 0xfa, 0x93, 0xfb, 0xc9, 0xeb, 0x27, 0x6f, 0x5f, 0x92, 0x36,
	0x6c, 0x48, 0x48, 0x18, 0x68, 0x59, 0xab, 0x08, 0x18, 0x24, 0xec, 0xd6, 0x99, 0xba, 0x24, 0xfc,
	0x05, 0xcd, 0xc6, 0x75, 0xb5, 0x7f, 0x75, 0xef, 0x9d, 0x0f, 0x9a, 0x26, 0xc7, 0x50, 0x7a, 0xf1,
	0xd7, 0x43, 0xcf, 0xc7, 0xa4, 0xb6, 0x09, 0xa6, 0xe8, 0x4e, 0x86, 0x6d, 0x3c, 0xc4, 0x41, 0x0a,
	0x25, 0x28, 0x6e, 0x84, 0x2c, 0x2d, 0xb0, 0xc7, 0xf0, 0x7e, 0x69, 0x73, 0x58, 0xc1, 0x1b, 0x84,
	0x8d, 0xe2, 0xa5, 0xce, 0xed, 0x67, 0xe0, 0x06, 0xe2, 0x5f, 0xd9, 0x82, 0xfe, 0x43, 0x2e, 0x77,
	0x0c, 0x9b, 0x0d, 0x3c, 0xdd, 0xc9, 0xbf, 0xd0, 0x41, 0x94, 0x2a, 0x00, 0xa5, 0xa4, 0xd2, 0xbe,
	0x67, 0xfb, 0xbd, 0x27, 0xa7, 0x7f, 0x28, 0x39, 0xd9, 0x81, 0x7c, 0x3f, 0xc4, 0xac, 0xdb, 0x14,
	0x36, 0x35, 0xc7, 0x82, 0xc6, 0x6b, 0x34, 0x6b, 0x78, 0x2e, 0xe2, 0x23, 0xe6, 0x47, 0x67, 0x98,
	0xa7, 0xbd, 0xc5, 0x29, 0xf3, 0x9f, 0xe8, 0xd9, 0xe8, 0x2c, 0xfc, 0x0a, 0xdd, 0x0e, 0x5a, 0x88,
	0xd8, 0x4e, 0xee, 0x9a, 0xdd, 0xf4, 0xe2, 0xe7, 0xb8, 0xdd, 0x64, 0xfb, 0x05, 0x05, 0x68, 0xcd,
	0x53, 0xb0, 0xe3, 0xba, 0x66, 0x37, 0x56, 0xfc, 0xea, 0xb4, 0x0f, 0xdb, 0x1f, 0xdf, 0x2e, 0x79,
	0x1a, 0xd5, 0x2e, 0x1d, 0x3d, 0x8f, 0x21, 0x7b, 0x7f, 0x53, 0xcd, 0x9f, 0x0a, 0x74, 0xf8, 0xd8,
	0xde, 0xd4, 0x77, 0xff, 0x03, 0x00, 0x00, 0xff, 0xff, 0x08, 0x97, 0x58, 0xae, 0x76, 0x03, 0x00,
	0x00,
}
