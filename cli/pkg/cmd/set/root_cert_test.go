package set_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/supergloo/cli/pkg/helpers"
	"github.com/solo-io/supergloo/cli/test/utils"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
)

var _ = Describe("RootCert", func() {
	mesh := core.Metadata{Namespace: "my", Name: "mesh"}
	secret := core.Metadata{Namespace: "my", Name: "secret"}
	BeforeEach(func() {
		helpers.UseMemoryClients()
		helpers.MustMeshClient().Write(&v1.Mesh{
			Metadata: mesh,
			MeshType: &v1.Mesh_Istio{Istio: &v1.IstioMesh{}},
		}, clients.WriteOpts{})
		helpers.MustTlsSecretClient().Write(&v1.TlsSecret{Metadata: secret}, clients.WriteOpts{})
	})

	It("updates the root cert ref on an existing mesh", func() {
		err := utils.Supergloo(fmt.Sprintf("set rootcert --target-mesh "+
			"%v.%v --tls-secret %v.%v", mesh.Namespace, mesh.Name, secret.Namespace, secret.Name))
		Expect(err).NotTo(HaveOccurred())
		meshWithCert, err := helpers.MustMeshClient().Read(mesh.Namespace, mesh.Name, clients.ReadOpts{})
		Expect(err).NotTo(HaveOccurred())
		Expect(meshWithCert.MtlsConfig).NotTo(BeNil())
		Expect(meshWithCert.MtlsConfig.MtlsEnabled).To(BeTrue())
		Expect(meshWithCert.MtlsConfig.RootCertificate).NotTo(BeNil())
		Expect(*meshWithCert.MtlsConfig.RootCertificate).To(Equal(secret.Ref()))
	})

	It("sets the root cert to nil on an existing mesh if no tls secret provided", func() {
		err := utils.Supergloo(fmt.Sprintf("set rootcert --target-mesh "+
			"%v.%v", mesh.Namespace, mesh.Name))
		Expect(err).NotTo(HaveOccurred())
		meshWithCert, err := helpers.MustMeshClient().Read(mesh.Namespace, mesh.Name, clients.ReadOpts{})
		Expect(err).NotTo(HaveOccurred())
		Expect(meshWithCert.MtlsConfig).NotTo(BeNil())
		Expect(meshWithCert.MtlsConfig.MtlsEnabled).To(BeFalse()) // persists whatever was in storage
		Expect(meshWithCert.MtlsConfig.RootCertificate).To(BeNil())
	})
})
