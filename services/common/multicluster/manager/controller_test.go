package manager_test

import (
	"context"
	"strconv"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	. "github.com/solo-io/go-utils/testutils"
	"github.com/solo-io/service-mesh-hub/services/common/multicluster/manager"
	. "github.com/solo-io/service-mesh-hub/services/common/multicluster/manager/mocks"
	mock_controller_runtime "github.com/solo-io/service-mesh-hub/test/mocks/controller-runtime"
	"k8s.io/client-go/rest"
)

var _ = Describe("mc_manager", func() {
	var (
		ctrl            *gomock.Controller
		mgr             *mock_controller_runtime.MockManager
		asyncMgr        *MockAsyncManager
		asyncMgrFactory *MockAsyncManagerFactory
		ctx             context.Context
		informer        manager.AsyncManagerInformer
		configHandler   manager.KubeConfigHandler
		managerHandler  *MockAsyncManagerHandler
		cfg             *rest.Config

		constErr = eris.New("hello")
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mgr = mock_controller_runtime.NewMockManager(ctrl)
		managerHandler = NewMockAsyncManagerHandler(ctrl)
		asyncMgr = NewMockAsyncManager(ctrl)
		asyncMgrFactory = NewMockAsyncManagerFactory(ctrl)
		ctx = context.TODO()
		managerController := manager.NewAsyncManagerControllerFromLocal(ctx, mgr, asyncMgrFactory)
		informer, configHandler = managerController, managerController
		cfg = &rest.Config{}
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Context("receiver", func() {
		It("will throw an error if configHandler exists", func() {
			err := informer.RemoveHandler("")
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(manager.InformerNotRegisteredError))
		})

		It("can properly add a unique receiver", func() {
			Expect(informer.AddHandler(managerHandler, "")).NotTo(HaveOccurred())
		})

		It("can properly add many unique handlers", func() {
			for i := 0; i < 100; i++ {
				Expect(informer.AddHandler(managerHandler, strconv.Itoa(i))).NotTo(HaveOccurred())
			}
		})
	})

	Context("async manager controller", func() {
		var (
			clusterName = "cluster-name"
		)

		Context("no handlers", func() {
			Context("add", func() {
				It("will return an error if factory fails", func() {
					asyncMgrFactory.EXPECT().New(ctx, cfg, gomock.Any()).
						Return(nil, manager.AsyncManagerFactoryError(constErr, clusterName))
					err := configHandler.ClusterAdded(cfg, clusterName)
					Expect(err).To(HaveOccurred())
					Expect(err).To(HaveInErrorChain(manager.AsyncManagerFactoryError(constErr, clusterName)))
				})
				It("will return an error if manager fails to start", func() {
					asyncMgrFactory.EXPECT().New(ctx, cfg, gomock.Any()).
						Return(asyncMgr, nil)
					asyncMgr.EXPECT().Start().Return(eris.New("hello"))
					err := configHandler.ClusterAdded(cfg, clusterName)
					Expect(err).To(HaveOccurred())
					Expect(err).To(HaveInErrorChain(manager.AsyncManagerStartError(constErr, clusterName)))
				})
			})
			Context("delete", func() {
				It("will return an error if get manager fails", func() {
					err := configHandler.ClusterRemoved(clusterName)
					Expect(err).To(HaveOccurred())
					Expect(err).To(HaveInErrorChain(manager.NoManagerForClusterError(clusterName)))
				})
			})
		})

		Context("with handlers", func() {
			var (
				handlerName = "handler"
			)
			BeforeEach(func() {
				err := informer.AddHandler(managerHandler, handlerName)
				Expect(err).NotTo(HaveOccurred())
			})
			Context("add", func() {
				It("will fail if any handler fails", func() {
					asyncMgrFactory.EXPECT().New(ctx, cfg, gomock.Any()).
						Return(asyncMgr, nil)
					asyncMgr.EXPECT().Start().Return(nil)
					asyncMgr.EXPECT().Context().Return(ctx)
					managerHandler.EXPECT().ClusterAdded(ctx, asyncMgr, clusterName).Return(constErr)
					err := configHandler.ClusterAdded(cfg, clusterName)
					Expect(err).To(HaveOccurred())
					Expect(err).To(HaveInErrorChain(manager.InformerAddFailedError(constErr, handlerName, clusterName)))
				})
				It("will succeed when all handlers succeed", func() {
					handler2 := "handler-2"
					err := informer.AddHandler(managerHandler, handler2)
					Expect(err).NotTo(HaveOccurred())
					asyncMgrFactory.EXPECT().New(ctx, cfg, gomock.Any()).
						Return(asyncMgr, nil)
					asyncMgr.EXPECT().Start().Return(nil)
					asyncMgr.EXPECT().Context().Return(ctx).Times(2)
					managerHandler.EXPECT().ClusterAdded(ctx, asyncMgr, clusterName).Return(nil).Times(2)
					err = configHandler.ClusterAdded(cfg, clusterName)
					Expect(err).NotTo(HaveOccurred())
				})
			})
			Context("delete", func() {
				var (
					handlerMap *manager.AsyncManagerHandlerMap
					managerMap *manager.AsyncManagerMap
				)
				BeforeEach(func() {
					handlerMap = manager.NewAsyncManagerHandler()
					err := handlerMap.SetHandler(handlerName, managerHandler)
					managerMap = manager.NewAsyncManagerMap()
					Expect(err).NotTo(HaveOccurred())
					managerController := manager.NewAsyncManagerController(ctx, handlerMap,
						managerMap, asyncMgrFactory)
					informer, configHandler = managerController, managerController
				})
				It("will fail if manager doesn't exist", func() {
					err := configHandler.ClusterRemoved(clusterName)
					Expect(err).To(HaveOccurred())
					Expect(err).To(HaveInErrorChain(manager.NoManagerForClusterError(clusterName)))
				})
				It("will fail if any handler fails", func() {
					Expect(managerMap.SetManager(clusterName, asyncMgr)).NotTo(HaveOccurred())
					asyncMgr.EXPECT().Stop()
					managerHandler.EXPECT().ClusterRemoved(clusterName).Return(constErr)
					err := configHandler.ClusterRemoved(clusterName)
					Expect(err).To(HaveOccurred())
					Expect(err).To(HaveInErrorChain(manager.InformerDeleteFailedError(constErr,
						handlerName, clusterName)))
				})
				It("will succeed, call all handlers, and remove manager", func() {
					handler2 := "handler-2"
					Expect(managerMap.SetManager(clusterName, asyncMgr)).NotTo(HaveOccurred())
					Expect(handlerMap.SetHandler(handler2, managerHandler)).NotTo(HaveOccurred())
					asyncMgr.EXPECT().Stop()
					managerHandler.EXPECT().ClusterRemoved(clusterName).Return(nil).Times(2)
					err := configHandler.ClusterRemoved(clusterName)
					Expect(err).NotTo(HaveOccurred())
					_, ok := managerMap.GetManager(handlerName)
					Expect(ok).To(BeFalse())
				})
			})
		})
	})
})
