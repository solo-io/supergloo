package setup

import (
	"context"
	"time"

	"github.com/solo-io/solo-kit/pkg/api/v1/eventloop"
	"github.com/solo-io/supergloo/pkg/api/clientset"
	"github.com/solo-io/supergloo/pkg/registration"

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
	It("runs the config event loop", func() {

		cs, err := clientset.ClientsetFromContext(ctx)
		Expect(err).NotTo(HaveOccurred())

		mockSyncer := &mockConfigSyncer{}

		configEmitter := v1.NewConfigEmitter(
			cs.Supergloo.Mesh,
			cs.Supergloo.MeshIngress,
			cs.Supergloo.MeshGroup,
			cs.Supergloo.RoutingRule,
			cs.Supergloo.SecurityRule,
			cs.Supergloo.TlsSecret,
			cs.Supergloo.Upstream,
			cs.Discovery.Pod,
		)

		el := v1.NewConfigEventLoop(configEmitter, mockSyncer)
		var runner = func(ctx context.Context, enabled registration.EnabledConfigLoops) (eventloop.EventLoop, error) {
			return el, nil
		}

		go func() {
			defer GinkgoRecover()
			err = registration.RunConfigLoop(ctx, registration.EnabledConfigLoops{Istio: true}, runner)
			Expect(err).NotTo(HaveOccurred())
		}()

		// create an install crd, ensure our sync gets called
		mesh := &v1.Mesh{
			Metadata: core.Metadata{Name: "myinstall", Namespace: namespace},
		}
		_, err = cs.Supergloo.Mesh.Write(mesh, clients.WriteOpts{})
		Expect(err).NotTo(HaveOccurred())

		Eventually(func() *v1.ConfigSnapshot {
			return mockSyncer.received
		}, time.Second*5, time.Second/2).Should(Not(BeNil()))
	})
})

type mockConfigSyncer struct {
	received *v1.ConfigSnapshot
}

func (s *mockConfigSyncer) Sync(ctx context.Context, snap *v1.ConfigSnapshot) error {
	s.received = snap
	return nil
}
