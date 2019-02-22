package setup

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/testutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/solo-kit/test/setup"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
)

var _ = Describe("Setup", func() {
	var (
		namespace string
		ctx       context.Context
		cancel    func()
	)
	BeforeEach(func() {
		namespace = "a" + testutils.RandString(6)
		err := setup.SetupKubeForTest(namespace)
		Expect(err).NotTo(HaveOccurred())
		ctx, cancel = context.WithCancel(context.TODO())
	})
	AfterEach(func() {
		setup.TeardownKube(namespace)
		cancel()
	})
	It("runs the install event loop", func() {

		cs, err := createClients(ctx)
		Expect(err).NotTo(HaveOccurred())
		errHandler := func(err error) {
			Expect(err).NotTo(HaveOccurred())
		}

		mockSyncer := &mockInstallSyncer{}

		go func() {
			defer GinkgoRecover()
			err = runInstallEventLoop(ctx, errHandler, cs, v1.InstallSyncers{mockSyncer})
			Expect(err).NotTo(HaveOccurred())
		}()

		// create an install crd, ensure our sync gets called
		install := &v1.Install{
			Metadata: core.Metadata{Name: "myinstall", Namespace: namespace},
		}
		_, err = cs.InstallClient.Write(install, clients.WriteOpts{})
		Expect(err).NotTo(HaveOccurred())

		Eventually(func() *v1.InstallSnapshot {
			return mockSyncer.received
		}, time.Second*5).Should(Not(BeNil()))
	})
})

type mockInstallSyncer struct {
	received *v1.InstallSnapshot
}

func (s *mockInstallSyncer) Sync(ctx context.Context, snap *v1.InstallSnapshot) error {
	s.received = snap
	return nil
}
