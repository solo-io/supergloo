package util

import (
	"github.com/gogo/protobuf/types"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	istiov1 "github.com/solo-io/supergloo/pkg/api/external/istio/encryption/v1"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/constants"
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

const (
	SecretNamespace       = "foo"
	SecretNameMissing     = "missing"
	SecretNameMissingRoot = "missing_root"
	SecretNameMissingKey  = "missing_key"
	SecretNameValid       = "valid"
)

func GetInstall(mtls *types.BoolValue, ref *core.ResourceRef) *v1.Install {
	return GetInstallFromEnc(GetEncryption(mtls, ref))
}

func GetIstioSecret(namespace string, name string) *istiov1.IstioCacertsSecret {
	return &istiov1.IstioCacertsSecret{
		Metadata: core.Metadata{
			Namespace: namespace,
			Name:      name,
		},
	}
}

func GetTestSecrets() istiov1.IstioCacertsSecretList {
	// NOTE: the ordering here matters, should be alphabetical by secret name
	var list istiov1.IstioCacertsSecretList
	missingKey := GetIstioSecret(SecretNamespace, SecretNameMissingKey)
	missingKey.RootCert = "root"
	list = append(list, missingKey)
	missingRoot := GetIstioSecret(SecretNamespace, SecretNameMissingRoot)
	list = append(list, missingRoot)
	valid := GetIstioSecret(SecretNamespace, SecretNameValid)
	valid.RootCert = "root"
	valid.CaKey = "key"
	list = append(list, valid)
	return list
}

func GetInstallWithoutMeshType(path string, meshName string, install bool) *v1.Install {
	return &v1.Install{
		Metadata: core.Metadata{
			Namespace: constants.SuperglooNamespace,
			Name:      meshName,
		},
		ChartLocator: &v1.HelmChartLocator{
			Kind: &v1.HelmChartLocator_ChartPath{
				ChartPath: &v1.HelmChartPath{
					Path: path,
				},
			},
		},
		Enabled: &types.BoolValue{
			Value: install,
		},
	}
}

func GetSnapshot(install *v1.Install) *v1.InstallSnapshot {
	return &v1.InstallSnapshot{
		Installs: v1.InstallsByNamespace{
			constants.SuperglooNamespace: v1.InstallList{
				install,
			},
		},
		Istiocerts: istiov1.IstiocertsByNamespace{
			SecretNamespace: GetTestSecrets(),
		},
	}
}
