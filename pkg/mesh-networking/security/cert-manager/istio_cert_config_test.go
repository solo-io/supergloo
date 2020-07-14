package cert_manager_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/solo-io/go-utils/testutils"
	smh_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1/types"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	smh_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/types"
	smh_security_types "github.com/solo-io/service-mesh-hub/pkg/api/security.smh.solo.io/v1alpha1/types"
	cert_manager "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/security/cert-manager"
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
		mesh := &smh_discovery.Mesh{
			Spec: smh_discovery_types.MeshSpec{
				MeshType: &smh_discovery_types.MeshSpec_Linkerd{
					Linkerd: &smh_discovery_types.MeshSpec_LinkerdMesh{},
				},
			},
		}
		_, err := istioConfigProdcer.ConfigureCertificateInfo(nil, mesh)
		Expect(err).To(HaveOccurred())
		Expect(err).To(HaveInErrorChain(cert_manager.IncorrectMeshTypeError(mesh)))
	})

	It("will return default values if citadel info isn't discovered", func() {
		mesh := &smh_discovery.Mesh{
			Spec: smh_discovery_types.MeshSpec{
				MeshType: &smh_discovery_types.MeshSpec_Istio1_5_{
					Istio1_5: &smh_discovery_types.MeshSpec_Istio1_5{
						Metadata: &smh_discovery_types.MeshSpec_IstioMesh{
							Installation: &smh_discovery_types.MeshSpec_MeshInstallation{
								InstallationNamespace: istioNamespace,
							},
						},
					},
				},
			},
		}
		matchCertConfig := &smh_security_types.VirtualMeshCertificateSigningRequestSpec_CertConfig{
			Hosts: []string{
				cert_manager.BuildSpiffeURI(
					cert_manager.DefaultTrustDomain,
					istioNamespace,
					cert_manager.DefaultCitadelServiceAccount,
				),
			},
			Org:      cert_manager.DefaultIstioOrg,
			MeshType: smh_core_types.MeshType_ISTIO1_5,
		}
		certConfig, err := istioConfigProdcer.ConfigureCertificateInfo(nil, mesh)
		Expect(err).NotTo(HaveOccurred())
		Expect(certConfig).To(Equal(matchCertConfig))
	})

	It("will return default the values present in the cert config when discovered", func() {
		citadelInfo := &smh_discovery_types.MeshSpec_IstioMesh_CitadelInfo{
			TrustDomain:           "test.domain",
			CitadelNamespace:      "testns",
			CitadelServiceAccount: "testsa",
		}
		mesh := &smh_discovery.Mesh{
			Spec: smh_discovery_types.MeshSpec{
				MeshType: &smh_discovery_types.MeshSpec_Istio1_6_{
					Istio1_6: &smh_discovery_types.MeshSpec_Istio1_6{
						Metadata: &smh_discovery_types.MeshSpec_IstioMesh{
							Installation: &smh_discovery_types.MeshSpec_MeshInstallation{
								InstallationNamespace: istioNamespace,
							},
							CitadelInfo: citadelInfo,
						},
					},
				},
			},
		}
		matchCertConfig := &smh_security_types.VirtualMeshCertificateSigningRequestSpec_CertConfig{
			Hosts: []string{
				cert_manager.BuildSpiffeURI(
					citadelInfo.GetTrustDomain(),
					citadelInfo.GetCitadelNamespace(),
					citadelInfo.GetCitadelServiceAccount(),
				),
			},
			Org:      cert_manager.DefaultIstioOrg,
			MeshType: smh_core_types.MeshType_ISTIO1_6,
		}
		certConfig, err := istioConfigProdcer.ConfigureCertificateInfo(nil, mesh)
		Expect(err).NotTo(HaveOccurred())
		Expect(certConfig).To(Equal(matchCertConfig))
	})
})
