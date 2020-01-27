// This file contains generated Deepcopy methods for
// Protobuf types with oneofs

package v1alpha1

import (
	fmt "fmt"

	proto "github.com/gogo/protobuf/proto"

	math "math"

	_ "github.com/gogo/protobuf/gogoproto"

	_ "github.com/gogo/protobuf/types"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// DeepCopyInto supports using AttributeManifest within kubernetes types, where deepcopy-gen is used.
func (in *DataSource) DeepCopyInto(out *DataSource) {
	p := proto.Clone(in).(*DataSource)
	*out = *p
}

// DeepCopyInto supports using AttributeManifest within kubernetes types, where deepcopy-gen is used.
func (in *AsyncDataSource) DeepCopyInto(out *AsyncDataSource) {
	p := proto.Clone(in).(*AsyncDataSource)
	*out = *p
}

// DeepCopyInto supports using AttributeManifest within kubernetes types, where deepcopy-gen is used.
func (in *TransportSocket) DeepCopyInto(out *TransportSocket) {
	p := proto.Clone(in).(*TransportSocket)
	*out = *p
}

// DeepCopyInto supports using AttributeManifest within kubernetes types, where deepcopy-gen is used.
func (in *SocketOption) DeepCopyInto(out *SocketOption) {
	p := proto.Clone(in).(*SocketOption)
	*out = *p
}

// DeepCopyInto supports using AttributeManifest within kubernetes types, where deepcopy-gen is used.
func (in *HttpUri) DeepCopyInto(out *HttpUri) {
	p := proto.Clone(in).(*HttpUri)
	*out = *p
}

// DeepCopyInto supports using AttributeManifest within kubernetes types, where deepcopy-gen is used.
func (in *Route) DeepCopyInto(out *Route) {
	p := proto.Clone(in).(*Route)
	*out = *p
}

// DeepCopyInto supports using AttributeManifest within kubernetes types, where deepcopy-gen is used.
func (in *RouteMatch) DeepCopyInto(out *RouteMatch) {
	p := proto.Clone(in).(*RouteMatch)
	*out = *p
}

// DeepCopyInto supports using AttributeManifest within kubernetes types, where deepcopy-gen is used.
func (in *CorsPolicy) DeepCopyInto(out *CorsPolicy) {
	p := proto.Clone(in).(*CorsPolicy)
	*out = *p
}

// DeepCopyInto supports using AttributeManifest within kubernetes types, where deepcopy-gen is used.
func (in *RouteAction) DeepCopyInto(out *RouteAction) {
	p := proto.Clone(in).(*RouteAction)
	*out = *p
}

// DeepCopyInto supports using AttributeManifest within kubernetes types, where deepcopy-gen is used.
func (in *RedirectAction) DeepCopyInto(out *RedirectAction) {
	p := proto.Clone(in).(*RedirectAction)
	*out = *p
}

// DeepCopyInto supports using AttributeManifest within kubernetes types, where deepcopy-gen is used.
func (in *HeaderMatcher) DeepCopyInto(out *HeaderMatcher) {
	p := proto.Clone(in).(*HeaderMatcher)
	*out = *p
}

// DeepCopyInto supports using AttributeManifest within kubernetes types, where deepcopy-gen is used.
func (in *FieldRules) DeepCopyInto(out *FieldRules) {
	p := proto.Clone(in).(*FieldRules)
	*out = *p
}

// DeepCopyInto supports using AttributeManifest within kubernetes types, where deepcopy-gen is used.
func (in *StringRules) DeepCopyInto(out *StringRules) {
	p := proto.Clone(in).(*StringRules)
	*out = *p
}

// DeepCopyInto supports using AttributeManifest within kubernetes types, where deepcopy-gen is used.
func (in *BytesRules) DeepCopyInto(out *BytesRules) {
	p := proto.Clone(in).(*BytesRules)
	*out = *p
}

// DeepCopyInto supports using AttributeManifest within kubernetes types, where deepcopy-gen is used.
func (in *Gateway) DeepCopyInto(out *Gateway) {
	p := proto.Clone(in).(*Gateway)
	*out = *p
}

// DeepCopyInto supports using AttributeManifest within kubernetes types, where deepcopy-gen is used.
func (in *Route) DeepCopyInto(out *Route) {
	p := proto.Clone(in).(*Route)
	*out = *p
}

// DeepCopyInto supports using AttributeManifest within kubernetes types, where deepcopy-gen is used.
func (in *DelegateAction) DeepCopyInto(out *DelegateAction) {
	p := proto.Clone(in).(*DelegateAction)
	*out = *p
}

// DeepCopyInto supports using AttributeManifest within kubernetes types, where deepcopy-gen is used.
func (in *HealthCheck) DeepCopyInto(out *HealthCheck) {
	p := proto.Clone(in).(*HealthCheck)
	*out = *p
}

// DeepCopyInto supports using AttributeManifest within kubernetes types, where deepcopy-gen is used.
func (in *Transformation) DeepCopyInto(out *Transformation) {
	p := proto.Clone(in).(*Transformation)
	*out = *p
}

// DeepCopyInto supports using AttributeManifest within kubernetes types, where deepcopy-gen is used.
func (in *Extraction) DeepCopyInto(out *Extraction) {
	p := proto.Clone(in).(*Extraction)
	*out = *p
}

// DeepCopyInto supports using AttributeManifest within kubernetes types, where deepcopy-gen is used.
func (in *TransformationTemplate) DeepCopyInto(out *TransformationTemplate) {
	p := proto.Clone(in).(*TransformationTemplate)
	*out = *p
}

// DeepCopyInto supports using AttributeManifest within kubernetes types, where deepcopy-gen is used.
func (in *Transformation) DeepCopyInto(out *Transformation) {
	p := proto.Clone(in).(*Transformation)
	*out = *p
}

// DeepCopyInto supports using AttributeManifest within kubernetes types, where deepcopy-gen is used.
func (in *ListenerReport) DeepCopyInto(out *ListenerReport) {
	p := proto.Clone(in).(*ListenerReport)
	*out = *p
}

// DeepCopyInto supports using AttributeManifest within kubernetes types, where deepcopy-gen is used.
func (in *ServerVersion) DeepCopyInto(out *ServerVersion) {
	p := proto.Clone(in).(*ServerVersion)
	*out = *p
}

// DeepCopyInto supports using AttributeManifest within kubernetes types, where deepcopy-gen is used.
func (in *Matcher) DeepCopyInto(out *Matcher) {
	p := proto.Clone(in).(*Matcher)
	*out = *p
}

// DeepCopyInto supports using AttributeManifest within kubernetes types, where deepcopy-gen is used.
func (in *ExtAuthExtension) DeepCopyInto(out *ExtAuthExtension) {
	p := proto.Clone(in).(*ExtAuthExtension)
	*out = *p
}

// DeepCopyInto supports using AttributeManifest within kubernetes types, where deepcopy-gen is used.
func (in *Jwks) DeepCopyInto(out *Jwks) {
	p := proto.Clone(in).(*Jwks)
	*out = *p
}

// DeepCopyInto supports using AttributeManifest within kubernetes types, where deepcopy-gen is used.
func (in *Action) DeepCopyInto(out *Action) {
	p := proto.Clone(in).(*Action)
	*out = *p
}

// DeepCopyInto supports using AttributeManifest within kubernetes types, where deepcopy-gen is used.
func (in *HeaderMatcher) DeepCopyInto(out *HeaderMatcher) {
	p := proto.Clone(in).(*HeaderMatcher)
	*out = *p
}

// DeepCopyInto supports using AttributeManifest within kubernetes types, where deepcopy-gen is used.
func (in *CoreRuleSet) DeepCopyInto(out *CoreRuleSet) {
	p := proto.Clone(in).(*CoreRuleSet)
	*out = *p
}

// DeepCopyInto supports using AttributeManifest within kubernetes types, where deepcopy-gen is used.
func (in *LoadBalancerConfig) DeepCopyInto(out *LoadBalancerConfig) {
	p := proto.Clone(in).(*LoadBalancerConfig)
	*out = *p
}

// DeepCopyInto supports using AttributeManifest within kubernetes types, where deepcopy-gen is used.
func (in *RouteOptions) DeepCopyInto(out *RouteOptions) {
	p := proto.Clone(in).(*RouteOptions)
	*out = *p
}

// DeepCopyInto supports using AttributeManifest within kubernetes types, where deepcopy-gen is used.
func (in *DestinationSpec) DeepCopyInto(out *DestinationSpec) {
	p := proto.Clone(in).(*DestinationSpec)
	*out = *p
}

// DeepCopyInto supports using AttributeManifest within kubernetes types, where deepcopy-gen is used.
func (in *AccessLog) DeepCopyInto(out *AccessLog) {
	p := proto.Clone(in).(*AccessLog)
	*out = *p
}

// DeepCopyInto supports using AttributeManifest within kubernetes types, where deepcopy-gen is used.
func (in *FileSink) DeepCopyInto(out *FileSink) {
	p := proto.Clone(in).(*FileSink)
	*out = *p
}

// DeepCopyInto supports using AttributeManifest within kubernetes types, where deepcopy-gen is used.
func (in *GrpcService) DeepCopyInto(out *GrpcService) {
	p := proto.Clone(in).(*GrpcService)
	*out = *p
}

// DeepCopyInto supports using AttributeManifest within kubernetes types, where deepcopy-gen is used.
func (in *TagFilter) DeepCopyInto(out *TagFilter) {
	p := proto.Clone(in).(*TagFilter)
	*out = *p
}

// DeepCopyInto supports using AttributeManifest within kubernetes types, where deepcopy-gen is used.
func (in *HashPolicy) DeepCopyInto(out *HashPolicy) {
	p := proto.Clone(in).(*HashPolicy)
	*out = *p
}

// DeepCopyInto supports using AttributeManifest within kubernetes types, where deepcopy-gen is used.
func (in *ProtocolUpgradeConfig) DeepCopyInto(out *ProtocolUpgradeConfig) {
	p := proto.Clone(in).(*ProtocolUpgradeConfig)
	*out = *p
}

// DeepCopyInto supports using AttributeManifest within kubernetes types, where deepcopy-gen is used.
func (in *ServiceSpec) DeepCopyInto(out *ServiceSpec) {
	p := proto.Clone(in).(*ServiceSpec)
	*out = *p
}

// DeepCopyInto supports using AttributeManifest within kubernetes types, where deepcopy-gen is used.
func (in *Listener) DeepCopyInto(out *Listener) {
	p := proto.Clone(in).(*Listener)
	*out = *p
}

// DeepCopyInto supports using AttributeManifest within kubernetes types, where deepcopy-gen is used.
func (in *Route) DeepCopyInto(out *Route) {
	p := proto.Clone(in).(*Route)
	*out = *p
}

// DeepCopyInto supports using AttributeManifest within kubernetes types, where deepcopy-gen is used.
func (in *RouteAction) DeepCopyInto(out *RouteAction) {
	p := proto.Clone(in).(*RouteAction)
	*out = *p
}

// DeepCopyInto supports using AttributeManifest within kubernetes types, where deepcopy-gen is used.
func (in *Destination) DeepCopyInto(out *Destination) {
	p := proto.Clone(in).(*Destination)
	*out = *p
}

// DeepCopyInto supports using AttributeManifest within kubernetes types, where deepcopy-gen is used.
func (in *RedirectAction) DeepCopyInto(out *RedirectAction) {
	p := proto.Clone(in).(*RedirectAction)
	*out = *p
}

// DeepCopyInto supports using AttributeManifest within kubernetes types, where deepcopy-gen is used.
func (in *Secret) DeepCopyInto(out *Secret) {
	p := proto.Clone(in).(*Secret)
	*out = *p
}

// DeepCopyInto supports using AttributeManifest within kubernetes types, where deepcopy-gen is used.
func (in *Settings) DeepCopyInto(out *Settings) {
	p := proto.Clone(in).(*Settings)
	*out = *p
}

// DeepCopyInto supports using AttributeManifest within kubernetes types, where deepcopy-gen is used.
func (in *SslConfig) DeepCopyInto(out *SslConfig) {
	p := proto.Clone(in).(*SslConfig)
	*out = *p
}

// DeepCopyInto supports using AttributeManifest within kubernetes types, where deepcopy-gen is used.
func (in *UpstreamSslConfig) DeepCopyInto(out *UpstreamSslConfig) {
	p := proto.Clone(in).(*UpstreamSslConfig)
	*out = *p
}

// DeepCopyInto supports using AttributeManifest within kubernetes types, where deepcopy-gen is used.
func (in *Upstream) DeepCopyInto(out *Upstream) {
	p := proto.Clone(in).(*Upstream)
	*out = *p
}

// DeepCopyInto supports using AttributeManifest within kubernetes types, where deepcopy-gen is used.
func (in *RoutingRuleSpec) DeepCopyInto(out *RoutingRuleSpec) {
	p := proto.Clone(in).(*RoutingRuleSpec)
	*out = *p
}

// DeepCopyInto supports using AttributeManifest within kubernetes types, where deepcopy-gen is used.
func (in *FaultInjection) DeepCopyInto(out *FaultInjection) {
	p := proto.Clone(in).(*FaultInjection)
	*out = *p
}

// DeepCopyInto supports using AttributeManifest within kubernetes types, where deepcopy-gen is used.
func (in *Matcher) DeepCopyInto(out *Matcher) {
	p := proto.Clone(in).(*Matcher)
	*out = *p
}

// DeepCopyInto supports using AttributeManifest within kubernetes types, where deepcopy-gen is used.
func (in *StringMatch) DeepCopyInto(out *StringMatch) {
	p := proto.Clone(in).(*StringMatch)
	*out = *p
}

// DeepCopyInto supports using AttributeManifest within kubernetes types, where deepcopy-gen is used.
func (in *HeaderMatcher) DeepCopyInto(out *HeaderMatcher) {
	p := proto.Clone(in).(*HeaderMatcher)
	*out = *p
}

// DeepCopyInto supports using AttributeManifest within kubernetes types, where deepcopy-gen is used.
func (in *QueryParameterMatcher) DeepCopyInto(out *QueryParameterMatcher) {
	p := proto.Clone(in).(*QueryParameterMatcher)
	*out = *p
}

// DeepCopyInto supports using AttributeManifest within kubernetes types, where deepcopy-gen is used.
func (in *MeshSpec) DeepCopyInto(out *MeshSpec) {
	p := proto.Clone(in).(*MeshSpec)
	*out = *p
}

// DeepCopyInto supports using AttributeManifest within kubernetes types, where deepcopy-gen is used.
func (in *StringMatch) DeepCopyInto(out *StringMatch) {
	p := proto.Clone(in).(*StringMatch)
	*out = *p
}

// DeepCopyInto supports using AttributeManifest within kubernetes types, where deepcopy-gen is used.
func (in *PeerAuthenticationMethod) DeepCopyInto(out *PeerAuthenticationMethod) {
	p := proto.Clone(in).(*PeerAuthenticationMethod)
	*out = *p
}

// DeepCopyInto supports using AttributeManifest within kubernetes types, where deepcopy-gen is used.
func (in *PortSelector) DeepCopyInto(out *PortSelector) {
	p := proto.Clone(in).(*PortSelector)
	*out = *p
}

// DeepCopyInto supports using AttributeManifest within kubernetes types, where deepcopy-gen is used.
func (in *MeshIngress) DeepCopyInto(out *MeshIngress) {
	p := proto.Clone(in).(*MeshIngress)
	*out = *p
}

// DeepCopyInto supports using AttributeManifest within kubernetes types, where deepcopy-gen is used.
func (in *Mesh) DeepCopyInto(out *Mesh) {
	p := proto.Clone(in).(*Mesh)
	*out = *p
}

// DeepCopyInto supports using AttributeManifest within kubernetes types, where deepcopy-gen is used.
func (in *RbacMode) DeepCopyInto(out *RbacMode) {
	p := proto.Clone(in).(*RbacMode)
	*out = *p
}

// DeepCopyInto supports using AttributeManifest within kubernetes types, where deepcopy-gen is used.
func (in *PodSelector) DeepCopyInto(out *PodSelector) {
	p := proto.Clone(in).(*PodSelector)
	*out = *p
}

// DeepCopyInto supports using AttributeManifest within kubernetes types, where deepcopy-gen is used.
func (in *HttpRule) DeepCopyInto(out *HttpRule) {
	p := proto.Clone(in).(*HttpRule)
	*out = *p
}

// DeepCopyInto supports using AttributeManifest within kubernetes types, where deepcopy-gen is used.
func (in *Value) DeepCopyInto(out *Value) {
	p := proto.Clone(in).(*Value)
	*out = *p
}

// DeepCopyInto supports using AttributeManifest within kubernetes types, where deepcopy-gen is used.
func (in *AttributeValue) DeepCopyInto(out *AttributeValue) {
	p := proto.Clone(in).(*AttributeValue)
	*out = *p
}

// DeepCopyInto supports using AttributeManifest within kubernetes types, where deepcopy-gen is used.
func (in *LoadBalancerSettings) DeepCopyInto(out *LoadBalancerSettings) {
	p := proto.Clone(in).(*LoadBalancerSettings)
	*out = *p
}

// DeepCopyInto supports using AttributeManifest within kubernetes types, where deepcopy-gen is used.
func (in *StringMatch) DeepCopyInto(out *StringMatch) {
	p := proto.Clone(in).(*StringMatch)
	*out = *p
}
