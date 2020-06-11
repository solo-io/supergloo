package mc_manager_test

import (
	"context"
	"time"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	. "github.com/solo-io/go-utils/testutils"
	. "github.com/solo-io/service-mesh-hub/pkg/common/compute-target/k8s"
	mock_controller_runtime "github.com/solo-io/service-mesh-hub/test/mocks/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

var _ = Describe("multi cluster manager", func() {
	var (
		ctrl  *gomock.Controller
		mgr   *mock_controller_runtime.MockManager
		cache *mock_controller_runtime.MockCache
		ctx   context.Context
		async AsyncManager
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mgr = mock_controller_runtime.NewMockManager(ctrl)
		cache = mock_controller_runtime.NewMockCache(ctrl)
		ctx = context.TODO()
		async = NewAsyncManager(ctx, mgr)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("can return manager properly", func() {
		Expect(async.Manager()).To(Equal(mgr))
	})

	It("can stop", func() {
		async.Stop()
		_, ok := <-async.Context().Done()
		Expect(ok).To(BeFalse())
		Expect(async.Context().Err()).To(Equal(context.Canceled))
	})

	Context("start", func() {
		var (
			testErr     = eris.New("hello")
			fakeOptions = func(innerCtx context.Context, innerMgr manager.Manager) error {
				Expect(innerMgr).To(Equal(mgr))
				return testErr
			}
		)
		It("will signal back an error if start fails", func() {
			mgr.EXPECT().GetCache().Return(cache)
			mgr.EXPECT().Start(async.Context().Done()).Return(testErr)
			cache.EXPECT().WaitForCacheSync(gomock.Any()).Return(true)
			Expect(async.Start()).NotTo(HaveOccurred())
			Eventually(func() error { return async.Error() }, time.Second*1).
				Should(HaveInErrorChain(ManagerStartError(eris.New("hello"))))
			_, ok := <-async.GotError()
			Expect(ok).To(BeFalse())
		})

		It("will return an error if wait for cache sync returns false", func() {
			mgr.EXPECT().GetCache().Return(cache)
			mgr.EXPECT().Start(async.Context().Done()).Return(nil).AnyTimes()
			cache.EXPECT().WaitForCacheSync(gomock.Any()).Return(false)
			Eventually(func() error { return async.Error() }, time.Second*1).ShouldNot(HaveOccurred())
			err := async.Start()
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(ManagerCacheSyncError))
		})

		It("will return an error if a pre start fails", func() {
			mgr.EXPECT().GetCache().Return(cache)
			// must be set to AnyTimes as running with -p doesn't garauntee the goroutine runs in time
			mgr.EXPECT().Start(async.Context().Done()).Return(nil).AnyTimes()
			cache.EXPECT().WaitForCacheSync(gomock.Any()).Return(true)
			Eventually(func() error { return async.Error() }, time.Second*1).ShouldNot(HaveOccurred())
			err := async.Start(fakeOptions)
			Expect(err).To(HaveOccurred())
			Expect(err).To(HaveInErrorChain(ManagerStartOptionsFuncError(eris.New("hello"))))
		})
	})
})
