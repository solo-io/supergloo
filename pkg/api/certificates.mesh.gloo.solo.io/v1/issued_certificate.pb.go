// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.26.0
// 	protoc        v3.6.1
// source: github.com/solo-io/gloo-mesh/api/certificates/v1/issued_certificate.proto

package v1

import (
	_ "github.com/golang/protobuf/ptypes/duration"
	_ "github.com/golang/protobuf/ptypes/timestamp"
	_ "github.com/solo-io/protoc-gen-ext/extproto"
	v1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// Possible states in which an IssuedCertificate can exist.
type IssuedCertificateStatus_State int32

const (
	// The IssuedCertificate has yet to be picked up by the agent.
	IssuedCertificateStatus_PENDING IssuedCertificateStatus_State = 0
	// The agent has created a local private key
	// and a CertificateRequest for the IssuedCertificate.
	// In this state, the agent is waiting for the Issuer
	// to issue certificates for the CertificateRequest before proceeding.
	IssuedCertificateStatus_REQUESTED IssuedCertificateStatus_State = 1
	// The certificate has been issued. Any pods that require restarting will be restarted at this point.
	IssuedCertificateStatus_ISSUED IssuedCertificateStatus_State = 2
	// The reply from the Issuer has been processed and
	// the agent has placed the final certificate secret
	// in the target location specified by the IssuedCertificate.
	IssuedCertificateStatus_FINISHED IssuedCertificateStatus_State = 3
	// Processing the certificate workflow failed.
	IssuedCertificateStatus_FAILED IssuedCertificateStatus_State = 4
)

// Enum value maps for IssuedCertificateStatus_State.
var (
	IssuedCertificateStatus_State_name = map[int32]string{
		0: "PENDING",
		1: "REQUESTED",
		2: "ISSUED",
		3: "FINISHED",
		4: "FAILED",
	}
	IssuedCertificateStatus_State_value = map[string]int32{
		"PENDING":   0,
		"REQUESTED": 1,
		"ISSUED":    2,
		"FINISHED":  3,
		"FAILED":    4,
	}
)

func (x IssuedCertificateStatus_State) Enum() *IssuedCertificateStatus_State {
	p := new(IssuedCertificateStatus_State)
	*p = x
	return p
}

func (x IssuedCertificateStatus_State) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (IssuedCertificateStatus_State) Descriptor() protoreflect.EnumDescriptor {
	return file_github_com_solo_io_gloo_mesh_api_certificates_v1_issued_certificate_proto_enumTypes[0].Descriptor()
}

func (IssuedCertificateStatus_State) Type() protoreflect.EnumType {
	return &file_github_com_solo_io_gloo_mesh_api_certificates_v1_issued_certificate_proto_enumTypes[0]
}

func (x IssuedCertificateStatus_State) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use IssuedCertificateStatus_State.Descriptor instead.
func (IssuedCertificateStatus_State) EnumDescriptor() ([]byte, []int) {
	return file_github_com_solo_io_gloo_mesh_api_certificates_v1_issued_certificate_proto_rawDescGZIP(), []int{2, 0}
}

//
//IssuedCertificates are used to issue SSL certificates
//to remote Kubernetes clusters from a central (out-of-cluster) Certificate Authority.
//
//When an IssuedCertificate is created, a certificate is issued to a remote cluster by a central Certificate Authority via
//the following workflow:
//
//1. The Certificate Issuer creates the IssuedCertificate resource on the remote cluster
//2. The Certificate Signature Requesting Agent installed to the remote cluster generates
//a Certificate Signing Request and writes it to the status of the IssuedCertificate
//3. Finally, the Certificate Issuer generates signed a certificate for the CSR and writes
//it back as Kubernetes Secret in the remote cluster.
//
//Trust can therefore be established across clusters without requiring
//private keys to ever leave the node.
type IssuedCertificateSpec struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	//
	//A list of hostnames and IPs to generate a certificate for.
	//This can also be set to the identity running the workload,
	//e.g. a Kubernetes service account.
	//
	//Generally for an Istio CA this will take the form `spiffe://cluster.local/ns/istio-system/sa/citadel`.
	//
	//"cluster.local" may be replaced by the root of trust domain for the mesh.
	Hosts []string `protobuf:"bytes,1,rep,name=hosts,proto3" json:"hosts,omitempty"`
	// DEPRECATED: in favor of `common_cert_options.org_name`
	Org string `protobuf:"bytes,2,opt,name=org,proto3" json:"org,omitempty"`
	// DEPRECATED: in favor of `gloo_mesh_ca.signing_certificate_secret`
	// The secret containing the root SSL certificate used to sign this IssuedCertificate (located in the certificate issuer's cluster).
	SigningCertificateSecret *v1.ObjectRef `protobuf:"bytes,3,opt,name=signing_certificate_secret,json=signingCertificateSecret,proto3" json:"signing_certificate_secret,omitempty"`
	// The secret containing the SSL certificate to be generated for this IssuedCertificate (located in the Gloo Mesh agent's cluster).
	// If nil, the sidecar agent stores the signing certificate in memory. (Enterprise only)
	IssuedCertificateSecret *v1.ObjectRef `protobuf:"bytes,4,opt,name=issued_certificate_secret,json=issuedCertificateSecret,proto3" json:"issued_certificate_secret,omitempty"`
	// A reference to a PodBounceDirective specifying a list of Kubernetes pods to bounce
	// (delete and cause a restart) when the certificate is issued.
	//
	// Istio-controlled pods require restarting in order for Envoy proxies to pick up the newly issued certificate
	// due to [this issue](https://github.com/istio/istio/issues/22993).
	//
	// This will include the control plane pods as well as any Pods
	// which share a data plane with the target mesh.
	PodBounceDirective *v1.ObjectRef `protobuf:"bytes,5,opt,name=pod_bounce_directive,json=podBounceDirective,proto3" json:"pod_bounce_directive,omitempty"`
	// Set of options to configure the intermediate certificate being generated
	CertOptions *CommonCertOptions `protobuf:"bytes,6,opt,name=cert_options,json=certOptions,proto3" json:"cert_options,omitempty"`
	// The location of the certificate authority to sign this certificate
	//
	// Types that are assignable to CertificateAuthority:
	//	*IssuedCertificateSpec_GlooMeshCa
	//	*IssuedCertificateSpec_AgentCa
	CertificateAuthority isIssuedCertificateSpec_CertificateAuthority `protobuf_oneof:"certificate_authority"`
}

func (x *IssuedCertificateSpec) Reset() {
	*x = IssuedCertificateSpec{}
	if protoimpl.UnsafeEnabled {
		mi := &file_github_com_solo_io_gloo_mesh_api_certificates_v1_issued_certificate_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *IssuedCertificateSpec) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*IssuedCertificateSpec) ProtoMessage() {}

func (x *IssuedCertificateSpec) ProtoReflect() protoreflect.Message {
	mi := &file_github_com_solo_io_gloo_mesh_api_certificates_v1_issued_certificate_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use IssuedCertificateSpec.ProtoReflect.Descriptor instead.
func (*IssuedCertificateSpec) Descriptor() ([]byte, []int) {
	return file_github_com_solo_io_gloo_mesh_api_certificates_v1_issued_certificate_proto_rawDescGZIP(), []int{0}
}

func (x *IssuedCertificateSpec) GetHosts() []string {
	if x != nil {
		return x.Hosts
	}
	return nil
}

func (x *IssuedCertificateSpec) GetOrg() string {
	if x != nil {
		return x.Org
	}
	return ""
}

func (x *IssuedCertificateSpec) GetSigningCertificateSecret() *v1.ObjectRef {
	if x != nil {
		return x.SigningCertificateSecret
	}
	return nil
}

func (x *IssuedCertificateSpec) GetIssuedCertificateSecret() *v1.ObjectRef {
	if x != nil {
		return x.IssuedCertificateSecret
	}
	return nil
}

func (x *IssuedCertificateSpec) GetPodBounceDirective() *v1.ObjectRef {
	if x != nil {
		return x.PodBounceDirective
	}
	return nil
}

func (x *IssuedCertificateSpec) GetCertOptions() *CommonCertOptions {
	if x != nil {
		return x.CertOptions
	}
	return nil
}

func (m *IssuedCertificateSpec) GetCertificateAuthority() isIssuedCertificateSpec_CertificateAuthority {
	if m != nil {
		return m.CertificateAuthority
	}
	return nil
}

func (x *IssuedCertificateSpec) GetGlooMeshCa() *RootCertificateAuthority {
	if x, ok := x.GetCertificateAuthority().(*IssuedCertificateSpec_GlooMeshCa); ok {
		return x.GlooMeshCa
	}
	return nil
}

func (x *IssuedCertificateSpec) GetAgentCa() *IntermediateCertificateAuthority {
	if x, ok := x.GetCertificateAuthority().(*IssuedCertificateSpec_AgentCa); ok {
		return x.AgentCa
	}
	return nil
}

type isIssuedCertificateSpec_CertificateAuthority interface {
	isIssuedCertificateSpec_CertificateAuthority()
}

type IssuedCertificateSpec_GlooMeshCa struct {
	// Gloo Mesh CA options
	GlooMeshCa *RootCertificateAuthority `protobuf:"bytes,7,opt,name=gloo_mesh_ca,json=glooMeshCa,proto3,oneof"`
}

type IssuedCertificateSpec_AgentCa struct {
	// Agent CA options
	AgentCa *IntermediateCertificateAuthority `protobuf:"bytes,8,opt,name=agent_ca,json=agentCa,proto3,oneof"`
}

func (*IssuedCertificateSpec_GlooMeshCa) isIssuedCertificateSpec_CertificateAuthority() {}

func (*IssuedCertificateSpec_AgentCa) isIssuedCertificateSpec_CertificateAuthority() {}

// Set of options which represent the certificate authorities the management cluster can use
// to sign the intermediate certs.
type RootCertificateAuthority struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Certificate authority which gloo-mesh management will use to sign the intermediate cert
	//
	// Types that are assignable to CertificateAuthority:
	//	*RootCertificateAuthority_SigningCertificateSecret
	CertificateAuthority isRootCertificateAuthority_CertificateAuthority `protobuf_oneof:"certificate_authority"`
}

func (x *RootCertificateAuthority) Reset() {
	*x = RootCertificateAuthority{}
	if protoimpl.UnsafeEnabled {
		mi := &file_github_com_solo_io_gloo_mesh_api_certificates_v1_issued_certificate_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RootCertificateAuthority) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RootCertificateAuthority) ProtoMessage() {}

func (x *RootCertificateAuthority) ProtoReflect() protoreflect.Message {
	mi := &file_github_com_solo_io_gloo_mesh_api_certificates_v1_issued_certificate_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RootCertificateAuthority.ProtoReflect.Descriptor instead.
func (*RootCertificateAuthority) Descriptor() ([]byte, []int) {
	return file_github_com_solo_io_gloo_mesh_api_certificates_v1_issued_certificate_proto_rawDescGZIP(), []int{1}
}

func (m *RootCertificateAuthority) GetCertificateAuthority() isRootCertificateAuthority_CertificateAuthority {
	if m != nil {
		return m.CertificateAuthority
	}
	return nil
}

func (x *RootCertificateAuthority) GetSigningCertificateSecret() *v1.ObjectRef {
	if x, ok := x.GetCertificateAuthority().(*RootCertificateAuthority_SigningCertificateSecret); ok {
		return x.SigningCertificateSecret
	}
	return nil
}

type isRootCertificateAuthority_CertificateAuthority interface {
	isRootCertificateAuthority_CertificateAuthority()
}

type RootCertificateAuthority_SigningCertificateSecret struct {
	SigningCertificateSecret *v1.ObjectRef `protobuf:"bytes,1,opt,name=signing_certificate_secret,json=signingCertificateSecret,proto3,oneof"`
}

func (*RootCertificateAuthority_SigningCertificateSecret) isRootCertificateAuthority_CertificateAuthority() {
}

// The IssuedCertificate status is written by the CertificateRequesting agent.
type IssuedCertificateStatus struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The most recent generation observed in the the IssuedCertificate metadata.
	// If the `observedGeneration` does not match `metadata.generation`, the Gloo Mesh agent has not processed the most
	// recent version of this IssuedCertificate.
	ObservedGeneration int64 `protobuf:"varint,1,opt,name=observed_generation,json=observedGeneration,proto3" json:"observed_generation,omitempty"`
	// Any error observed which prevented the CertificateRequest from being processed.
	// If the error is empty, the request has been processed successfully.
	Error string `protobuf:"bytes,2,opt,name=error,proto3" json:"error,omitempty"`
	// The current state of the IssuedCertificate workflow, reported by the agent.
	State IssuedCertificateStatus_State `protobuf:"varint,3,opt,name=state,proto3,enum=certificates.mesh.gloo.solo.io.IssuedCertificateStatus_State" json:"state,omitempty"`
}

func (x *IssuedCertificateStatus) Reset() {
	*x = IssuedCertificateStatus{}
	if protoimpl.UnsafeEnabled {
		mi := &file_github_com_solo_io_gloo_mesh_api_certificates_v1_issued_certificate_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *IssuedCertificateStatus) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*IssuedCertificateStatus) ProtoMessage() {}

func (x *IssuedCertificateStatus) ProtoReflect() protoreflect.Message {
	mi := &file_github_com_solo_io_gloo_mesh_api_certificates_v1_issued_certificate_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use IssuedCertificateStatus.ProtoReflect.Descriptor instead.
func (*IssuedCertificateStatus) Descriptor() ([]byte, []int) {
	return file_github_com_solo_io_gloo_mesh_api_certificates_v1_issued_certificate_proto_rawDescGZIP(), []int{2}
}

func (x *IssuedCertificateStatus) GetObservedGeneration() int64 {
	if x != nil {
		return x.ObservedGeneration
	}
	return 0
}

func (x *IssuedCertificateStatus) GetError() string {
	if x != nil {
		return x.Error
	}
	return ""
}

func (x *IssuedCertificateStatus) GetState() IssuedCertificateStatus_State {
	if x != nil {
		return x.State
	}
	return IssuedCertificateStatus_PENDING
}

var File_github_com_solo_io_gloo_mesh_api_certificates_v1_issued_certificate_proto protoreflect.FileDescriptor

var file_github_com_solo_io_gloo_mesh_api_certificates_v1_issued_certificate_proto_rawDesc = []byte{
	0x0a, 0x49, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x73, 0x6f, 0x6c,
	0x6f, 0x2d, 0x69, 0x6f, 0x2f, 0x67, 0x6c, 0x6f, 0x6f, 0x2d, 0x6d, 0x65, 0x73, 0x68, 0x2f, 0x61,
	0x70, 0x69, 0x2f, 0x63, 0x65, 0x72, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x65, 0x73, 0x2f,
	0x76, 0x31, 0x2f, 0x69, 0x73, 0x73, 0x75, 0x65, 0x64, 0x5f, 0x63, 0x65, 0x72, 0x74, 0x69, 0x66,
	0x69, 0x63, 0x61, 0x74, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x1e, 0x63, 0x65, 0x72,
	0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x65, 0x73, 0x2e, 0x6d, 0x65, 0x73, 0x68, 0x2e, 0x67,
	0x6c, 0x6f, 0x6f, 0x2e, 0x73, 0x6f, 0x6c, 0x6f, 0x2e, 0x69, 0x6f, 0x1a, 0x41, 0x67, 0x69, 0x74,
	0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x73, 0x6f, 0x6c, 0x6f, 0x2d, 0x69, 0x6f, 0x2f,
	0x67, 0x6c, 0x6f, 0x6f, 0x2d, 0x6d, 0x65, 0x73, 0x68, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x63, 0x65,
	0x72, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x65, 0x73, 0x2f, 0x76, 0x31, 0x2f, 0x63, 0x61,
	0x5f, 0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x2e,
	0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x73, 0x6f, 0x6c, 0x6f, 0x2d,
	0x69, 0x6f, 0x2f, 0x73, 0x6b, 0x76, 0x32, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x63, 0x6f, 0x72, 0x65,
	0x2f, 0x76, 0x31, 0x2f, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1f,
	0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f,
	0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a,
	0x1e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66,
	0x2f, 0x64, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a,
	0x12, 0x65, 0x78, 0x74, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x65, 0x78, 0x74, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x22, 0xf1, 0x04, 0x0a, 0x15, 0x49, 0x73, 0x73, 0x75, 0x65, 0x64, 0x43, 0x65,
	0x72, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x65, 0x53, 0x70, 0x65, 0x63, 0x12, 0x14, 0x0a,
	0x05, 0x68, 0x6f, 0x73, 0x74, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x09, 0x52, 0x05, 0x68, 0x6f,
	0x73, 0x74, 0x73, 0x12, 0x10, 0x0a, 0x03, 0x6f, 0x72, 0x67, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x03, 0x6f, 0x72, 0x67, 0x12, 0x5a, 0x0a, 0x1a, 0x73, 0x69, 0x67, 0x6e, 0x69, 0x6e, 0x67,
	0x5f, 0x63, 0x65, 0x72, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x65, 0x5f, 0x73, 0x65, 0x63,
	0x72, 0x65, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1c, 0x2e, 0x63, 0x6f, 0x72, 0x65,
	0x2e, 0x73, 0x6b, 0x76, 0x32, 0x2e, 0x73, 0x6f, 0x6c, 0x6f, 0x2e, 0x69, 0x6f, 0x2e, 0x4f, 0x62,
	0x6a, 0x65, 0x63, 0x74, 0x52, 0x65, 0x66, 0x52, 0x18, 0x73, 0x69, 0x67, 0x6e, 0x69, 0x6e, 0x67,
	0x43, 0x65, 0x72, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x65, 0x53, 0x65, 0x63, 0x72, 0x65,
	0x74, 0x12, 0x58, 0x0a, 0x19, 0x69, 0x73, 0x73, 0x75, 0x65, 0x64, 0x5f, 0x63, 0x65, 0x72, 0x74,
	0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x65, 0x5f, 0x73, 0x65, 0x63, 0x72, 0x65, 0x74, 0x18, 0x04,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x1c, 0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x73, 0x6b, 0x76, 0x32,
	0x2e, 0x73, 0x6f, 0x6c, 0x6f, 0x2e, 0x69, 0x6f, 0x2e, 0x4f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x52,
	0x65, 0x66, 0x52, 0x17, 0x69, 0x73, 0x73, 0x75, 0x65, 0x64, 0x43, 0x65, 0x72, 0x74, 0x69, 0x66,
	0x69, 0x63, 0x61, 0x74, 0x65, 0x53, 0x65, 0x63, 0x72, 0x65, 0x74, 0x12, 0x4e, 0x0a, 0x14, 0x70,
	0x6f, 0x64, 0x5f, 0x62, 0x6f, 0x75, 0x6e, 0x63, 0x65, 0x5f, 0x64, 0x69, 0x72, 0x65, 0x63, 0x74,
	0x69, 0x76, 0x65, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1c, 0x2e, 0x63, 0x6f, 0x72, 0x65,
	0x2e, 0x73, 0x6b, 0x76, 0x32, 0x2e, 0x73, 0x6f, 0x6c, 0x6f, 0x2e, 0x69, 0x6f, 0x2e, 0x4f, 0x62,
	0x6a, 0x65, 0x63, 0x74, 0x52, 0x65, 0x66, 0x52, 0x12, 0x70, 0x6f, 0x64, 0x42, 0x6f, 0x75, 0x6e,
	0x63, 0x65, 0x44, 0x69, 0x72, 0x65, 0x63, 0x74, 0x69, 0x76, 0x65, 0x12, 0x54, 0x0a, 0x0c, 0x63,
	0x65, 0x72, 0x74, 0x5f, 0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0x06, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x31, 0x2e, 0x63, 0x65, 0x72, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x65, 0x73,
	0x2e, 0x6d, 0x65, 0x73, 0x68, 0x2e, 0x67, 0x6c, 0x6f, 0x6f, 0x2e, 0x73, 0x6f, 0x6c, 0x6f, 0x2e,
	0x69, 0x6f, 0x2e, 0x43, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x43, 0x65, 0x72, 0x74, 0x4f, 0x70, 0x74,
	0x69, 0x6f, 0x6e, 0x73, 0x52, 0x0b, 0x63, 0x65, 0x72, 0x74, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e,
	0x73, 0x12, 0x5c, 0x0a, 0x0c, 0x67, 0x6c, 0x6f, 0x6f, 0x5f, 0x6d, 0x65, 0x73, 0x68, 0x5f, 0x63,
	0x61, 0x18, 0x07, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x38, 0x2e, 0x63, 0x65, 0x72, 0x74, 0x69, 0x66,
	0x69, 0x63, 0x61, 0x74, 0x65, 0x73, 0x2e, 0x6d, 0x65, 0x73, 0x68, 0x2e, 0x67, 0x6c, 0x6f, 0x6f,
	0x2e, 0x73, 0x6f, 0x6c, 0x6f, 0x2e, 0x69, 0x6f, 0x2e, 0x52, 0x6f, 0x6f, 0x74, 0x43, 0x65, 0x72,
	0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x65, 0x41, 0x75, 0x74, 0x68, 0x6f, 0x72, 0x69, 0x74,
	0x79, 0x48, 0x00, 0x52, 0x0a, 0x67, 0x6c, 0x6f, 0x6f, 0x4d, 0x65, 0x73, 0x68, 0x43, 0x61, 0x12,
	0x5d, 0x0a, 0x08, 0x61, 0x67, 0x65, 0x6e, 0x74, 0x5f, 0x63, 0x61, 0x18, 0x08, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x40, 0x2e, 0x63, 0x65, 0x72, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x65, 0x73,
	0x2e, 0x6d, 0x65, 0x73, 0x68, 0x2e, 0x67, 0x6c, 0x6f, 0x6f, 0x2e, 0x73, 0x6f, 0x6c, 0x6f, 0x2e,
	0x69, 0x6f, 0x2e, 0x49, 0x6e, 0x74, 0x65, 0x72, 0x6d, 0x65, 0x64, 0x69, 0x61, 0x74, 0x65, 0x43,
	0x65, 0x72, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x65, 0x41, 0x75, 0x74, 0x68, 0x6f, 0x72,
	0x69, 0x74, 0x79, 0x48, 0x00, 0x52, 0x07, 0x61, 0x67, 0x65, 0x6e, 0x74, 0x43, 0x61, 0x42, 0x17,
	0x0a, 0x15, 0x63, 0x65, 0x72, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x65, 0x5f, 0x61, 0x75,
	0x74, 0x68, 0x6f, 0x72, 0x69, 0x74, 0x79, 0x22, 0x91, 0x01, 0x0a, 0x18, 0x52, 0x6f, 0x6f, 0x74,
	0x43, 0x65, 0x72, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x65, 0x41, 0x75, 0x74, 0x68, 0x6f,
	0x72, 0x69, 0x74, 0x79, 0x12, 0x5c, 0x0a, 0x1a, 0x73, 0x69, 0x67, 0x6e, 0x69, 0x6e, 0x67, 0x5f,
	0x63, 0x65, 0x72, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x65, 0x5f, 0x73, 0x65, 0x63, 0x72,
	0x65, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1c, 0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e,
	0x73, 0x6b, 0x76, 0x32, 0x2e, 0x73, 0x6f, 0x6c, 0x6f, 0x2e, 0x69, 0x6f, 0x2e, 0x4f, 0x62, 0x6a,
	0x65, 0x63, 0x74, 0x52, 0x65, 0x66, 0x48, 0x00, 0x52, 0x18, 0x73, 0x69, 0x67, 0x6e, 0x69, 0x6e,
	0x67, 0x43, 0x65, 0x72, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x65, 0x53, 0x65, 0x63, 0x72,
	0x65, 0x74, 0x42, 0x17, 0x0a, 0x15, 0x63, 0x65, 0x72, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74,
	0x65, 0x5f, 0x61, 0x75, 0x74, 0x68, 0x6f, 0x72, 0x69, 0x74, 0x79, 0x22, 0x80, 0x02, 0x0a, 0x17,
	0x49, 0x73, 0x73, 0x75, 0x65, 0x64, 0x43, 0x65, 0x72, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74,
	0x65, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x2f, 0x0a, 0x13, 0x6f, 0x62, 0x73, 0x65, 0x72,
	0x76, 0x65, 0x64, 0x5f, 0x67, 0x65, 0x6e, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x03, 0x52, 0x12, 0x6f, 0x62, 0x73, 0x65, 0x72, 0x76, 0x65, 0x64, 0x47, 0x65,
	0x6e, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x14, 0x0a, 0x05, 0x65, 0x72, 0x72, 0x6f,
	0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x12, 0x53,
	0x0a, 0x05, 0x73, 0x74, 0x61, 0x74, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x3d, 0x2e,
	0x63, 0x65, 0x72, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x65, 0x73, 0x2e, 0x6d, 0x65, 0x73,
	0x68, 0x2e, 0x67, 0x6c, 0x6f, 0x6f, 0x2e, 0x73, 0x6f, 0x6c, 0x6f, 0x2e, 0x69, 0x6f, 0x2e, 0x49,
	0x73, 0x73, 0x75, 0x65, 0x64, 0x43, 0x65, 0x72, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x65,
	0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x2e, 0x53, 0x74, 0x61, 0x74, 0x65, 0x52, 0x05, 0x73, 0x74,
	0x61, 0x74, 0x65, 0x22, 0x49, 0x0a, 0x05, 0x53, 0x74, 0x61, 0x74, 0x65, 0x12, 0x0b, 0x0a, 0x07,
	0x50, 0x45, 0x4e, 0x44, 0x49, 0x4e, 0x47, 0x10, 0x00, 0x12, 0x0d, 0x0a, 0x09, 0x52, 0x45, 0x51,
	0x55, 0x45, 0x53, 0x54, 0x45, 0x44, 0x10, 0x01, 0x12, 0x0a, 0x0a, 0x06, 0x49, 0x53, 0x53, 0x55,
	0x45, 0x44, 0x10, 0x02, 0x12, 0x0c, 0x0a, 0x08, 0x46, 0x49, 0x4e, 0x49, 0x53, 0x48, 0x45, 0x44,
	0x10, 0x03, 0x12, 0x0a, 0x0a, 0x06, 0x46, 0x41, 0x49, 0x4c, 0x45, 0x44, 0x10, 0x04, 0x42, 0x4c,
	0x5a, 0x46, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x73, 0x6f, 0x6c,
	0x6f, 0x2d, 0x69, 0x6f, 0x2f, 0x67, 0x6c, 0x6f, 0x6f, 0x2d, 0x6d, 0x65, 0x73, 0x68, 0x2f, 0x70,
	0x6b, 0x67, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x63, 0x65, 0x72, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61,
	0x74, 0x65, 0x73, 0x2e, 0x6d, 0x65, 0x73, 0x68, 0x2e, 0x67, 0x6c, 0x6f, 0x6f, 0x2e, 0x73, 0x6f,
	0x6c, 0x6f, 0x2e, 0x69, 0x6f, 0x2f, 0x76, 0x31, 0xc0, 0xf5, 0x04, 0x01, 0x62, 0x06, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_github_com_solo_io_gloo_mesh_api_certificates_v1_issued_certificate_proto_rawDescOnce sync.Once
	file_github_com_solo_io_gloo_mesh_api_certificates_v1_issued_certificate_proto_rawDescData = file_github_com_solo_io_gloo_mesh_api_certificates_v1_issued_certificate_proto_rawDesc
)

func file_github_com_solo_io_gloo_mesh_api_certificates_v1_issued_certificate_proto_rawDescGZIP() []byte {
	file_github_com_solo_io_gloo_mesh_api_certificates_v1_issued_certificate_proto_rawDescOnce.Do(func() {
		file_github_com_solo_io_gloo_mesh_api_certificates_v1_issued_certificate_proto_rawDescData = protoimpl.X.CompressGZIP(file_github_com_solo_io_gloo_mesh_api_certificates_v1_issued_certificate_proto_rawDescData)
	})
	return file_github_com_solo_io_gloo_mesh_api_certificates_v1_issued_certificate_proto_rawDescData
}

var file_github_com_solo_io_gloo_mesh_api_certificates_v1_issued_certificate_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_github_com_solo_io_gloo_mesh_api_certificates_v1_issued_certificate_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_github_com_solo_io_gloo_mesh_api_certificates_v1_issued_certificate_proto_goTypes = []interface{}{
	(IssuedCertificateStatus_State)(0),       // 0: certificates.mesh.gloo.solo.io.IssuedCertificateStatus.State
	(*IssuedCertificateSpec)(nil),            // 1: certificates.mesh.gloo.solo.io.IssuedCertificateSpec
	(*RootCertificateAuthority)(nil),         // 2: certificates.mesh.gloo.solo.io.RootCertificateAuthority
	(*IssuedCertificateStatus)(nil),          // 3: certificates.mesh.gloo.solo.io.IssuedCertificateStatus
	(*v1.ObjectRef)(nil),                     // 4: core.skv2.solo.io.ObjectRef
	(*CommonCertOptions)(nil),                // 5: certificates.mesh.gloo.solo.io.CommonCertOptions
	(*IntermediateCertificateAuthority)(nil), // 6: certificates.mesh.gloo.solo.io.IntermediateCertificateAuthority
}
var file_github_com_solo_io_gloo_mesh_api_certificates_v1_issued_certificate_proto_depIdxs = []int32{
	4, // 0: certificates.mesh.gloo.solo.io.IssuedCertificateSpec.signing_certificate_secret:type_name -> core.skv2.solo.io.ObjectRef
	4, // 1: certificates.mesh.gloo.solo.io.IssuedCertificateSpec.issued_certificate_secret:type_name -> core.skv2.solo.io.ObjectRef
	4, // 2: certificates.mesh.gloo.solo.io.IssuedCertificateSpec.pod_bounce_directive:type_name -> core.skv2.solo.io.ObjectRef
	5, // 3: certificates.mesh.gloo.solo.io.IssuedCertificateSpec.cert_options:type_name -> certificates.mesh.gloo.solo.io.CommonCertOptions
	2, // 4: certificates.mesh.gloo.solo.io.IssuedCertificateSpec.gloo_mesh_ca:type_name -> certificates.mesh.gloo.solo.io.RootCertificateAuthority
	6, // 5: certificates.mesh.gloo.solo.io.IssuedCertificateSpec.agent_ca:type_name -> certificates.mesh.gloo.solo.io.IntermediateCertificateAuthority
	4, // 6: certificates.mesh.gloo.solo.io.RootCertificateAuthority.signing_certificate_secret:type_name -> core.skv2.solo.io.ObjectRef
	0, // 7: certificates.mesh.gloo.solo.io.IssuedCertificateStatus.state:type_name -> certificates.mesh.gloo.solo.io.IssuedCertificateStatus.State
	8, // [8:8] is the sub-list for method output_type
	8, // [8:8] is the sub-list for method input_type
	8, // [8:8] is the sub-list for extension type_name
	8, // [8:8] is the sub-list for extension extendee
	0, // [0:8] is the sub-list for field type_name
}

func init() { file_github_com_solo_io_gloo_mesh_api_certificates_v1_issued_certificate_proto_init() }
func file_github_com_solo_io_gloo_mesh_api_certificates_v1_issued_certificate_proto_init() {
	if File_github_com_solo_io_gloo_mesh_api_certificates_v1_issued_certificate_proto != nil {
		return
	}
	file_github_com_solo_io_gloo_mesh_api_certificates_v1_ca_options_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_github_com_solo_io_gloo_mesh_api_certificates_v1_issued_certificate_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*IssuedCertificateSpec); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_github_com_solo_io_gloo_mesh_api_certificates_v1_issued_certificate_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RootCertificateAuthority); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_github_com_solo_io_gloo_mesh_api_certificates_v1_issued_certificate_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*IssuedCertificateStatus); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	file_github_com_solo_io_gloo_mesh_api_certificates_v1_issued_certificate_proto_msgTypes[0].OneofWrappers = []interface{}{
		(*IssuedCertificateSpec_GlooMeshCa)(nil),
		(*IssuedCertificateSpec_AgentCa)(nil),
	}
	file_github_com_solo_io_gloo_mesh_api_certificates_v1_issued_certificate_proto_msgTypes[1].OneofWrappers = []interface{}{
		(*RootCertificateAuthority_SigningCertificateSecret)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_github_com_solo_io_gloo_mesh_api_certificates_v1_issued_certificate_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_github_com_solo_io_gloo_mesh_api_certificates_v1_issued_certificate_proto_goTypes,
		DependencyIndexes: file_github_com_solo_io_gloo_mesh_api_certificates_v1_issued_certificate_proto_depIdxs,
		EnumInfos:         file_github_com_solo_io_gloo_mesh_api_certificates_v1_issued_certificate_proto_enumTypes,
		MessageInfos:      file_github_com_solo_io_gloo_mesh_api_certificates_v1_issued_certificate_proto_msgTypes,
	}.Build()
	File_github_com_solo_io_gloo_mesh_api_certificates_v1_issued_certificate_proto = out.File
	file_github_com_solo_io_gloo_mesh_api_certificates_v1_issued_certificate_proto_rawDesc = nil
	file_github_com_solo_io_gloo_mesh_api_certificates_v1_issued_certificate_proto_goTypes = nil
	file_github_com_solo_io_gloo_mesh_api_certificates_v1_issued_certificate_proto_depIdxs = nil
}
