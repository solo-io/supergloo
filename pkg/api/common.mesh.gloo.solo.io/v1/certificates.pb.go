// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.6.1
// source: github.com/solo-io/gloo-mesh/api/common/v1/certificates.proto

package v1

import (
	reflect "reflect"
	sync "sync"

	proto "github.com/golang/protobuf/proto"
	_ "github.com/solo-io/protoc-gen-ext/extproto"
	v1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// This is a compile-time assertion that a sufficiently up-to-date version
// of the legacy proto package is being used.
const _ = proto.ProtoPackageIsVersion4

type VaultCA struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// ca_path is the mount path of the Vault PKI backend's `sign` endpoint, e.g:
	// "my_pki_mount/sign/my-role-name".
	CaPath string `protobuf:"bytes,1,opt,name=ca_path,json=caPath,proto3" json:"ca_path,omitempty"`
	// ca_path is the mount path of the Vault PKI backend's `generate` endpoint, e.g:
	// "my_pki_mount/intermediate/generate/exported".
	// exported is necessary here as istio needs access to the private key
	CsrPath string `protobuf:"bytes,2,opt,name=csr_path,json=csrPath,proto3" json:"csr_path,omitempty"`
	// Server is the connection address for the Vault server, e.g: "https://vault.example.com:8200".
	Server string `protobuf:"bytes,3,opt,name=server,proto3" json:"server,omitempty"`
	// PEM encoded CA bundle used to validate Vault server certificate. Only used
	// if the Server URL is using HTTPS protocol. This parameter is ignored for
	// plain HTTP protocol connection. If not set the system root certificates
	// are used to validate the TLS connection.
	CaBundle []byte `protobuf:"bytes,4,opt,name=CaBundle,proto3" json:"CaBundle,omitempty"`
	// Name of the vault namespace. Namespaces is a set of features within Vault Enterprise that allows Vault environments to support Secure Multi-tenancy. e.g: "ns1"
	// More about namespaces can be found here https://www.vaultproject.io/docs/enterprise/namespaces
	Namespace string `protobuf:"bytes,5,opt,name=namespace,proto3" json:"namespace,omitempty"`
	// CommonCertOptions cert_options = 5;
	//
	// Types that are assignable to AuthType:
	//	*VaultCA_TokenSecretRef
	//	*VaultCA_KubernetesAuth
	AuthType isVaultCA_AuthType `protobuf_oneof:"auth_type"`
}

func (x *VaultCA) Reset() {
	*x = VaultCA{}
	if protoimpl.UnsafeEnabled {
		mi := &file_github_com_solo_io_gloo_mesh_api_common_v1_certificates_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *VaultCA) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*VaultCA) ProtoMessage() {}

func (x *VaultCA) ProtoReflect() protoreflect.Message {
	mi := &file_github_com_solo_io_gloo_mesh_api_common_v1_certificates_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use VaultCA.ProtoReflect.Descriptor instead.
func (*VaultCA) Descriptor() ([]byte, []int) {
	return file_github_com_solo_io_gloo_mesh_api_common_v1_certificates_proto_rawDescGZIP(), []int{0}
}

func (x *VaultCA) GetCaPath() string {
	if x != nil {
		return x.CaPath
	}
	return ""
}

func (x *VaultCA) GetCsrPath() string {
	if x != nil {
		return x.CsrPath
	}
	return ""
}

func (x *VaultCA) GetServer() string {
	if x != nil {
		return x.Server
	}
	return ""
}

func (x *VaultCA) GetCaBundle() []byte {
	if x != nil {
		return x.CaBundle
	}
	return nil
}

func (x *VaultCA) GetNamespace() string {
	if x != nil {
		return x.Namespace
	}
	return ""
}

func (m *VaultCA) GetAuthType() isVaultCA_AuthType {
	if m != nil {
		return m.AuthType
	}
	return nil
}

func (x *VaultCA) GetTokenSecretRef() *v1.ObjectRef {
	if x, ok := x.GetAuthType().(*VaultCA_TokenSecretRef); ok {
		return x.TokenSecretRef
	}
	return nil
}

func (x *VaultCA) GetKubernetesAuth() *VaultCA_Kubernetes {
	if x, ok := x.GetAuthType().(*VaultCA_KubernetesAuth); ok {
		return x.KubernetesAuth
	}
	return nil
}

type isVaultCA_AuthType interface {
	isVaultCA_AuthType()
}

type VaultCA_TokenSecretRef struct {
	// TokenSecretRef authenticates with Vault by presenting a token.
	TokenSecretRef *v1.ObjectRef `protobuf:"bytes,6,opt,name=token_secret_ref,json=tokenSecretRef,proto3,oneof"`
}

type VaultCA_KubernetesAuth struct {
	// Kubernetes authenticates with Vault by passing the ServiceAccount
	// token stored in the named Secret resource to the Vault server.
	KubernetesAuth *VaultCA_Kubernetes `protobuf:"bytes,8,opt,name=kubernetes_auth,json=kubernetesAuth,proto3,oneof"`
}

func (*VaultCA_TokenSecretRef) isVaultCA_AuthType() {}

func (*VaultCA_KubernetesAuth) isVaultCA_AuthType() {}

// Configuration for generating a self-signed root certificate.
// Uses the X.509 format, RFC5280.
type CommonCertOptions struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Number of days before root cert expires. Defaults to 365.
	TtlDays uint32 `protobuf:"varint,1,opt,name=ttl_days,json=ttlDays,proto3" json:"ttl_days,omitempty"`
	// Size in bytes of the root cert's private key. Defaults to 4096.
	RsaKeySizeBytes uint32 `protobuf:"varint,2,opt,name=rsa_key_size_bytes,json=rsaKeySizeBytes,proto3" json:"rsa_key_size_bytes,omitempty"`
	// Root cert organization name. Defaults to "gloo-mesh".
	OrgName string `protobuf:"bytes,3,opt,name=org_name,json=orgName,proto3" json:"org_name,omitempty"`
}

func (x *CommonCertOptions) Reset() {
	*x = CommonCertOptions{}
	if protoimpl.UnsafeEnabled {
		mi := &file_github_com_solo_io_gloo_mesh_api_common_v1_certificates_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CommonCertOptions) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CommonCertOptions) ProtoMessage() {}

func (x *CommonCertOptions) ProtoReflect() protoreflect.Message {
	mi := &file_github_com_solo_io_gloo_mesh_api_common_v1_certificates_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CommonCertOptions.ProtoReflect.Descriptor instead.
func (*CommonCertOptions) Descriptor() ([]byte, []int) {
	return file_github_com_solo_io_gloo_mesh_api_common_v1_certificates_proto_rawDescGZIP(), []int{1}
}

func (x *CommonCertOptions) GetTtlDays() uint32 {
	if x != nil {
		return x.TtlDays
	}
	return 0
}

func (x *CommonCertOptions) GetRsaKeySizeBytes() uint32 {
	if x != nil {
		return x.RsaKeySizeBytes
	}
	return 0
}

func (x *CommonCertOptions) GetOrgName() string {
	if x != nil {
		return x.OrgName
	}
	return ""
}

type VaultCA_Kubernetes struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The Vault mountPath here is the mount path to use when authenticating with
	// Vault. For example, setting a value to `/v1/auth/foo`, will use the path
	// `/v1/auth/foo/login` to authenticate with Vault. If unspecified, the
	// default value "/v1/auth/kubernetes" will be used.
	Path string `protobuf:"bytes,1,opt,name=path,proto3" json:"path,omitempty"`
	// Reference to a service account
	SaRef *v1.ObjectRef `protobuf:"bytes,2,opt,name=sa_ref,json=saRef,proto3" json:"sa_ref,omitempty"`
	// Key in the token to search for the sa_token
	// Default to "token"
	SecretTokenKey string `protobuf:"bytes,3,opt,name=secret_token_key,json=secretTokenKey,proto3" json:"secret_token_key,omitempty"`
	// A required field containing the Vault Role to assume. A Role binds a
	// Kubernetes ServiceAccount with a set of Vault policies.
	Role string `protobuf:"bytes,4,opt,name=role,proto3" json:"role,omitempty"`
}

func (x *VaultCA_Kubernetes) Reset() {
	*x = VaultCA_Kubernetes{}
	if protoimpl.UnsafeEnabled {
		mi := &file_github_com_solo_io_gloo_mesh_api_common_v1_certificates_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *VaultCA_Kubernetes) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*VaultCA_Kubernetes) ProtoMessage() {}

func (x *VaultCA_Kubernetes) ProtoReflect() protoreflect.Message {
	mi := &file_github_com_solo_io_gloo_mesh_api_common_v1_certificates_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use VaultCA_Kubernetes.ProtoReflect.Descriptor instead.
func (*VaultCA_Kubernetes) Descriptor() ([]byte, []int) {
	return file_github_com_solo_io_gloo_mesh_api_common_v1_certificates_proto_rawDescGZIP(), []int{0, 0}
}

func (x *VaultCA_Kubernetes) GetPath() string {
	if x != nil {
		return x.Path
	}
	return ""
}

func (x *VaultCA_Kubernetes) GetSaRef() *v1.ObjectRef {
	if x != nil {
		return x.SaRef
	}
	return nil
}

func (x *VaultCA_Kubernetes) GetSecretTokenKey() string {
	if x != nil {
		return x.SecretTokenKey
	}
	return ""
}

func (x *VaultCA_Kubernetes) GetRole() string {
	if x != nil {
		return x.Role
	}
	return ""
}

var File_github_com_solo_io_gloo_mesh_api_common_v1_certificates_proto protoreflect.FileDescriptor

var file_github_com_solo_io_gloo_mesh_api_common_v1_certificates_proto_rawDesc = []byte{
	0x0a, 0x3d, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x73, 0x6f, 0x6c,
	0x6f, 0x2d, 0x69, 0x6f, 0x2f, 0x67, 0x6c, 0x6f, 0x6f, 0x2d, 0x6d, 0x65, 0x73, 0x68, 0x2f, 0x61,
	0x70, 0x69, 0x2f, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2f, 0x76, 0x31, 0x2f, 0x63, 0x65, 0x72,
	0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x65, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12,
	0x18, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x6d, 0x65, 0x73, 0x68, 0x2e, 0x67, 0x6c, 0x6f,
	0x6f, 0x2e, 0x73, 0x6f, 0x6c, 0x6f, 0x2e, 0x69, 0x6f, 0x1a, 0x2e, 0x67, 0x69, 0x74, 0x68, 0x75,
	0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x73, 0x6f, 0x6c, 0x6f, 0x2d, 0x69, 0x6f, 0x2f, 0x73, 0x6b,
	0x76, 0x32, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x63, 0x6f, 0x72, 0x65, 0x2f, 0x76, 0x31, 0x2f, 0x63,
	0x6f, 0x72, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x12, 0x65, 0x78, 0x74, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x2f, 0x65, 0x78, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xd5, 0x03,
	0x0a, 0x07, 0x56, 0x61, 0x75, 0x6c, 0x74, 0x43, 0x41, 0x12, 0x17, 0x0a, 0x07, 0x63, 0x61, 0x5f,
	0x70, 0x61, 0x74, 0x68, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x63, 0x61, 0x50, 0x61,
	0x74, 0x68, 0x12, 0x19, 0x0a, 0x08, 0x63, 0x73, 0x72, 0x5f, 0x70, 0x61, 0x74, 0x68, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x63, 0x73, 0x72, 0x50, 0x61, 0x74, 0x68, 0x12, 0x16, 0x0a,
	0x06, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x73,
	0x65, 0x72, 0x76, 0x65, 0x72, 0x12, 0x1a, 0x0a, 0x08, 0x43, 0x61, 0x42, 0x75, 0x6e, 0x64, 0x6c,
	0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x08, 0x43, 0x61, 0x42, 0x75, 0x6e, 0x64, 0x6c,
	0x65, 0x12, 0x1c, 0x0a, 0x09, 0x6e, 0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65, 0x18, 0x05,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x6e, 0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65, 0x12,
	0x48, 0x0a, 0x10, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x5f, 0x73, 0x65, 0x63, 0x72, 0x65, 0x74, 0x5f,
	0x72, 0x65, 0x66, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1c, 0x2e, 0x63, 0x6f, 0x72, 0x65,
	0x2e, 0x73, 0x6b, 0x76, 0x32, 0x2e, 0x73, 0x6f, 0x6c, 0x6f, 0x2e, 0x69, 0x6f, 0x2e, 0x4f, 0x62,
	0x6a, 0x65, 0x63, 0x74, 0x52, 0x65, 0x66, 0x48, 0x00, 0x52, 0x0e, 0x74, 0x6f, 0x6b, 0x65, 0x6e,
	0x53, 0x65, 0x63, 0x72, 0x65, 0x74, 0x52, 0x65, 0x66, 0x12, 0x57, 0x0a, 0x0f, 0x6b, 0x75, 0x62,
	0x65, 0x72, 0x6e, 0x65, 0x74, 0x65, 0x73, 0x5f, 0x61, 0x75, 0x74, 0x68, 0x18, 0x08, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x2c, 0x2e, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x6d, 0x65, 0x73, 0x68,
	0x2e, 0x67, 0x6c, 0x6f, 0x6f, 0x2e, 0x73, 0x6f, 0x6c, 0x6f, 0x2e, 0x69, 0x6f, 0x2e, 0x56, 0x61,
	0x75, 0x6c, 0x74, 0x43, 0x41, 0x2e, 0x4b, 0x75, 0x62, 0x65, 0x72, 0x6e, 0x65, 0x74, 0x65, 0x73,
	0x48, 0x00, 0x52, 0x0e, 0x6b, 0x75, 0x62, 0x65, 0x72, 0x6e, 0x65, 0x74, 0x65, 0x73, 0x41, 0x75,
	0x74, 0x68, 0x1a, 0x93, 0x01, 0x0a, 0x0a, 0x4b, 0x75, 0x62, 0x65, 0x72, 0x6e, 0x65, 0x74, 0x65,
	0x73, 0x12, 0x12, 0x0a, 0x04, 0x70, 0x61, 0x74, 0x68, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x04, 0x70, 0x61, 0x74, 0x68, 0x12, 0x33, 0x0a, 0x06, 0x73, 0x61, 0x5f, 0x72, 0x65, 0x66, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1c, 0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x73, 0x6b, 0x76,
	0x32, 0x2e, 0x73, 0x6f, 0x6c, 0x6f, 0x2e, 0x69, 0x6f, 0x2e, 0x4f, 0x62, 0x6a, 0x65, 0x63, 0x74,
	0x52, 0x65, 0x66, 0x52, 0x05, 0x73, 0x61, 0x52, 0x65, 0x66, 0x12, 0x28, 0x0a, 0x10, 0x73, 0x65,
	0x63, 0x72, 0x65, 0x74, 0x5f, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x5f, 0x6b, 0x65, 0x79, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x0e, 0x73, 0x65, 0x63, 0x72, 0x65, 0x74, 0x54, 0x6f, 0x6b, 0x65,
	0x6e, 0x4b, 0x65, 0x79, 0x12, 0x12, 0x0a, 0x04, 0x72, 0x6f, 0x6c, 0x65, 0x18, 0x04, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x04, 0x72, 0x6f, 0x6c, 0x65, 0x42, 0x0b, 0x0a, 0x09, 0x61, 0x75, 0x74, 0x68,
	0x5f, 0x74, 0x79, 0x70, 0x65, 0x22, 0x76, 0x0a, 0x11, 0x43, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x43,
	0x65, 0x72, 0x74, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x12, 0x19, 0x0a, 0x08, 0x74, 0x74,
	0x6c, 0x5f, 0x64, 0x61, 0x79, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x07, 0x74, 0x74,
	0x6c, 0x44, 0x61, 0x79, 0x73, 0x12, 0x2b, 0x0a, 0x12, 0x72, 0x73, 0x61, 0x5f, 0x6b, 0x65, 0x79,
	0x5f, 0x73, 0x69, 0x7a, 0x65, 0x5f, 0x62, 0x79, 0x74, 0x65, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x0d, 0x52, 0x0f, 0x72, 0x73, 0x61, 0x4b, 0x65, 0x79, 0x53, 0x69, 0x7a, 0x65, 0x42, 0x79, 0x74,
	0x65, 0x73, 0x12, 0x19, 0x0a, 0x08, 0x6f, 0x72, 0x67, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x6f, 0x72, 0x67, 0x4e, 0x61, 0x6d, 0x65, 0x42, 0x46, 0x5a,
	0x40, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x73, 0x6f, 0x6c, 0x6f,
	0x2d, 0x69, 0x6f, 0x2f, 0x67, 0x6c, 0x6f, 0x6f, 0x2d, 0x6d, 0x65, 0x73, 0x68, 0x2f, 0x70, 0x6b,
	0x67, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x6d, 0x65, 0x73,
	0x68, 0x2e, 0x67, 0x6c, 0x6f, 0x6f, 0x2e, 0x73, 0x6f, 0x6c, 0x6f, 0x2e, 0x69, 0x6f, 0x2f, 0x76,
	0x31, 0xc0, 0xf5, 0x04, 0x01, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_github_com_solo_io_gloo_mesh_api_common_v1_certificates_proto_rawDescOnce sync.Once
	file_github_com_solo_io_gloo_mesh_api_common_v1_certificates_proto_rawDescData = file_github_com_solo_io_gloo_mesh_api_common_v1_certificates_proto_rawDesc
)

func file_github_com_solo_io_gloo_mesh_api_common_v1_certificates_proto_rawDescGZIP() []byte {
	file_github_com_solo_io_gloo_mesh_api_common_v1_certificates_proto_rawDescOnce.Do(func() {
		file_github_com_solo_io_gloo_mesh_api_common_v1_certificates_proto_rawDescData = protoimpl.X.CompressGZIP(file_github_com_solo_io_gloo_mesh_api_common_v1_certificates_proto_rawDescData)
	})
	return file_github_com_solo_io_gloo_mesh_api_common_v1_certificates_proto_rawDescData
}

var file_github_com_solo_io_gloo_mesh_api_common_v1_certificates_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_github_com_solo_io_gloo_mesh_api_common_v1_certificates_proto_goTypes = []interface{}{
	(*VaultCA)(nil),            // 0: common.mesh.gloo.solo.io.VaultCA
	(*CommonCertOptions)(nil),  // 1: common.mesh.gloo.solo.io.CommonCertOptions
	(*VaultCA_Kubernetes)(nil), // 2: common.mesh.gloo.solo.io.VaultCA.Kubernetes
	(*v1.ObjectRef)(nil),       // 3: core.skv2.solo.io.ObjectRef
}
var file_github_com_solo_io_gloo_mesh_api_common_v1_certificates_proto_depIdxs = []int32{
	3, // 0: common.mesh.gloo.solo.io.VaultCA.token_secret_ref:type_name -> core.skv2.solo.io.ObjectRef
	2, // 1: common.mesh.gloo.solo.io.VaultCA.kubernetes_auth:type_name -> common.mesh.gloo.solo.io.VaultCA.Kubernetes
	3, // 2: common.mesh.gloo.solo.io.VaultCA.Kubernetes.sa_ref:type_name -> core.skv2.solo.io.ObjectRef
	3, // [3:3] is the sub-list for method output_type
	3, // [3:3] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_github_com_solo_io_gloo_mesh_api_common_v1_certificates_proto_init() }
func file_github_com_solo_io_gloo_mesh_api_common_v1_certificates_proto_init() {
	if File_github_com_solo_io_gloo_mesh_api_common_v1_certificates_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_github_com_solo_io_gloo_mesh_api_common_v1_certificates_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*VaultCA); i {
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
		file_github_com_solo_io_gloo_mesh_api_common_v1_certificates_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CommonCertOptions); i {
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
		file_github_com_solo_io_gloo_mesh_api_common_v1_certificates_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*VaultCA_Kubernetes); i {
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
	file_github_com_solo_io_gloo_mesh_api_common_v1_certificates_proto_msgTypes[0].OneofWrappers = []interface{}{
		(*VaultCA_TokenSecretRef)(nil),
		(*VaultCA_KubernetesAuth)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_github_com_solo_io_gloo_mesh_api_common_v1_certificates_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_github_com_solo_io_gloo_mesh_api_common_v1_certificates_proto_goTypes,
		DependencyIndexes: file_github_com_solo_io_gloo_mesh_api_common_v1_certificates_proto_depIdxs,
		MessageInfos:      file_github_com_solo_io_gloo_mesh_api_common_v1_certificates_proto_msgTypes,
	}.Build()
	File_github_com_solo_io_gloo_mesh_api_common_v1_certificates_proto = out.File
	file_github_com_solo_io_gloo_mesh_api_common_v1_certificates_proto_rawDesc = nil
	file_github_com_solo_io_gloo_mesh_api_common_v1_certificates_proto_goTypes = nil
	file_github_com_solo_io_gloo_mesh_api_common_v1_certificates_proto_depIdxs = nil
}
