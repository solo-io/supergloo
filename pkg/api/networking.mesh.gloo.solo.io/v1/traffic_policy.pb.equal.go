// Code generated by protoc-gen-ext. DO NOT EDIT.
// source: github.com/solo-io/gloo-mesh/api/networking/v1/traffic_policy.proto

package v1

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"strings"

	"github.com/golang/protobuf/proto"
	equality "github.com/solo-io/protoc-gen-ext/pkg/equality"
)

// ensure the imports are used
var (
	_ = errors.New("")
	_ = fmt.Print
	_ = binary.LittleEndian
	_ = bytes.Compare
	_ = strings.Compare
	_ = equality.Equalizer(nil)
	_ = proto.Message(nil)
)

// Equal function
func (m *TrafficPolicySpec) Equal(that interface{}) bool {
	if that == nil {
		return m == nil
	}

	target, ok := that.(*TrafficPolicySpec)
	if !ok {
		that2, ok := that.(TrafficPolicySpec)
		if ok {
			target = &that2
		} else {
			return false
		}
	}
	if target == nil {
		return m == nil
	} else if m == nil {
		return false
	}

	if len(m.GetSourceSelector()) != len(target.GetSourceSelector()) {
		return false
	}
	for idx, v := range m.GetSourceSelector() {

		if h, ok := interface{}(v).(equality.Equalizer); ok {
			if !h.Equal(target.GetSourceSelector()[idx]) {
				return false
			}
		} else {
			if !proto.Equal(v, target.GetSourceSelector()[idx]) {
				return false
			}
		}

	}

	if len(m.GetDestinationSelector()) != len(target.GetDestinationSelector()) {
		return false
	}
	for idx, v := range m.GetDestinationSelector() {

		if h, ok := interface{}(v).(equality.Equalizer); ok {
			if !h.Equal(target.GetDestinationSelector()[idx]) {
				return false
			}
		} else {
			if !proto.Equal(v, target.GetDestinationSelector()[idx]) {
				return false
			}
		}

	}

	if len(m.GetHttpRequestMatchers()) != len(target.GetHttpRequestMatchers()) {
		return false
	}
	for idx, v := range m.GetHttpRequestMatchers() {

		if h, ok := interface{}(v).(equality.Equalizer); ok {
			if !h.Equal(target.GetHttpRequestMatchers()[idx]) {
				return false
			}
		} else {
			if !proto.Equal(v, target.GetHttpRequestMatchers()[idx]) {
				return false
			}
		}

	}

	if h, ok := interface{}(m.GetPolicy()).(equality.Equalizer); ok {
		if !h.Equal(target.GetPolicy()) {
			return false
		}
	} else {
		if !proto.Equal(m.GetPolicy(), target.GetPolicy()) {
			return false
		}
	}

	return true
}

// Equal function
func (m *TrafficPolicyStatus) Equal(that interface{}) bool {
	if that == nil {
		return m == nil
	}

	target, ok := that.(*TrafficPolicyStatus)
	if !ok {
		that2, ok := that.(TrafficPolicyStatus)
		if ok {
			target = &that2
		} else {
			return false
		}
	}
	if target == nil {
		return m == nil
	} else if m == nil {
		return false
	}

	if m.GetObservedGeneration() != target.GetObservedGeneration() {
		return false
	}

	if m.GetState() != target.GetState() {
		return false
	}

	if len(m.GetDestinations()) != len(target.GetDestinations()) {
		return false
	}
	for k, v := range m.GetDestinations() {

		if h, ok := interface{}(v).(equality.Equalizer); ok {
			if !h.Equal(target.GetDestinations()[k]) {
				return false
			}
		} else {
			if !proto.Equal(v, target.GetDestinations()[k]) {
				return false
			}
		}

	}

	if len(m.GetWorkloads()) != len(target.GetWorkloads()) {
		return false
	}
	for idx, v := range m.GetWorkloads() {

		if strings.Compare(v, target.GetWorkloads()[idx]) != 0 {
			return false
		}

	}

	if len(m.GetErrors()) != len(target.GetErrors()) {
		return false
	}
	for idx, v := range m.GetErrors() {

		if strings.Compare(v, target.GetErrors()[idx]) != 0 {
			return false
		}

	}

	return true
}

// Equal function
func (m *TrafficPolicySpec_HttpMatcher) Equal(that interface{}) bool {
	if that == nil {
		return m == nil
	}

	target, ok := that.(*TrafficPolicySpec_HttpMatcher)
	if !ok {
		that2, ok := that.(TrafficPolicySpec_HttpMatcher)
		if ok {
			target = &that2
		} else {
			return false
		}
	}
	if target == nil {
		return m == nil
	} else if m == nil {
		return false
	}

	if len(m.GetHeaders()) != len(target.GetHeaders()) {
		return false
	}
	for idx, v := range m.GetHeaders() {

		if h, ok := interface{}(v).(equality.Equalizer); ok {
			if !h.Equal(target.GetHeaders()[idx]) {
				return false
			}
		} else {
			if !proto.Equal(v, target.GetHeaders()[idx]) {
				return false
			}
		}

	}

	if len(m.GetQueryParameters()) != len(target.GetQueryParameters()) {
		return false
	}
	for idx, v := range m.GetQueryParameters() {

		if h, ok := interface{}(v).(equality.Equalizer); ok {
			if !h.Equal(target.GetQueryParameters()[idx]) {
				return false
			}
		} else {
			if !proto.Equal(v, target.GetQueryParameters()[idx]) {
				return false
			}
		}

	}

	if strings.Compare(m.GetMethod(), target.GetMethod()) != 0 {
		return false
	}

	switch m.PathSpecifier.(type) {

	case *TrafficPolicySpec_HttpMatcher_Prefix:
		if _, ok := target.PathSpecifier.(*TrafficPolicySpec_HttpMatcher_Prefix); !ok {
			return false
		}

		if strings.Compare(m.GetPrefix(), target.GetPrefix()) != 0 {
			return false
		}

	case *TrafficPolicySpec_HttpMatcher_Exact:
		if _, ok := target.PathSpecifier.(*TrafficPolicySpec_HttpMatcher_Exact); !ok {
			return false
		}

		if strings.Compare(m.GetExact(), target.GetExact()) != 0 {
			return false
		}

	case *TrafficPolicySpec_HttpMatcher_Regex:
		if _, ok := target.PathSpecifier.(*TrafficPolicySpec_HttpMatcher_Regex); !ok {
			return false
		}

		if strings.Compare(m.GetRegex(), target.GetRegex()) != 0 {
			return false
		}

	default:
		// m is nil but target is not nil
		if m.PathSpecifier != target.PathSpecifier {
			return false
		}
	}

	return true
}

// Equal function
func (m *TrafficPolicySpec_Policy) Equal(that interface{}) bool {
	if that == nil {
		return m == nil
	}

	target, ok := that.(*TrafficPolicySpec_Policy)
	if !ok {
		that2, ok := that.(TrafficPolicySpec_Policy)
		if ok {
			target = &that2
		} else {
			return false
		}
	}
	if target == nil {
		return m == nil
	} else if m == nil {
		return false
	}

	if h, ok := interface{}(m.GetTrafficShift()).(equality.Equalizer); ok {
		if !h.Equal(target.GetTrafficShift()) {
			return false
		}
	} else {
		if !proto.Equal(m.GetTrafficShift(), target.GetTrafficShift()) {
			return false
		}
	}

	if h, ok := interface{}(m.GetFaultInjection()).(equality.Equalizer); ok {
		if !h.Equal(target.GetFaultInjection()) {
			return false
		}
	} else {
		if !proto.Equal(m.GetFaultInjection(), target.GetFaultInjection()) {
			return false
		}
	}

	if h, ok := interface{}(m.GetRequestTimeout()).(equality.Equalizer); ok {
		if !h.Equal(target.GetRequestTimeout()) {
			return false
		}
	} else {
		if !proto.Equal(m.GetRequestTimeout(), target.GetRequestTimeout()) {
			return false
		}
	}

	if h, ok := interface{}(m.GetRetries()).(equality.Equalizer); ok {
		if !h.Equal(target.GetRetries()) {
			return false
		}
	} else {
		if !proto.Equal(m.GetRetries(), target.GetRetries()) {
			return false
		}
	}

	if h, ok := interface{}(m.GetCorsPolicy()).(equality.Equalizer); ok {
		if !h.Equal(target.GetCorsPolicy()) {
			return false
		}
	} else {
		if !proto.Equal(m.GetCorsPolicy(), target.GetCorsPolicy()) {
			return false
		}
	}

	if h, ok := interface{}(m.GetMirror()).(equality.Equalizer); ok {
		if !h.Equal(target.GetMirror()) {
			return false
		}
	} else {
		if !proto.Equal(m.GetMirror(), target.GetMirror()) {
			return false
		}
	}

	if h, ok := interface{}(m.GetHeaderManipulation()).(equality.Equalizer); ok {
		if !h.Equal(target.GetHeaderManipulation()) {
			return false
		}
	} else {
		if !proto.Equal(m.GetHeaderManipulation(), target.GetHeaderManipulation()) {
			return false
		}
	}

	if h, ok := interface{}(m.GetOutlierDetection()).(equality.Equalizer); ok {
		if !h.Equal(target.GetOutlierDetection()) {
			return false
		}
	} else {
		if !proto.Equal(m.GetOutlierDetection(), target.GetOutlierDetection()) {
			return false
		}
	}

	if h, ok := interface{}(m.GetMtls()).(equality.Equalizer); ok {
		if !h.Equal(target.GetMtls()) {
			return false
		}
	} else {
		if !proto.Equal(m.GetMtls(), target.GetMtls()) {
			return false
		}
	}

	return true
}

// Equal function
func (m *TrafficPolicySpec_HttpMatcher_QueryParameterMatcher) Equal(that interface{}) bool {
	if that == nil {
		return m == nil
	}

	target, ok := that.(*TrafficPolicySpec_HttpMatcher_QueryParameterMatcher)
	if !ok {
		that2, ok := that.(TrafficPolicySpec_HttpMatcher_QueryParameterMatcher)
		if ok {
			target = &that2
		} else {
			return false
		}
	}
	if target == nil {
		return m == nil
	} else if m == nil {
		return false
	}

	if strings.Compare(m.GetName(), target.GetName()) != 0 {
		return false
	}

	if strings.Compare(m.GetValue(), target.GetValue()) != 0 {
		return false
	}

	if m.GetRegex() != target.GetRegex() {
		return false
	}

	return true
}

// Equal function
func (m *TrafficPolicySpec_Policy_RetryPolicy) Equal(that interface{}) bool {
	if that == nil {
		return m == nil
	}

	target, ok := that.(*TrafficPolicySpec_Policy_RetryPolicy)
	if !ok {
		that2, ok := that.(TrafficPolicySpec_Policy_RetryPolicy)
		if ok {
			target = &that2
		} else {
			return false
		}
	}
	if target == nil {
		return m == nil
	} else if m == nil {
		return false
	}

	if m.GetAttempts() != target.GetAttempts() {
		return false
	}

	if h, ok := interface{}(m.GetPerTryTimeout()).(equality.Equalizer); ok {
		if !h.Equal(target.GetPerTryTimeout()) {
			return false
		}
	} else {
		if !proto.Equal(m.GetPerTryTimeout(), target.GetPerTryTimeout()) {
			return false
		}
	}

	return true
}

// Equal function
func (m *TrafficPolicySpec_Policy_MultiDestination) Equal(that interface{}) bool {
	if that == nil {
		return m == nil
	}

	target, ok := that.(*TrafficPolicySpec_Policy_MultiDestination)
	if !ok {
		that2, ok := that.(TrafficPolicySpec_Policy_MultiDestination)
		if ok {
			target = &that2
		} else {
			return false
		}
	}
	if target == nil {
		return m == nil
	} else if m == nil {
		return false
	}

	if len(m.GetDestinations()) != len(target.GetDestinations()) {
		return false
	}
	for idx, v := range m.GetDestinations() {

		if h, ok := interface{}(v).(equality.Equalizer); ok {
			if !h.Equal(target.GetDestinations()[idx]) {
				return false
			}
		} else {
			if !proto.Equal(v, target.GetDestinations()[idx]) {
				return false
			}
		}

	}

	return true
}

// Equal function
func (m *TrafficPolicySpec_Policy_FaultInjection) Equal(that interface{}) bool {
	if that == nil {
		return m == nil
	}

	target, ok := that.(*TrafficPolicySpec_Policy_FaultInjection)
	if !ok {
		that2, ok := that.(TrafficPolicySpec_Policy_FaultInjection)
		if ok {
			target = &that2
		} else {
			return false
		}
	}
	if target == nil {
		return m == nil
	} else if m == nil {
		return false
	}

	if m.GetPercentage() != target.GetPercentage() {
		return false
	}

	switch m.FaultInjectionType.(type) {

	case *TrafficPolicySpec_Policy_FaultInjection_FixedDelay:
		if _, ok := target.FaultInjectionType.(*TrafficPolicySpec_Policy_FaultInjection_FixedDelay); !ok {
			return false
		}

		if h, ok := interface{}(m.GetFixedDelay()).(equality.Equalizer); ok {
			if !h.Equal(target.GetFixedDelay()) {
				return false
			}
		} else {
			if !proto.Equal(m.GetFixedDelay(), target.GetFixedDelay()) {
				return false
			}
		}

	case *TrafficPolicySpec_Policy_FaultInjection_Abort_:
		if _, ok := target.FaultInjectionType.(*TrafficPolicySpec_Policy_FaultInjection_Abort_); !ok {
			return false
		}

		if h, ok := interface{}(m.GetAbort()).(equality.Equalizer); ok {
			if !h.Equal(target.GetAbort()) {
				return false
			}
		} else {
			if !proto.Equal(m.GetAbort(), target.GetAbort()) {
				return false
			}
		}

	default:
		// m is nil but target is not nil
		if m.FaultInjectionType != target.FaultInjectionType {
			return false
		}
	}

	return true
}

// Equal function
func (m *TrafficPolicySpec_Policy_HeaderManipulation) Equal(that interface{}) bool {
	if that == nil {
		return m == nil
	}

	target, ok := that.(*TrafficPolicySpec_Policy_HeaderManipulation)
	if !ok {
		that2, ok := that.(TrafficPolicySpec_Policy_HeaderManipulation)
		if ok {
			target = &that2
		} else {
			return false
		}
	}
	if target == nil {
		return m == nil
	} else if m == nil {
		return false
	}

	if len(m.GetRemoveResponseHeaders()) != len(target.GetRemoveResponseHeaders()) {
		return false
	}
	for idx, v := range m.GetRemoveResponseHeaders() {

		if strings.Compare(v, target.GetRemoveResponseHeaders()[idx]) != 0 {
			return false
		}

	}

	if len(m.GetAppendResponseHeaders()) != len(target.GetAppendResponseHeaders()) {
		return false
	}
	for k, v := range m.GetAppendResponseHeaders() {

		if strings.Compare(v, target.GetAppendResponseHeaders()[k]) != 0 {
			return false
		}

	}

	if len(m.GetRemoveRequestHeaders()) != len(target.GetRemoveRequestHeaders()) {
		return false
	}
	for idx, v := range m.GetRemoveRequestHeaders() {

		if strings.Compare(v, target.GetRemoveRequestHeaders()[idx]) != 0 {
			return false
		}

	}

	if len(m.GetAppendRequestHeaders()) != len(target.GetAppendRequestHeaders()) {
		return false
	}
	for k, v := range m.GetAppendRequestHeaders() {

		if strings.Compare(v, target.GetAppendRequestHeaders()[k]) != 0 {
			return false
		}

	}

	return true
}

// Equal function
func (m *TrafficPolicySpec_Policy_CorsPolicy) Equal(that interface{}) bool {
	if that == nil {
		return m == nil
	}

	target, ok := that.(*TrafficPolicySpec_Policy_CorsPolicy)
	if !ok {
		that2, ok := that.(TrafficPolicySpec_Policy_CorsPolicy)
		if ok {
			target = &that2
		} else {
			return false
		}
	}
	if target == nil {
		return m == nil
	} else if m == nil {
		return false
	}

	if len(m.GetAllowOrigins()) != len(target.GetAllowOrigins()) {
		return false
	}
	for idx, v := range m.GetAllowOrigins() {

		if h, ok := interface{}(v).(equality.Equalizer); ok {
			if !h.Equal(target.GetAllowOrigins()[idx]) {
				return false
			}
		} else {
			if !proto.Equal(v, target.GetAllowOrigins()[idx]) {
				return false
			}
		}

	}

	if len(m.GetAllowMethods()) != len(target.GetAllowMethods()) {
		return false
	}
	for idx, v := range m.GetAllowMethods() {

		if strings.Compare(v, target.GetAllowMethods()[idx]) != 0 {
			return false
		}

	}

	if len(m.GetAllowHeaders()) != len(target.GetAllowHeaders()) {
		return false
	}
	for idx, v := range m.GetAllowHeaders() {

		if strings.Compare(v, target.GetAllowHeaders()[idx]) != 0 {
			return false
		}

	}

	if len(m.GetExposeHeaders()) != len(target.GetExposeHeaders()) {
		return false
	}
	for idx, v := range m.GetExposeHeaders() {

		if strings.Compare(v, target.GetExposeHeaders()[idx]) != 0 {
			return false
		}

	}

	if h, ok := interface{}(m.GetMaxAge()).(equality.Equalizer); ok {
		if !h.Equal(target.GetMaxAge()) {
			return false
		}
	} else {
		if !proto.Equal(m.GetMaxAge(), target.GetMaxAge()) {
			return false
		}
	}

	if h, ok := interface{}(m.GetAllowCredentials()).(equality.Equalizer); ok {
		if !h.Equal(target.GetAllowCredentials()) {
			return false
		}
	} else {
		if !proto.Equal(m.GetAllowCredentials(), target.GetAllowCredentials()) {
			return false
		}
	}

	return true
}

// Equal function
func (m *TrafficPolicySpec_Policy_Mirror) Equal(that interface{}) bool {
	if that == nil {
		return m == nil
	}

	target, ok := that.(*TrafficPolicySpec_Policy_Mirror)
	if !ok {
		that2, ok := that.(TrafficPolicySpec_Policy_Mirror)
		if ok {
			target = &that2
		} else {
			return false
		}
	}
	if target == nil {
		return m == nil
	} else if m == nil {
		return false
	}

	if m.GetPercentage() != target.GetPercentage() {
		return false
	}

	if m.GetPort() != target.GetPort() {
		return false
	}

	switch m.DestinationType.(type) {

	case *TrafficPolicySpec_Policy_Mirror_KubeService:
		if _, ok := target.DestinationType.(*TrafficPolicySpec_Policy_Mirror_KubeService); !ok {
			return false
		}

		if h, ok := interface{}(m.GetKubeService()).(equality.Equalizer); ok {
			if !h.Equal(target.GetKubeService()) {
				return false
			}
		} else {
			if !proto.Equal(m.GetKubeService(), target.GetKubeService()) {
				return false
			}
		}

	default:
		// m is nil but target is not nil
		if m.DestinationType != target.DestinationType {
			return false
		}
	}

	return true
}

// Equal function
func (m *TrafficPolicySpec_Policy_OutlierDetection) Equal(that interface{}) bool {
	if that == nil {
		return m == nil
	}

	target, ok := that.(*TrafficPolicySpec_Policy_OutlierDetection)
	if !ok {
		that2, ok := that.(TrafficPolicySpec_Policy_OutlierDetection)
		if ok {
			target = &that2
		} else {
			return false
		}
	}
	if target == nil {
		return m == nil
	} else if m == nil {
		return false
	}

	if m.GetConsecutiveErrors() != target.GetConsecutiveErrors() {
		return false
	}

	if h, ok := interface{}(m.GetInterval()).(equality.Equalizer); ok {
		if !h.Equal(target.GetInterval()) {
			return false
		}
	} else {
		if !proto.Equal(m.GetInterval(), target.GetInterval()) {
			return false
		}
	}

	if h, ok := interface{}(m.GetBaseEjectionTime()).(equality.Equalizer); ok {
		if !h.Equal(target.GetBaseEjectionTime()) {
			return false
		}
	} else {
		if !proto.Equal(m.GetBaseEjectionTime(), target.GetBaseEjectionTime()) {
			return false
		}
	}

	if m.GetMaxEjectionPercent() != target.GetMaxEjectionPercent() {
		return false
	}

	return true
}

// Equal function
func (m *TrafficPolicySpec_Policy_MTLS) Equal(that interface{}) bool {
	if that == nil {
		return m == nil
	}

	target, ok := that.(*TrafficPolicySpec_Policy_MTLS)
	if !ok {
		that2, ok := that.(TrafficPolicySpec_Policy_MTLS)
		if ok {
			target = &that2
		} else {
			return false
		}
	}
	if target == nil {
		return m == nil
	} else if m == nil {
		return false
	}

	if h, ok := interface{}(m.GetIstio()).(equality.Equalizer); ok {
		if !h.Equal(target.GetIstio()) {
			return false
		}
	} else {
		if !proto.Equal(m.GetIstio(), target.GetIstio()) {
			return false
		}
	}

	return true
}

// Equal function
func (m *TrafficPolicySpec_Policy_MultiDestination_WeightedDestination) Equal(that interface{}) bool {
	if that == nil {
		return m == nil
	}

	target, ok := that.(*TrafficPolicySpec_Policy_MultiDestination_WeightedDestination)
	if !ok {
		that2, ok := that.(TrafficPolicySpec_Policy_MultiDestination_WeightedDestination)
		if ok {
			target = &that2
		} else {
			return false
		}
	}
	if target == nil {
		return m == nil
	} else if m == nil {
		return false
	}

	if m.GetWeight() != target.GetWeight() {
		return false
	}

	switch m.DestinationType.(type) {

	case *TrafficPolicySpec_Policy_MultiDestination_WeightedDestination_KubeService:
		if _, ok := target.DestinationType.(*TrafficPolicySpec_Policy_MultiDestination_WeightedDestination_KubeService); !ok {
			return false
		}

		if h, ok := interface{}(m.GetKubeService()).(equality.Equalizer); ok {
			if !h.Equal(target.GetKubeService()) {
				return false
			}
		} else {
			if !proto.Equal(m.GetKubeService(), target.GetKubeService()) {
				return false
			}
		}

	case *TrafficPolicySpec_Policy_MultiDestination_WeightedDestination_VirtualDestination:
		if _, ok := target.DestinationType.(*TrafficPolicySpec_Policy_MultiDestination_WeightedDestination_VirtualDestination); !ok {
			return false
		}

		if h, ok := interface{}(m.GetVirtualDestination()).(equality.Equalizer); ok {
			if !h.Equal(target.GetVirtualDestination()) {
				return false
			}
		} else {
			if !proto.Equal(m.GetVirtualDestination(), target.GetVirtualDestination()) {
				return false
			}
		}

	default:
		// m is nil but target is not nil
		if m.DestinationType != target.DestinationType {
			return false
		}
	}

	return true
}

// Equal function
func (m *TrafficPolicySpec_Policy_MultiDestination_WeightedDestination_KubeDestination) Equal(that interface{}) bool {
	if that == nil {
		return m == nil
	}

	target, ok := that.(*TrafficPolicySpec_Policy_MultiDestination_WeightedDestination_KubeDestination)
	if !ok {
		that2, ok := that.(TrafficPolicySpec_Policy_MultiDestination_WeightedDestination_KubeDestination)
		if ok {
			target = &that2
		} else {
			return false
		}
	}
	if target == nil {
		return m == nil
	} else if m == nil {
		return false
	}

	if strings.Compare(m.GetName(), target.GetName()) != 0 {
		return false
	}

	if strings.Compare(m.GetNamespace(), target.GetNamespace()) != 0 {
		return false
	}

	if strings.Compare(m.GetClusterName(), target.GetClusterName()) != 0 {
		return false
	}

	if len(m.GetSubset()) != len(target.GetSubset()) {
		return false
	}
	for k, v := range m.GetSubset() {

		if strings.Compare(v, target.GetSubset()[k]) != 0 {
			return false
		}

	}

	if m.GetPort() != target.GetPort() {
		return false
	}

	return true
}

// Equal function
func (m *TrafficPolicySpec_Policy_MultiDestination_WeightedDestination_VirtualDestinationReference) Equal(that interface{}) bool {
	if that == nil {
		return m == nil
	}

	target, ok := that.(*TrafficPolicySpec_Policy_MultiDestination_WeightedDestination_VirtualDestinationReference)
	if !ok {
		that2, ok := that.(TrafficPolicySpec_Policy_MultiDestination_WeightedDestination_VirtualDestinationReference)
		if ok {
			target = &that2
		} else {
			return false
		}
	}
	if target == nil {
		return m == nil
	} else if m == nil {
		return false
	}

	if strings.Compare(m.GetName(), target.GetName()) != 0 {
		return false
	}

	if strings.Compare(m.GetNamespace(), target.GetNamespace()) != 0 {
		return false
	}

	if len(m.GetSubset()) != len(target.GetSubset()) {
		return false
	}
	for k, v := range m.GetSubset() {

		if strings.Compare(v, target.GetSubset()[k]) != 0 {
			return false
		}

	}

	return true
}

// Equal function
func (m *TrafficPolicySpec_Policy_FaultInjection_Abort) Equal(that interface{}) bool {
	if that == nil {
		return m == nil
	}

	target, ok := that.(*TrafficPolicySpec_Policy_FaultInjection_Abort)
	if !ok {
		that2, ok := that.(TrafficPolicySpec_Policy_FaultInjection_Abort)
		if ok {
			target = &that2
		} else {
			return false
		}
	}
	if target == nil {
		return m == nil
	} else if m == nil {
		return false
	}

	if m.GetHttpStatus() != target.GetHttpStatus() {
		return false
	}

	return true
}

// Equal function
func (m *TrafficPolicySpec_Policy_CorsPolicy_StringMatch) Equal(that interface{}) bool {
	if that == nil {
		return m == nil
	}

	target, ok := that.(*TrafficPolicySpec_Policy_CorsPolicy_StringMatch)
	if !ok {
		that2, ok := that.(TrafficPolicySpec_Policy_CorsPolicy_StringMatch)
		if ok {
			target = &that2
		} else {
			return false
		}
	}
	if target == nil {
		return m == nil
	} else if m == nil {
		return false
	}

	switch m.MatchType.(type) {

	case *TrafficPolicySpec_Policy_CorsPolicy_StringMatch_Exact:
		if _, ok := target.MatchType.(*TrafficPolicySpec_Policy_CorsPolicy_StringMatch_Exact); !ok {
			return false
		}

		if strings.Compare(m.GetExact(), target.GetExact()) != 0 {
			return false
		}

	case *TrafficPolicySpec_Policy_CorsPolicy_StringMatch_Prefix:
		if _, ok := target.MatchType.(*TrafficPolicySpec_Policy_CorsPolicy_StringMatch_Prefix); !ok {
			return false
		}

		if strings.Compare(m.GetPrefix(), target.GetPrefix()) != 0 {
			return false
		}

	case *TrafficPolicySpec_Policy_CorsPolicy_StringMatch_Regex:
		if _, ok := target.MatchType.(*TrafficPolicySpec_Policy_CorsPolicy_StringMatch_Regex); !ok {
			return false
		}

		if strings.Compare(m.GetRegex(), target.GetRegex()) != 0 {
			return false
		}

	default:
		// m is nil but target is not nil
		if m.MatchType != target.MatchType {
			return false
		}
	}

	return true
}

// Equal function
func (m *TrafficPolicySpec_Policy_MTLS_Istio) Equal(that interface{}) bool {
	if that == nil {
		return m == nil
	}

	target, ok := that.(*TrafficPolicySpec_Policy_MTLS_Istio)
	if !ok {
		that2, ok := that.(TrafficPolicySpec_Policy_MTLS_Istio)
		if ok {
			target = &that2
		} else {
			return false
		}
	}
	if target == nil {
		return m == nil
	} else if m == nil {
		return false
	}

	if m.GetTlsMode() != target.GetTlsMode() {
		return false
	}

	return true
}