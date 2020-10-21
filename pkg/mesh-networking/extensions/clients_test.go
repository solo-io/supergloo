package extensions_test

import (
	"context"
	"time"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/extensions/v1alpha1"
	mock_extensions "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/extensions/mocks"

	. "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/extensions"
)

var _ = Describe("Clients", func() {
	var (
		ctl                *gomock.Controller
		client             *mock_extensions.MockNetworkingExtensionsClient
		notificationStream *mock_extensions.MockNetworkingExtensions_WatchPushNotificationsClient
		ctx                = context.TODO()
	)
	BeforeEach(func() {
		ctl = gomock.NewController(GinkgoT())
		client = mock_extensions.NewMockNetworkingExtensionsClient(ctl)
		notificationStream = mock_extensions.NewMockNetworkingExtensions_WatchPushNotificationsClient(ctl)
	})
	AfterEach(func() {
		ctl.Finish()
	})
	It("watches notifications and calls a push function on receipt of notification", func() {
		ctx, cancel := context.WithCancel(ctx)

		client.EXPECT().WatchPushNotifications(ctx, &v1alpha1.WatchPushNotificationsRequest{}).Return(notificationStream, nil)

		notificationStream.EXPECT().Recv().Return(&v1alpha1.PushNotification{}, nil).AnyTimes()

		var counter int
		err := Clients{client}.WatchPushNotifications(ctx, func() {
			counter++
		})
		Expect(err).NotTo(HaveOccurred())

		Eventually(func() int {
			return counter
		}).Should(BeNumerically(">", 50))

		// expect cancelled context to stop calling push func
		cancel()
		time.Sleep(time.Millisecond)

		lastVal := counter

		Consistently(func() int {
			return counter
		}).Should(Equal(lastVal))
	})
})
