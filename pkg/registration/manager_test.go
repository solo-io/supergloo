package registration

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("registration helpers", func() {
	var (
		manager *Manager
	)
	BeforeEach(func() {
		manager = NewManager()
	})

	Context("manager", func() {

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
			manager.publish(EnabledConfigLoops{})
			Eventually(func() int {
				return recievedUpdates
			}, time.Second*15, time.Second/2).Should(Equal(4))

		})

	})

	Context("listener", func() {

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

		})

	})

})
