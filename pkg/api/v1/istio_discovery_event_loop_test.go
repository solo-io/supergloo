// Code generated by solo-kit. DO NOT EDIT.

// +build solokit

package v1

import (
	"context"
	"sync"
	"time"

	istio_authentication_v1alpha1 "github.com/solo-io/supergloo/pkg/api/external/istio/authorization/v1alpha1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/memory"
)

var _ = Describe("IstioDiscoveryEventLoop", func() {
	var (
		namespace string
		emitter   IstioDiscoveryEmitter
		err       error
	)

	BeforeEach(func() {

		podClientFactory := &factory.MemoryResourceClientFactory{
			Cache: memory.NewInMemoryResourceCache(),
		}
		podClient, err := NewPodClient(podClientFactory)
		Expect(err).NotTo(HaveOccurred())

		meshClientFactory := &factory.MemoryResourceClientFactory{
			Cache: memory.NewInMemoryResourceCache(),
		}
		meshClient, err := NewMeshClient(meshClientFactory)
		Expect(err).NotTo(HaveOccurred())

		installClientFactory := &factory.MemoryResourceClientFactory{
			Cache: memory.NewInMemoryResourceCache(),
		}
		installClient, err := NewInstallClient(installClientFactory)
		Expect(err).NotTo(HaveOccurred())

		meshPolicyClientFactory := &factory.MemoryResourceClientFactory{
			Cache: memory.NewInMemoryResourceCache(),
		}
		meshPolicyClient, err := istio_authentication_v1alpha1.NewMeshPolicyClient(meshPolicyClientFactory)
		Expect(err).NotTo(HaveOccurred())

		emitter = NewIstioDiscoveryEmitter(podClient, meshClient, installClient, meshPolicyClient)
	})
	It("runs sync function on a new snapshot", func() {
		_, err = emitter.Pod().Write(NewPod(namespace, "jerry"), clients.WriteOpts{})
		Expect(err).NotTo(HaveOccurred())
		_, err = emitter.Mesh().Write(NewMesh(namespace, "jerry"), clients.WriteOpts{})
		Expect(err).NotTo(HaveOccurred())
		_, err = emitter.Install().Write(NewInstall(namespace, "jerry"), clients.WriteOpts{})
		Expect(err).NotTo(HaveOccurred())
		_, err = emitter.MeshPolicy().Write(istio_authentication_v1alpha1.NewMeshPolicy(namespace, "jerry"), clients.WriteOpts{})
		Expect(err).NotTo(HaveOccurred())
		sync := &mockIstioDiscoverySyncer{}
		el := NewIstioDiscoveryEventLoop(emitter, sync)
		_, err := el.Run([]string{namespace}, clients.WatchOpts{})
		Expect(err).NotTo(HaveOccurred())
		Eventually(sync.Synced, 5*time.Second).Should(BeTrue())
	})
})

type mockIstioDiscoverySyncer struct {
	synced bool
	mutex  sync.Mutex
}

func (s *mockIstioDiscoverySyncer) Synced() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.synced
}

func (s *mockIstioDiscoverySyncer) Sync(ctx context.Context, snap *IstioDiscoverySnapshot) error {
	s.mutex.Lock()
	s.synced = true
	s.mutex.Unlock()
	return nil
}
