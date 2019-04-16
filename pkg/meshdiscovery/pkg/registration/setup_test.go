package registration

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
	It("runs the registration event loop", func() {

		cs, err := clientset.ClientsetFromContext(ctx)
		Expect(err).NotTo(HaveOccurred())
		errHandler := func(err error) {
			Expect(err).NotTo(HaveOccurred())
		}

		mockSyncer := &mockRegistrationSyncer{}

		go func() {
			defer GinkgoRecover()
			err = runRegistrationEventLoop(ctx, errHandler, cs, v1.RegistrationSyncers{mockSyncer})
			Expect(err).NotTo(HaveOccurred())
		}()

		// create a mesh crd, ensure our sync gets called
		registration := &v1.Mesh{
			Metadata: core.Metadata{Name: "myregistration", Namespace: namespace},
			MeshType: &v1.Mesh_Istio{},
		}
		_, err = cs.Input.Mesh.Write(registration, clients.WriteOpts{})
		Expect(err).NotTo(HaveOccurred())

		Eventually(func() *v1.RegistrationSnapshot {
			return mockSyncer.received
		}, time.Second*5).Should(Not(BeNil()))
	})
})

type mockRegistrationSyncer struct {
	received *v1.RegistrationSnapshot
}

func (s *mockRegistrationSyncer) Sync(ctx context.Context, snap *v1.RegistrationSnapshot) error {
	s.received = snap
	return nil
}
