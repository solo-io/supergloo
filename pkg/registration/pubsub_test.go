package registration

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/eventloop"
)

var _ = Describe("registration helpers", func() {
	var (
		manager *PubSub
		ctx     context.Context
		cancel  context.CancelFunc
	)
	BeforeEach(func() {
		manager = NewPubsub()
		ctx, cancel = context.WithCancel(context.TODO())
	})

	Context("pubsub", func() {

		It("Allows for multiple subscribers to be added and removed", func() {
			watches := make([]Reciever, 3)
			for i := 0; i < 4; i++ {
				watches = append(watches, manager.Subscribe())
			}
			Expect(manager.subscriberCache).To(HaveLen(4))

			for _, watch := range watches {
				manager.Unsubscribe(watch)
				Expect(manager.subscriberCache).NotTo(ContainElement(watch))
			}
		})

		It("send updates to all available recievers", func() {
			watches := make([]Reciever, 3)
			recievedUpdates := 0
			for i := 0; i < 4; i++ {
				reciever := manager.Subscribe()
				watches = append(watches, reciever)
				go func() {
					<-reciever
					recievedUpdates++
				}()
			}
			manager.publish(ctx, EnabledConfigLoops{})
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
			subscriber = NewSubscriber(ctx, manager, cl)
		})

		It("cancelling the parent context automatically unsubscribes the subscriber", func() {
			Expect(manager.subscriberCache).To(HaveLen(1))
			cancel()
			Expect(manager.subscriberCache).To(HaveLen(0))
		})

		It("send updates to all available recievers", func() {

		})

	})

})

type mockConfigLoop struct {
}

func (*mockConfigLoop) Enabled(enabled EnabledConfigLoops) bool {
	return true
}

func (*mockConfigLoop) Start(ctx context.Context, enabled EnabledConfigLoops) (eventloop.EventLoop, error) {
	return nil, nil
}

func newMockConfigLoop() *mockConfigLoop {
	return &mockConfigLoop{}
}
