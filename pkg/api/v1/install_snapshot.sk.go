// Code generated by solo-kit. DO NOT EDIT.

package v1

import (
	encryption_istio_io "github.com/solo-io/supergloo/pkg/api/external/istio/encryption/v1"

	"github.com/solo-io/solo-kit/pkg/utils/hashutils"
	"go.uber.org/zap"
)

type InstallSnapshot struct {
	Istiocerts encryption_istio_io.IstiocertsByNamespace
	Installs   InstallsByNamespace
}

func (s InstallSnapshot) Clone() InstallSnapshot {
	return InstallSnapshot{
		Istiocerts: s.Istiocerts.Clone(),
		Installs:   s.Installs.Clone(),
	}
}

func (s InstallSnapshot) Hash() uint64 {
	return hashutils.HashAll(
		s.hashIstiocerts(),
		s.hashInstalls(),
	)
}

func (s InstallSnapshot) hashIstiocerts() uint64 {
	return hashutils.HashAll(s.Istiocerts.List().AsInterfaces()...)
}

func (s InstallSnapshot) hashInstalls() uint64 {
	return hashutils.HashAll(s.Installs.List().AsInterfaces()...)
}

func (s InstallSnapshot) HashFields() []zap.Field {
	var fields []zap.Field
	fields = append(fields, zap.Uint64("istiocerts", s.hashIstiocerts()))
	fields = append(fields, zap.Uint64("installs", s.hashInstalls()))

	return append(fields, zap.Uint64("snapshotHash", s.Hash()))
}
