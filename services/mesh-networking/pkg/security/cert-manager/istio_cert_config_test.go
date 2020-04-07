package cert_manager_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/solo-io/go-utils/testutils"
	core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	security_types "github.com/solo-io/service-mesh-hub/pkg/api/security.zephyr.solo.io/v1alpha1/types"
	cert_manager "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/security/cert-manager"
)

var _ = Describe("istio cert config", func() {

	var (
		istioConfigProdcer cert_manager.CertConfigProducer

		istioNamespace = "istio-system"
	)

	BeforeEach(func() {
		istioConfigProdcer = cert_manager.NewIstioCertConfigProducer()
	})

	It("will fail if mesh is not type istio", func() {
		mesh := &v1alpha1.Mesh{
			Spec: discovery_types.MeshSpec{
				MeshType: &discovery_types.MeshSpec_Linkerd{
					Linkerd: &discovery_types.MeshSpec_LinkerdMesh{},
				},
			},
		}
		_, err := istioConfigProdcer.ConfigureCertificateInfo(nil, mesh)
		Expect(err).To(HaveOccurred())
		Expect(err).To(HaveInErrorChain(cert_manager.IncorrectMeshTypeError(mesh)))
	})

	It("will return default values if citadel info isn't discovered", func() {
		mesh := &v1alpha1.Mesh{
			Spec: discovery_types.MeshSpec{
				MeshType: &discovery_types.MeshSpec_Istio{
					Istio: &discovery_types.MeshSpec_IstioMesh{
						Installation: &discovery_types.MeshSpec_MeshInstallation{
							InstallationNamespace: istioNamespace,
						},
					},
				},
			},
		}
		matchCertConfig := &security_types.VirtualMeshCertificateSigningRequestSpec_CertConfig{
			Hosts: []string{
				cert_manager.BuildSpiffeURI(
					cert_manager.DefaultTrustDomain,
					istioNamespace,
					cert_manager.DefaultCitadelServiceAccount,
				),
			},
			Org:      cert_manager.DefaultIstioOrg,
			MeshType: core_types.MeshType_ISTIO,
		}
		certConfig, err := istioConfigProdcer.ConfigureCertificateInfo(nil, mesh)
		Expect(err).NotTo(HaveOccurred())
		Expect(certConfig).To(Equal(matchCertConfig))
	})

	It("will return default the values present in the cert config when discovered", func() {
		citadelInfo := &discovery_types.MeshSpec_IstioMesh_CitadelInfo{
			TrustDomain:           "test.domain",
			CitadelNamespace:      "testns",
			CitadelServiceAccount: "testsa",
		}
		mesh := &v1alpha1.Mesh{
			Spec: discovery_types.MeshSpec{
				MeshType: &discovery_types.MeshSpec_Istio{
					Istio: &discovery_types.MeshSpec_IstioMesh{
						Installation: &discovery_types.MeshSpec_MeshInstallation{
							InstallationNamespace: istioNamespace,
						},
						CitadelInfo: citadelInfo,
					},
				},
			},
		}
		matchCertConfig := &security_types.VirtualMeshCertificateSigningRequestSpec_CertConfig{
			Hosts: []string{
				cert_manager.BuildSpiffeURI(
					citadelInfo.GetTrustDomain(),
					citadelInfo.GetCitadelNamespace(),
					citadelInfo.GetCitadelServiceAccount(),
				),
			},
			Org:      cert_manager.DefaultIstioOrg,
			MeshType: core_types.MeshType_ISTIO,
		}
		certConfig, err := istioConfigProdcer.ConfigureCertificateInfo(nil, mesh)
		Expect(err).NotTo(HaveOccurred())
		Expect(certConfig).To(Equal(matchCertConfig))
	})
})
