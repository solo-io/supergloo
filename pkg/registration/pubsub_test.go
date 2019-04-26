package registration

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/eventloop"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
)

var _ = Describe("registration helpers", func() {
	var (
		pubSub *PubSub
		ctx    context.Context
		cancel context.CancelFunc
	)
	BeforeEach(func() {
		pubSub = NewPubsub()
		ctx, cancel = context.WithCancel(context.TODO())
	})

	Context("pubsub", func() {

		It("Allows for multiple subscribers to be added and removed", func() {
			watches := make([]Receiver, 3)
			for i := 0; i < 4; i++ {
				watches = append(watches, pubSub.subscribe())
			}
			Expect(pubSub.subscriberCache).To(HaveLen(4))

			for _, watch := range watches {
				pubSub.unsubscribe(watch)
				Expect(pubSub.subscriberCache).NotTo(ContainElement(watch))
			}
		})

		It("send updates to all available recievers", func() {
			watches := make([]Receiver, 3)
			recievedUpdates := 0
			for i := 0; i < 4; i++ {
				reciever := pubSub.subscribe()
				watches = append(watches, reciever)
				go func() {
					<-reciever
					recievedUpdates++
				}()
			}
			pubSub.publish(ctx, EnabledConfigLoops{})
			Eventually(func() int {
				return recievedUpdates
			}, time.Second*15, time.Second/2).Should(Equal(4))

		})

	})

	Context("subscriber", func() {
		var (
			subscriber *Subscriber
			cl         *mockConfigLoop
		)

		BeforeEach(func() {
			cl = newMockConfigLoop()
			subscriber = NewSubscriber(ctx, pubSub, cl)
		})

		It("cancelling the parent context automatically unsubscribes the subscriber", func() {
			Expect(pubSub.subscriberCache).To(HaveLen(1))
			cancel()
			Eventually(func() int {
				return len(pubSub.subscriberCache)
			}, time.Second*10, time.Second/2).Should(Equal(0))
		})

		It("call enabled properly", func() {
			subscriber.Listen(ctx)
			pubSub.publish(ctx, EnabledConfigLoops{})
			Eventually(func() int {
				return len(cl.allLoops)
			}, time.Second*10, time.Second/2).Should(Equal(2))
			Expect(cl.startCalled).To(Equal(0))
			cancel()
		})

		It("calls start properly", func() {
			subscriber.Listen(ctx)
			pubSub.publish(ctx, EnabledConfigLoops{
				Istio: true,
			})
			Eventually(func() int {
				return cl.startCalled
			}, time.Second*10, time.Second/2).Should(Equal(1))
			pubSub.publish(ctx, EnabledConfigLoops{
				Istio: true,
			})
			pubSub.publish(ctx, EnabledConfigLoops{})
			Eventually(func() int {
				return cl.contextCancelled
			}, time.Second*10, time.Second/2).Should(Equal(1))
			pubSub.publish(ctx, EnabledConfigLoops{
				Istio: true,
			})
			Eventually(func() int {
				return cl.startCalled
			}, time.Second*10, time.Second/2).Should(Equal(2))
			cancel()
			Eventually(func() int {
				return cl.contextCancelled
			}, time.Second*10, time.Second/2).Should(Equal(2))
		})

	})

	Context("syncer", func() {
		var (
			subscriber *Subscriber
			cl         *mockConfigLoop
			syncer     *RegistrationSyncer
		)

		BeforeEach(func() {
			cl = newMockConfigLoop()
			subscriber = NewSubscriber(ctx, pubSub, cl)
			syncer = NewRegistrationSyncer(pubSub)
		})

		It("can handle a single mesh", func() {
			subscriber.Listen(ctx)
			snap := &v1.RegistrationSnapshot{
				Meshes: v1.MeshesByNamespace{
					"": v1.MeshList{
						{
							MeshType: &v1.Mesh_Istio{
								Istio: &v1.IstioMesh{},
							},
						},
					},
				},
			}
			err := syncer.Sync(ctx, snap)
			Expect(err).NotTo(HaveOccurred())

			Eventually(func() bool {
				if len(cl.allLoops) == 0 {
					return false
				}
				// use second as first is initial all false obj
				return cl.allLoops[1].Istio
			}, time.Second*10, time.Second/2).Should(BeTrue())
		})
		It("can handle a single mesh ingress", func() {
			subscriber.Listen(ctx)
			snap := &v1.RegistrationSnapshot{
				Meshingresses: v1.MeshingressesByNamespace{
					"": v1.MeshIngressList{
						{
							MeshIngressType: &v1.MeshIngress_Gloo{
								Gloo: &v1.GlooMeshIngress{},
							},
						},
					},
				},
			}
			err := syncer.Sync(ctx, snap)
			Expect(err).NotTo(HaveOccurred())

			Eventually(func() bool {
				if len(cl.allLoops) == 0 {
					return false
				}
				// use second as first is initial all false obj
				return cl.allLoops[1].Gloo
			}, time.Second*10, time.Second/2).Should(BeTrue())
		})

		It("can handle multiple subscribers", func() {
			subscriber2 := NewSubscriber(ctx, pubSub, cl)
			subscriber3 := NewSubscriber(ctx, pubSub, cl)
			subscriber.Listen(ctx)
			subscriber2.Listen(ctx)
			subscriber3.Listen(ctx)

			snap := &v1.RegistrationSnapshot{
				Meshingresses: v1.MeshingressesByNamespace{
					"": v1.MeshIngressList{
						{
							MeshIngressType: &v1.MeshIngress_Gloo{
								Gloo: &v1.GlooMeshIngress{},
							},
						},
					},
				},
				Meshes: v1.MeshesByNamespace{
					"": v1.MeshList{
						{
							MeshType: &v1.Mesh_Istio{
								Istio: &v1.IstioMesh{},
							},
						},
					},
				},
			}
			err := syncer.Sync(ctx, snap)
			Expect(err).NotTo(HaveOccurred())

			Eventually(func() int {
				return len(cl.allLoops)
			}, time.Second*10, time.Second/2).Should(Equal(6))
			cancel()

			Eventually(func() int {
				return cl.contextCancelled
			}, time.Second*10, time.Second/2).Should(Equal(3))
			Expect(cl.allLoops[5].Gloo).To(BeTrue())
			Expect(cl.allLoops[5].Istio).To(BeTrue())
			Expect(cl.allLoops[5].Linkerd).To(BeFalse())
			Expect(cl.allLoops[5].AppMesh).To(BeFalse())
		})
	})

})

type mockConfigLoop struct {
	startCalled      int
	contextCancelled int
	allLoops         []EnabledConfigLoops
}

func (mcl *mockConfigLoop) Enabled(enabled EnabledConfigLoops) bool {
	mcl.allLoops = append(mcl.allLoops, enabled)
	return enabled.Istio
}

func (mcl *mockConfigLoop) Start(ctx context.Context, enabled EnabledConfigLoops) (eventloop.EventLoop, error) {
	mcl.startCalled++
	go func() {
		<-ctx.Done()
		mcl.contextCancelled++
	}()
	return nil, nil
}

func newMockConfigLoop() *mockConfigLoop {
	return &mockConfigLoop{startCalled: 0}
}
