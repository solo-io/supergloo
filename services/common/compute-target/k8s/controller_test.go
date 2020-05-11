package mc_manager_test

import (
	"context"
	"strconv"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	. "github.com/solo-io/go-utils/testutils"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common/aws_creds"
	"github.com/solo-io/service-mesh-hub/pkg/kubeconfig"
	mock_kubeconfig "github.com/solo-io/service-mesh-hub/pkg/kubeconfig/mocks"
	compute_target "github.com/solo-io/service-mesh-hub/services/common/compute-target"
	. "github.com/solo-io/service-mesh-hub/services/common/compute-target/k8s"
	. "github.com/solo-io/service-mesh-hub/services/common/compute-target/k8s/mocks"
	k8s_core_types "k8s.io/api/core/v1"
	k8s_meta_types "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd/api"
)

var _ = Describe("mc_manager", func() {
	var (
		ctrl              *gomock.Controller
		asyncMgr          *MockAsyncManager
		asyncMgrFactory   *MockAsyncManagerFactory
		ctx               context.Context
		informer          AsyncManagerInformer
		configHandler     compute_target.ComputeTargetCredentialsHandler
		managerHandler    *MockAsyncManagerHandler
		cfg               *rest.Config
		mockKubeConverter *mock_kubeconfig.MockConverter
		constErr          = eris.New("hello")
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
		managerHandler = NewMockAsyncManagerHandler(ctrl)
		asyncMgr = NewMockAsyncManager(ctrl)
		asyncMgrFactory = NewMockAsyncManagerFactory(ctrl)
		mockKubeConverter = mock_kubeconfig.NewMockConverter(ctrl)
		managerController := NewAsyncManagerController(asyncMgrFactory, mockKubeConverter)
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
			Expect(err).To(Equal(InformerNotRegisteredError))
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
			clusterName            = "cluster-name"
			expectSecretConversion = func() *k8s_core_types.Secret {
				byteConfig := []byte(`
apiVersion: v1
clusters:
- cluster:
    server: https://localhost:9090
  name: k3s-default
contexts:
- context:
    cluster: k3s-default
    user: k3s-default
  name: k3s-default
current-context: k3s-default
kind: Config
preferences: {}
users:
- name: k3s-default
  user:
    password: admin
    username: admin

`)
				secret := &k8s_core_types.Secret{
					Type: k8s_core_types.SecretTypeOpaque,
					ObjectMeta: k8s_meta_types.ObjectMeta{
						Name: clusterName,
					},
					Data: map[string][]byte{
						clusterName: byteConfig,
					},
				}
				mockKubeConverter.EXPECT().
					SecretToConfig(secret).
					Return(clusterName, &kubeconfig.ConvertedConfigs{
						RestConfig: &rest.Config{},
						ApiConfig: &api.Config{
							CurrentContext: "current-context",
						},
					}, nil)
				return secret
			}
		)

		Context("no handlers", func() {
			Context("add", func() {
				It("will ignore non k8s cluster secret", func() {
					secret := &k8s_core_types.Secret{
						Type: aws_creds.AWSSecretType,
					}
					err := configHandler.ComputeTargetAdded(ctx, secret)
					Expect(err).To(BeNil())
				})

				It("will return an error if fail to convert secret to kubeconfig", func() {
					secret := &k8s_core_types.Secret{
						ObjectMeta: k8s_meta_types.ObjectMeta{
							Name:      "name",
							Namespace: "namespace",
						},
						Type: k8s_core_types.SecretTypeOpaque,
					}
					mockKubeConverter.EXPECT().
						SecretToConfig(secret).
						Return("", nil, eris.New(""))
					err := configHandler.ComputeTargetAdded(ctx, secret)
					Expect(err).To(HaveInErrorChain(KubeConfigInvalidFormatError(err, secret.GetName(), secret.GetNamespace())))
				})

				It("will return an error if factory fails", func() {
					asyncMgrFactory.EXPECT().New(ctx, cfg, gomock.Any()).
						Return(nil, AsyncManagerFactoryError(constErr, clusterName))
					secret := expectSecretConversion()
					err := configHandler.ComputeTargetAdded(ctx, secret)
					Expect(err).To(HaveOccurred())
					Expect(err).To(HaveInErrorChain(AsyncManagerFactoryError(constErr, clusterName)))
				})

				It("will return an error if manager fails to start", func() {
					asyncMgrFactory.EXPECT().New(ctx, cfg, gomock.Any()).
						Return(asyncMgr, nil)
					asyncMgr.EXPECT().Start().Return(eris.New("hello"))
					secret := expectSecretConversion()
					err := configHandler.ComputeTargetAdded(ctx, secret)
					Expect(err).To(HaveOccurred())
					Expect(err).To(HaveInErrorChain(AsyncManagerStartError(constErr, clusterName)))
				})
			})

			Context("remove", func() {
				It("will ignore non k8s secrets", func() {
					secret := &k8s_core_types.Secret{
						Type: aws_creds.AWSSecretType,
					}
					err := configHandler.ComputeTargetRemoved(ctx, secret)
					Expect(err).To(BeNil())
				})

				It("will return an error if get manager fails", func() {
					secret := &k8s_core_types.Secret{
						ObjectMeta: k8s_meta_types.ObjectMeta{Name: "cluster-name"},
						Type:       k8s_core_types.SecretTypeOpaque,
					}
					err := configHandler.ComputeTargetRemoved(ctx, secret)
					Expect(err).To(HaveOccurred())
					Expect(err).To(HaveInErrorChain(NoManagerForClusterError(secret.GetName())))
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
					secret := expectSecretConversion()
					err := configHandler.ComputeTargetAdded(ctx, secret)
					Expect(err).To(HaveOccurred())
					Expect(err).To(HaveInErrorChain(InformerAddFailedError(constErr, handlerName, clusterName)))
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
					secret := expectSecretConversion()
					err = configHandler.ComputeTargetAdded(ctx, secret)
					Expect(err).NotTo(HaveOccurred())
				})
			})

			Context("remove", func() {
				var (
					handlerMap *AsyncManagerHandlerMap
					managerMap *AsyncManagerMap
					secret     = &k8s_core_types.Secret{
						ObjectMeta: k8s_meta_types.ObjectMeta{Name: "cluster-name"},
						Type:       k8s_core_types.SecretTypeOpaque,
					}
				)

				BeforeEach(func() {
					handlerMap = NewAsyncManagerHandler()
					err := handlerMap.SetHandler(handlerName, managerHandler)
					managerMap = NewAsyncManagerMap()
					Expect(err).NotTo(HaveOccurred())
					managerController := NewAsyncManagerControllerWithHandlers(handlerMap, managerMap, asyncMgrFactory, mockKubeConverter)
					informer, configHandler = managerController, managerController
				})

				It("will fail if manager doesn't exist", func() {
					err := configHandler.ComputeTargetRemoved(ctx, secret)
					Expect(err).To(HaveOccurred())
					Expect(err).To(HaveInErrorChain(NoManagerForClusterError(secret.GetName())))
				})

				It("will fail if any handler fails", func() {
					Expect(managerMap.SetManager(clusterName, asyncMgr)).NotTo(HaveOccurred())
					asyncMgr.EXPECT().Stop()
					managerHandler.EXPECT().ClusterRemoved(clusterName).Return(constErr)
					err := configHandler.ComputeTargetRemoved(ctx, secret)
					Expect(err).To(HaveOccurred())
					Expect(err).To(HaveInErrorChain(InformerDeleteFailedError(constErr,
						handlerName, clusterName)))
				})

				It("will succeed, call all handlers, and remove manager", func() {
					handler2 := "handler-2"
					Expect(managerMap.SetManager(clusterName, asyncMgr)).NotTo(HaveOccurred())
					Expect(handlerMap.SetHandler(handler2, managerHandler)).NotTo(HaveOccurred())
					asyncMgr.EXPECT().Stop()
					managerHandler.EXPECT().ClusterRemoved(clusterName).Return(nil).Times(2)
					err := configHandler.ComputeTargetRemoved(ctx, secret)
					Expect(err).NotTo(HaveOccurred())
					_, ok := managerMap.GetManager(handlerName)
					Expect(ok).To(BeFalse())
				})
			})
		})
	})
})
