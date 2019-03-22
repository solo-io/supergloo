package setup

import (
	"context"
	"time"

	"github.com/solo-io/supergloo/pkg/api/clientset"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/testutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/solo-kit/test/setup"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"

	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
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

		cs, err := clientset.ClientsetFromContext(ctx)
		Expect(err).NotTo(HaveOccurred())
		errHandler := func(err error) {
			defer GinkgoRecover()
			Expect(err).NotTo(HaveOccurred())
		}

		mockSyncer := &mockInstallSyncer{}

		go func() {
			defer GinkgoRecover()
			err = runConfigEventLoop(ctx, cs, errHandler, []v1.ConfigSyncer{mockSyncer})
			Expect(err).NotTo(HaveOccurred())
		}()

		// create an install crd, ensure our sync gets called
		install := &v1.Install{
			Metadata: core.Metadata{Name: "myinstall", Namespace: namespace},
		}
		_, err = cs.Input.Install.Write(install, clients.WriteOpts{})
		Expect(err).NotTo(HaveOccurred())

		Eventually(func() *v1.ConfigSnapshot {
			return mockSyncer.received
		}, time.Second*5).Should(Not(BeNil()))
	})
})

type mockInstallSyncer struct {
	received *v1.ConfigSnapshot
}

func (s *mockInstallSyncer) Sync(ctx context.Context, snap *v1.ConfigSnapshot) error {
	s.received = snap
	return nil
}
