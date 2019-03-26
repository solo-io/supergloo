package upstream_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	skclients "github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/supergloo/cli/pkg/helpers/clients"
	"github.com/solo-io/supergloo/cli/test/utils"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
)

var _ = Describe("edit upstream", func() {

	var (
		meshClient     v1.MeshClient
		upstreamClient gloov1.UpstreamClient

		ns   = "supergloo-system"
		name = "one"
	)

	BeforeEach(func() {
		clients.UseMemoryClients()
		meshClient = clients.MustMeshClient()
		upstreamClient = clients.MustUpstreamClient()
	})
	Context("validation", func() {
		It("returns err if name or namespace of mesh aren't present", func() {
			err := utils.Supergloo("set upstream mtls --mesh-namespace one")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("mesh resource name and namespace must be specified"))
		})

		It("returns err if name or namespace of upstream aren't present", func() {
			err := utils.Supergloo("set upstream mtls --namespace one --mesh-name one")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("upstream name and namespace must be specified"))
		})

		It("must be a valid mesh target", func() {
			mesh := &v1.Mesh{
				Metadata: core.Metadata{
					Name:      "two",
					Namespace: ns,
				},
			}
			_, err := meshClient.Write(mesh, skclients.WriteOpts{})
			Expect(err).NotTo(HaveOccurred())
			err = utils.Supergloo("set upstream mtls --name one --mesh-name one")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("unable to find mesh"))
		})
	})

	Context("successful edits", func() {
		BeforeEach(func() {
			mesh := &v1.Mesh{
				Metadata: core.Metadata{
					Name:      name,
					Namespace: ns,
				},
			}
			_, err := meshClient.Write(mesh, skclients.WriteOpts{})
			Expect(err).NotTo(HaveOccurred())
			us := &gloov1.Upstream{
				Metadata: core.Metadata{
					Name:      name,
					Namespace: ns,
				},
				UpstreamSpec: &gloov1.UpstreamSpec{},
			}
			_, err = upstreamClient.Write(us, skclients.WriteOpts{})
			Expect(err).NotTo(HaveOccurred())
		})
		It("can edit an upstream", func() {
			err := utils.Supergloo("set upstream mtls --name one --mesh-name one")
			Expect(err).NotTo(HaveOccurred())
			us, err := upstreamClient.Read(ns, name, skclients.ReadOpts{})
			Expect(err).NotTo(HaveOccurred())
			Expect(us.UpstreamSpec.SslConfig).NotTo(BeNil())
			sslConfig, ok := us.UpstreamSpec.SslConfig.SslSecrets.(*gloov1.UpstreamSslConfig_SslFiles)
			Expect(ok).To(BeTrue())
			Expect(sslConfig.SslFiles.RootCa).To(ContainSubstring(fmt.Sprintf("%s/%s", ns, name)))
			Expect(sslConfig.SslFiles.TlsKey).To(ContainSubstring(fmt.Sprintf("%s/%s", ns, name)))
			Expect(sslConfig.SslFiles.TlsCert).To(ContainSubstring(fmt.Sprintf("%s/%s", ns, name)))
		})
	})
})
