package istio_test

import (
	"context"
	"testing"

	"github.com/gogo/protobuf/types"
	"github.com/golang/mock/gomock"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/supergloo/mock/pkg/kube"
	"github.com/solo-io/supergloo/mock/pkg/secret"
	"github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/install/istio"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var T *testing.T

func TestSecret(t *testing.T) {
	RegisterFailHandler(Fail)
	T = t
	RunSpecs(t, "Shared Suite")
}

var _ = Describe("Istio Installer", func() {

	var (
		installer istio.IstioInstaller
	)

	BeforeEach(func() {
		ctrl := gomock.NewController(T)
		defer ctrl.Finish()

		mockCrdClient := mock_kube.NewMockCrdClient(ctrl)
		mockSecurityClient := mock_kube.NewMockSecurityClient(ctrl)
		mockSecretSyncer := mock_secret.NewMockSecretSyncer(ctrl)

		installer, err := istio.NewIstioInstaller(context.TODO(), mockCrdClient, mockSecurityClient, mockSecretSyncer)
		Expect(err).To(BeNil())
	})

	getRef := func(namespace, name string) *core.ResourceRef {
		return &core.ResourceRef{
			Namespace: namespace,
			Name:      name,
		}
	}

	getEncryption := func(mtls *types.BoolValue, ref *core.ResourceRef) *v1.Encryption {
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

	getInstallFromEnc := func(encryption *v1.Encryption) *v1.Install {
		return &v1.Install{
			Encryption: encryption,
		}
	}

	getInstall := func(mtls *types.BoolValue, ref *core.ResourceRef) *v1.Install {
		return getInstallFromEnc(getEncryption(mtls, ref))
	}

	getYaml := func(install *v1.Install) string {
		installer.GetOverridesYaml(install)
	}

	Describe("Should get correct overrides", func() {
		It("nil encryption", func() {
			yaml := getYaml(nil, nil)
		})
	})

})
