// Code generated by protoc-gen-ext. DO NOT EDIT.
// source: github.com/solo-io/gloo-mesh/api/certificates/v1/issued_certificate.proto

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
func (m *IssuedCertificateSpec) Equal(that interface{}) bool {
	if that == nil {
		return m == nil
	}

	target, ok := that.(*IssuedCertificateSpec)
	if !ok {
		that2, ok := that.(IssuedCertificateSpec)
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

	if len(m.GetHosts()) != len(target.GetHosts()) {
		return false
	}
	for idx, v := range m.GetHosts() {

		if strings.Compare(v, target.GetHosts()[idx]) != 0 {
			return false
		}

	}

	if strings.Compare(m.GetOrg(), target.GetOrg()) != 0 {
		return false
	}

	if h, ok := interface{}(m.GetIssuedCertificateSecret()).(equality.Equalizer); ok {
		if !h.Equal(target.GetIssuedCertificateSecret()) {
			return false
		}
	} else {
		if !proto.Equal(m.GetIssuedCertificateSecret(), target.GetIssuedCertificateSecret()) {
			return false
		}
	}

	if h, ok := interface{}(m.GetPodBounceDirective()).(equality.Equalizer); ok {
		if !h.Equal(target.GetPodBounceDirective()) {
			return false
		}
	} else {
		if !proto.Equal(m.GetPodBounceDirective(), target.GetPodBounceDirective()) {
			return false
		}
	}

	switch m.Signer.(type) {

	case *IssuedCertificateSpec_SigningCertificateSecret:
		if _, ok := target.Signer.(*IssuedCertificateSpec_SigningCertificateSecret); !ok {
			return false
		}

		if h, ok := interface{}(m.GetSigningCertificateSecret()).(equality.Equalizer); ok {
			if !h.Equal(target.GetSigningCertificateSecret()) {
				return false
			}
		} else {
			if !proto.Equal(m.GetSigningCertificateSecret(), target.GetSigningCertificateSecret()) {
				return false
			}
		}

	case *IssuedCertificateSpec_VaultCa:
		if _, ok := target.Signer.(*IssuedCertificateSpec_VaultCa); !ok {
			return false
		}

		if h, ok := interface{}(m.GetVaultCa()).(equality.Equalizer); ok {
			if !h.Equal(target.GetVaultCa()) {
				return false
			}
		} else {
			if !proto.Equal(m.GetVaultCa(), target.GetVaultCa()) {
				return false
			}
		}

	default:
		// m is nil but target is not nil
		if m.Signer != target.Signer {
			return false
		}
	}

	return true
}

// Equal function
func (m *IssuedCertificateStatus) Equal(that interface{}) bool {
	if that == nil {
		return m == nil
	}

	target, ok := that.(*IssuedCertificateStatus)
	if !ok {
		that2, ok := that.(IssuedCertificateStatus)
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

	if strings.Compare(m.GetError(), target.GetError()) != 0 {
		return false
	}

	if m.GetState() != target.GetState() {
		return false
	}

	return true
}
