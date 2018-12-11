package util

import (
	"github.com/gogo/protobuf/types"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/supergloo/pkg/api/v1"
)

func GetRef(namespace, name string) *core.ResourceRef {
	return &core.ResourceRef{
		Namespace: namespace,
		Name:      name,
	}
}

func GetEncryption(mtls *types.BoolValue, ref *core.ResourceRef) *v1.Encryption {
	encryption := &v1.Encryption{}
	if mtls == nil {
		return encryption
	}
	encryption.TlsEnabled = mtls.Value
	if ref == nil {
		return encryption
	}
	encryption.Secret = ref
	return encryption
}

func GetInstallFromEnc(encryption *v1.Encryption) *v1.Install {
	return &v1.Install{
		Encryption: encryption,
	}
}

func GetInstall(mtls *types.BoolValue, ref *core.ResourceRef) *v1.Install {
	return GetInstallFromEnc(GetEncryption(mtls, ref))
}
