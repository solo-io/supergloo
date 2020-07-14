package mc_manager_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/solo-io/go-utils/testutils"
	. "github.com/solo-io/service-mesh-hub/pkg/common/compute-target/k8s"
	. "github.com/solo-io/service-mesh-hub/pkg/common/compute-target/k8s/mocks"
	mock_controller_runtime "github.com/solo-io/service-mesh-hub/test/mocks/controller-runtime"
)

var _ = Describe("sync maps", func() {
	var (
		ctrl *gomock.Controller
	)
	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
	})

	AfterEach(func() {
		ctrl.Finish()
	})
	Context("AsyncManagerMap", func() {
		var (
			asyncManagerMap *AsyncManagerMap
			asyncManager    AsyncManager
			mockManager     *mock_controller_runtime.MockManager
		)

		BeforeEach(func() {
			asyncManagerMap = NewAsyncManagerMap()
			mockManager = mock_controller_runtime.NewMockManager(ctrl)
			asyncManager = NewAsyncManager(context.TODO(), mockManager)
		})

		It("errors on double add", func() {
			err := asyncManagerMap.SetManager("test", asyncManager)
			Expect(err).NotTo(HaveOccurred())
			err = asyncManagerMap.SetManager("test", asyncManager)
			Expect(err).To(HaveOccurred())
			Expect(err).To(HaveInErrorChain(AsyncManagerExistsError("test")))

		})

		It("get/set", func() {
			err := asyncManagerMap.SetManager("test", asyncManager)
			Expect(err).NotTo(HaveOccurred())
			val, ok := asyncManagerMap.GetManager("test")
			Expect(ok).To(BeTrue())
			Expect(val).To(BeEquivalentTo(asyncManager))
		})
		It("list/remove", func() {
			err := asyncManagerMap.SetManager("test-1", asyncManager)
			Expect(err).NotTo(HaveOccurred())
			err = asyncManagerMap.SetManager("test-2", asyncManager)
			Expect(err).NotTo(HaveOccurred())
			list := asyncManagerMap.ListManagersByName()
			Expect(list).To(HaveLen(2))
			delete(list, "test-1")
			Expect(list).To(HaveLen(1))
			Expect(asyncManagerMap.ListManagersByName()).To(HaveLen(2))
		})
	})

	Context("AsyncManagerHandlerMap", func() {
		var (
			managerHandlerMap   *AsyncManagerHandlerMap
			asyncManagerHandler *MockAsyncManagerHandler
		)

		BeforeEach(func() {
			asyncManagerHandler = NewMockAsyncManagerHandler(ctrl)
			managerHandlerMap = NewAsyncManagerHandler()
		})

		It("get/set", func() {
			err := managerHandlerMap.SetHandler("test", asyncManagerHandler)
			Expect(err).NotTo(HaveOccurred())
			val, ok := managerHandlerMap.GetHandler("test")
			Expect(ok).To(BeTrue())
			Expect(val).To(BeEquivalentTo(asyncManagerHandler))
		})
		It("list/remove", func() {
			err := managerHandlerMap.SetHandler("test-1", asyncManagerHandler)
			Expect(err).NotTo(HaveOccurred())
			err = managerHandlerMap.SetHandler("test-2", asyncManagerHandler)
			Expect(err).NotTo(HaveOccurred())
			list := managerHandlerMap.ListHandlersByName()
			Expect(list).To(HaveLen(2))
			delete(list, "test-1")
			Expect(list).To(HaveLen(1))
			Expect(managerHandlerMap.ListHandlersByName()).To(HaveLen(2))
		})
	})
})
