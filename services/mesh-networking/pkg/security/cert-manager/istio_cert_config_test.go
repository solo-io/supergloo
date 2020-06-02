package cert_manager_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/solo-io/go-utils/testutils"
	zephyr_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	zephyr_security_types "github.com/solo-io/service-mesh-hub/pkg/api/security.zephyr.solo.io/v1alpha1/types"
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
		mesh := &zephyr_discovery.Mesh{
			Spec: zephyr_discovery_types.MeshSpec{
				MeshType: &zephyr_discovery_types.MeshSpec_Linkerd{
					Linkerd: &zephyr_discovery_types.MeshSpec_LinkerdMesh{},
				},
			},
		}
		_, err := istioConfigProdcer.ConfigureCertificateInfo(nil, mesh)
		Expect(err).To(HaveOccurred())
		Expect(err).To(HaveInErrorChain(cert_manager.IncorrectMeshTypeError(mesh)))
	})

	It("will return default values if citadel info isn't discovered", func() {
		mesh := &zephyr_discovery.Mesh{
			Spec: zephyr_discovery_types.MeshSpec{
				MeshType: &zephyr_discovery_types.MeshSpec_Istio1_5_{
					Istio1_5: &zephyr_discovery_types.MeshSpec_Istio1_5{
						Metadata: &zephyr_discovery_types.MeshSpec_IstioMesh{
							Installation: &zephyr_discovery_types.MeshSpec_MeshInstallation{
								InstallationNamespace: istioNamespace,
							},
						},
					},
				},
			},
		}
		matchCertConfig := &zephyr_security_types.VirtualMeshCertificateSigningRequestSpec_CertConfig{
			Hosts: []string{
				cert_manager.BuildSpiffeURI(
					cert_manager.DefaultTrustDomain,
					istioNamespace,
					cert_manager.DefaultCitadelServiceAccount,
				),
			},
			Org:      cert_manager.DefaultIstioOrg,
			MeshType: zephyr_core_types.MeshType_ISTIO1_5,
		}
		certConfig, err := istioConfigProdcer.ConfigureCertificateInfo(nil, mesh)
		Expect(err).NotTo(HaveOccurred())
		Expect(certConfig).To(Equal(matchCertConfig))
	})

	It("will return default the values present in the cert config when discovered", func() {
		citadelInfo := &zephyr_discovery_types.MeshSpec_IstioMesh_CitadelInfo{
			TrustDomain:           "test.domain",
			CitadelNamespace:      "testns",
			CitadelServiceAccount: "testsa",
		}
		mesh := &zephyr_discovery.Mesh{
			Spec: zephyr_discovery_types.MeshSpec{
				MeshType: &zephyr_discovery_types.MeshSpec_Istio1_6_{
					Istio1_6: &zephyr_discovery_types.MeshSpec_Istio1_6{
						Metadata: &zephyr_discovery_types.MeshSpec_IstioMesh{
							Installation: &zephyr_discovery_types.MeshSpec_MeshInstallation{
								InstallationNamespace: istioNamespace,
							},
							CitadelInfo: citadelInfo,
						},
					},
				},
			},
		}
		matchCertConfig := &zephyr_security_types.VirtualMeshCertificateSigningRequestSpec_CertConfig{
			Hosts: []string{
				cert_manager.BuildSpiffeURI(
					citadelInfo.GetTrustDomain(),
					citadelInfo.GetCitadelNamespace(),
					citadelInfo.GetCitadelServiceAccount(),
				),
			},
			Org:      cert_manager.DefaultIstioOrg,
			MeshType: zephyr_core_types.MeshType_ISTIO1_6,
		}
		certConfig, err := istioConfigProdcer.ConfigureCertificateInfo(nil, mesh)
		Expect(err).NotTo(HaveOccurred())
		Expect(certConfig).To(Equal(matchCertConfig))
	})
})
