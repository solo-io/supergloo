// Code generated by protoc-gen-ext. DO NOT EDIT.
// source: github.com/solo-io/gloo-mesh/api/enterprise/networking/v1beta1/extauth_server_config.proto

package v1beta1

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
func (m *ExtauthServerConfigSpec) Equal(that interface{}) bool {
	if that == nil {
		return m == nil
	}

	target, ok := that.(*ExtauthServerConfigSpec)
	if !ok {
		that2, ok := that.(ExtauthServerConfigSpec)
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

	if len(m.GetServerConfigRefs()) != len(target.GetServerConfigRefs()) {
		return false
	}
	for idx, v := range m.GetServerConfigRefs() {

		if h, ok := interface{}(v).(equality.Equalizer); ok {
			if !h.Equal(target.GetServerConfigRefs()[idx]) {
				return false
			}
		} else {
			if !proto.Equal(v, target.GetServerConfigRefs()[idx]) {
				return false
			}
		}

	}

	if h, ok := interface{}(m.GetExtauthConfig()).(equality.Equalizer); ok {
		if !h.Equal(target.GetExtauthConfig()) {
			return false
		}
	} else {
		if !proto.Equal(m.GetExtauthConfig(), target.GetExtauthConfig()) {
			return false
		}
	}

	return true
}

// Equal function
func (m *ExtauthServerConfigStatus) Equal(that interface{}) bool {
	if that == nil {
		return m == nil
	}

	target, ok := that.(*ExtauthServerConfigStatus)
	if !ok {
		that2, ok := that.(ExtauthServerConfigStatus)
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

	if len(m.GetErrors()) != len(target.GetErrors()) {
		return false
	}
	for idx, v := range m.GetErrors() {

		if strings.Compare(v, target.GetErrors()[idx]) != 0 {
			return false
		}

	}

	if len(m.GetWarnings()) != len(target.GetWarnings()) {
		return false
	}
	for idx, v := range m.GetWarnings() {

		if strings.Compare(v, target.GetWarnings()[idx]) != 0 {
			return false
		}

	}

	if len(m.GetConfiguredServers()) != len(target.GetConfiguredServers()) {
		return false
	}
	for idx, v := range m.GetConfiguredServers() {

		if h, ok := interface{}(v).(equality.Equalizer); ok {
			if !h.Equal(target.GetConfiguredServers()[idx]) {
				return false
			}
		} else {
			if !proto.Equal(v, target.GetConfiguredServers()[idx]) {
				return false
			}
		}

	}

	return true
}
