package install_test

import (
	"context"
	"time"

	"github.com/solo-io/supergloo/cli/pkg/helpers"
	"github.com/solo-io/supergloo/cli/test/utils"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	skclients "github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/supergloo/cli/pkg/helpers/clients"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
)

var _ = Describe("Common install functions", func() {

	var (
		ctx           context.Context
		cancel        context.CancelFunc
		installClient v1.InstallClient
	)

	BeforeEach(func() {
		clients.UseMemoryClients()

		ctx, cancel = context.WithCancel(context.Background())
		installClient = clients.MustInstallClient()
	})

	AfterEach(func() {
		cancel()
	})

	Describe("waitUtilInstallAccepted", func() {

		var (
			err     error
			install = &v1.Install{
				Metadata: core.Metadata{
					Name:      "my-istio",
					Namespace: "my-namespace",
				},
				InstallationNamespace: "install-ns",
				Status: core.Status{
					State: core.Status_Pending,
				},
			}
		)

		BeforeEach(func() {
			install, err = installClient.Write(install, skclients.WriteOpts{})
			Expect(err).ToNot(HaveOccurred())
		})

		AfterEach(func() {
			err = installClient.Delete(install.Metadata.Namespace, install.Metadata.Name, skclients.DeleteOpts{})
			Expect(err).ToNot(HaveOccurred())
		})

		It("returns an error if install times out", func() {
			go func() {
				defer GinkgoRecover()
				utils.CompleteInstall(ctx, installClient, install.Metadata.Ref(), 1500*time.Millisecond)
			}()

			err = helpers.WaitForInstallStatus(ctx, install.Metadata.Ref(), core.Status_Accepted, time.Second)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(helpers.TimeoutError.Error()))
		})

		It("returns without error if install completes in time", func() {
			go func() {
				defer GinkgoRecover()
				utils.CompleteInstall(ctx, installClient, install.Metadata.Ref(), 50*time.Millisecond)
			}()

			err = helpers.WaitForInstallStatus(ctx, install.Metadata.Ref(), core.Status_Accepted, time.Second)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
