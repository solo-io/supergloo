package extensions_test

import (
	"context"
	"time"

	"go.uber.org/atomic"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.gloomesh.solo.io/extensions/v1alpha1"
	mock_extensions "github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/extensions/mocks"

	. "github.com/solo-io/gloo-mesh/pkg/mesh-networking/extensions"
)

var _ = Describe("Clientset", func() {
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

		client.EXPECT().WatchPushNotifications(ctx, &v1alpha1.WatchPushNotificationsRequest{}).Return(notificationStream, nil).AnyTimes()

		notificationStream.EXPECT().Recv().Return(&v1alpha1.PushNotification{}, nil).AnyTimes()

		counter := atomic.Int32{}
		err := Clients{client}.WatchPushNotifications(ctx, func(*v1alpha1.PushNotification) {
			counter.Inc()
		})
		Expect(err).NotTo(HaveOccurred())

		Eventually(func() int {
			return int(counter.Load())
		}).Should(BeNumerically(">", 50))

		// expect cancelled context to stop calling push func
		cancel()
		time.Sleep(time.Millisecond * 100)

		lastVal := int(counter.Load())

		Consistently(func() int {
			return int(counter.Load())
		}).Should(Equal(lastVal))
	})
})
