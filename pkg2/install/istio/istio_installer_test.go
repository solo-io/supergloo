package istio_test

import (
	"context"
	"testing"

	"github.com/solo-io/supergloo/test/util"

	"github.com/ghodss/yaml"
	"github.com/gogo/protobuf/types"
	"github.com/pkg/errors"
	"github.com/solo-io/supergloo/pkg/kube"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"

	"github.com/golang/mock/gomock"
	mock_kube "github.com/solo-io/supergloo/mock/pkg/kube"
	mock_secret "github.com/solo-io/supergloo/mock/pkg/secret"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/install/istio"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var T *testing.T

func TestIstioInstaller(t *testing.T) {
	RegisterFailHandler(Fail)
	T = t
	RunSpecs(t, "Shared Suite")
}

var _ = Describe("Istio Installer", func() {

	var (
		installer        *istio.IstioInstaller
		err              error
		mockCrdClient    *mock_kube.MockCrdClient
		mockSecretSyncer *mock_secret.MockSecretSyncer
	)

	BeforeEach(func() {
		ctrl := gomock.NewController(T)
		defer ctrl.Finish()

		mockCrdClient = mock_kube.NewMockCrdClient(ctrl)
		mockSecretSyncer = mock_secret.NewMockSecretSyncer(ctrl)

		installer, err = istio.NewIstioInstaller(mockCrdClient, nil, mockSecretSyncer)
		Expect(err).To(BeNil())
	})

	Describe("Should get correct overrides", func() {

		getYaml := func(install *v1.Install) string {
			return installer.GetOverridesYaml(install)
		}

		type MtlsOverrides struct {
			Enabled bool `json:"enabled"`
		}

		type GlobalOverrides struct {
			Mtls                        MtlsOverrides `json:"mtls"`
			Crds                        bool          `json:"crds"`
			ControlPlaneSecurityEnabled bool          `json:"controlPlaneSecurityEnabled"`
		}

		type SecurityOverrides struct {
			Enabled    bool `json:"enabled"`
			SelfSigned bool `json:"selfSigned"`
		}

		type Overrides struct {
			Global   GlobalOverrides   `json:"global"` // Affects YAML field names too.
			Security SecurityOverrides `json:"security"`
		}

		getExpectedOverrides := func(mtls, selfSigned bool) *Overrides {
			return &Overrides{
				Global: GlobalOverrides{
					ControlPlaneSecurityEnabled: true,
					Crds:                        false,
					Mtls: MtlsOverrides{
						Enabled: mtls,
					},
				},
				Security: SecurityOverrides{
					Enabled:    true,
					SelfSigned: selfSigned,
				},
			}
		}

		getActualOverrides := func(install *v1.Install) *Overrides {
			yamlStr := getYaml(install)
			yamlBytes := []byte(yamlStr)
			overrides := &Overrides{}
			err := yaml.Unmarshal(yamlBytes, overrides)
			Expect(err).To(BeNil())
			return overrides
		}

		It("nil encryption", func() {
			actual := getActualOverrides(util.GetInstallFromEnc(nil))
			expected := getExpectedOverrides(false, true)
			Expect(actual).To(BeEquivalentTo(expected))
		})

		It("empty encryption", func() {
			actual := getActualOverrides(util.GetInstall(nil, nil))
			expected := getExpectedOverrides(false, true)
			Expect(actual).To(BeEquivalentTo(expected))
		})

		It("false mtls nil secret", func() {
			actual := getActualOverrides(util.GetInstall(&types.BoolValue{Value: false}, nil))
			expected := getExpectedOverrides(false, true)
			Expect(actual).To(BeEquivalentTo(expected))
		})

		It("false mtls secret is ignored", func() {
			actual := getActualOverrides(util.GetInstall(&types.BoolValue{Value: false}, util.GetRef("foo", "bar")))
			expected := getExpectedOverrides(false, true)
			Expect(actual).To(BeEquivalentTo(expected))
		})

		It("true mtls nil secret", func() {
			actual := getActualOverrides(util.GetInstall(&types.BoolValue{Value: true}, nil))
			expected := getExpectedOverrides(true, true)
			Expect(actual).To(BeEquivalentTo(expected))
		})

		It("true mtls with secret", func() {
			actual := getActualOverrides(util.GetInstall(&types.BoolValue{Value: true}, util.GetRef("foo", "bar")))
			expected := getExpectedOverrides(true, false)
			Expect(actual).To(BeEquivalentTo(expected))
		})
	})

	Describe("do helm pre install", func() {

		installNamespace := "foo"
		testError := errors.Errorf("error")
		var encryption *v1.Encryption
		ctx := context.TODO()
		secretList := util.GetTestSecrets()

		getCrds := func() []*v1beta1.CustomResourceDefinition {
			crds, err := kube.CrdsFromManifest(istio.IstioCrdYaml)
			Expect(err).To(BeNil())
			return crds
		}

		It("error crd client", func() {
			mockCrdClient.EXPECT().CreateCrds(getCrds()).Return(testError)
			actual := installer.DoPreHelmInstall(ctx, installNamespace, util.GetInstallFromEnc(encryption), secretList)
			Expect(actual.Error()).Should(ContainSubstring("creating istio crds"))
		})

		It("error secret syncer", func() {
			mockCrdClient.EXPECT().CreateCrds(getCrds()).Return(nil)
			mockSecretSyncer.EXPECT().SyncSecret(ctx, installNamespace, encryption, secretList, true).Return(testError)
			actual := installer.DoPreHelmInstall(ctx, installNamespace, util.GetInstallFromEnc(encryption), secretList)
			Expect(actual.Error()).Should(ContainSubstring("syncing secret"))
		})

		It("succeeds", func() {
			mockCrdClient.EXPECT().CreateCrds(getCrds()).Return(nil)
			mockSecretSyncer.EXPECT().SyncSecret(ctx, installNamespace, encryption, secretList, true).Return(nil)
			actual := installer.DoPreHelmInstall(ctx, installNamespace, util.GetInstallFromEnc(encryption), secretList)
			Expect(actual).Should(BeNil())
		})
	})

	Describe("expected hard-coded values", func() {
		It("should match crb name", func() {
			Expect(installer.GetCrbName()).To(BeEquivalentTo(istio.CrbName))
		})

		It("should match default namespace", func() {
			Expect(installer.GetDefaultNamespace()).To(BeEquivalentTo(istio.DefaultNamespace))
		})
	})

})
