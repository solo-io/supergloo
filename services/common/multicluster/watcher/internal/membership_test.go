package internal_watcher_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	. "github.com/solo-io/go-utils/testutils"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common/aws_creds"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common/kube"
	mock_kube "github.com/solo-io/service-mesh-hub/cli/pkg/common/kube/mocks"
	mock_mc_manager "github.com/solo-io/service-mesh-hub/services/common/multicluster/manager/mocks"
	. "github.com/solo-io/service-mesh-hub/services/common/multicluster/watcher/internal"
	rest2 "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/rest"
	mock_rest_api "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/rest/mocks"
	k8s_core_types "k8s.io/api/core/v1"
	k8s_meta_types "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd/api"
)

var _ = Describe("multicluster-watcher", func() {

	var (
		ctrl       *gomock.Controller
		ctx        context.Context
		restConfig = &rest.Config{}

		byteConfig = []byte(`
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
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Context("cluster membership", func() {

		var (
			receiver           *mock_mc_manager.MockKubeConfigHandler
			kubeConverter      *mock_kube.MockConverter
			cmh                *MeshPlatformMembershipHandler
			awsAPICredsHandler *mock_rest_api.MockRestAPICredsHandler
			clusterName        = "cluster-name"
		)

		BeforeEach(func() {
			receiver = mock_mc_manager.NewMockKubeConfigHandler(ctrl)
			kubeConverter = mock_kube.NewMockConverter(ctrl)
			awsAPICredsHandler = mock_rest_api.NewMockRestAPICredsHandler(ctrl)
			cmh = NewMeshPlatformMembershipHandler(
				receiver,
				[]rest2.RestAPICredsHandler{awsAPICredsHandler},
				kubeConverter,
			)
		})

		Context("add cluster", func() {
			It("returns an error if the secret is malformed", func() {
				secret := &k8s_core_types.Secret{
					Type: k8s_core_types.SecretTypeOpaque,
				}
				kubeConverter.EXPECT().
					SecretToConfig(secret).
					Return("", nil, kube.NoDataInKubeConfigSecret(&k8s_core_types.Secret{}))

				resync, err := cmh.AddMemberMeshPlatform(ctx, secret)
				Expect(resync).To(BeFalse(), "resync should be false")
				Expect(err).To(HaveInErrorChain(kube.NoDataInKubeConfigSecret(&k8s_core_types.Secret{})))
			})

			It("returns an error if there is an invalid kube config string", func() {
				secret := &k8s_core_types.Secret{
					Type: k8s_core_types.SecretTypeOpaque,
					ObjectMeta: k8s_meta_types.ObjectMeta{
						Name: clusterName,
					},
					Data: map[string][]byte{
						clusterName: []byte("failing config"),
					},
				}

				kubeConverter.EXPECT().
					SecretToConfig(secret).
					Return("", nil, KubeConfigInvalidFormatError(eris.New("hello"), clusterName, clusterName, ""))

				resync, err := cmh.AddMemberMeshPlatform(ctx, secret)
				Expect(resync).To(BeFalse(), "resync should be false")
				Expect(err).To(HaveInErrorChain(KubeConfigInvalidFormatError(eris.New("hello"),
					clusterName, clusterName, "")))
			})

			It("returns an error if the receiver returns an error", func() {
				secret := &k8s_core_types.Secret{
					Type: k8s_core_types.SecretTypeOpaque,
					ObjectMeta: k8s_meta_types.ObjectMeta{
						Name: clusterName,
					},
					Data: map[string][]byte{
						clusterName: byteConfig,
					},
				}

				kubeConverter.EXPECT().
					SecretToConfig(secret).
					Return(clusterName, &kube.ConvertedConfigs{
						RestConfig: restConfig,
					}, nil)

				receiverError := eris.New("this is an error")
				receiver.EXPECT().ClusterAdded(gomock.Any(), clusterName).Return(receiverError)
				resync, err := cmh.AddMemberMeshPlatform(ctx, secret)
				Expect(resync).To(BeFalse())
				Expect(err).To(HaveInErrorChain(PlatformAddError(receiverError, clusterName)))
			})

			It("can successfully add a cluster", func() {
				secret := &k8s_core_types.Secret{
					Type: k8s_core_types.SecretTypeOpaque,
					ObjectMeta: k8s_meta_types.ObjectMeta{
						Name: clusterName,
					},
					Data: map[string][]byte{
						clusterName: byteConfig,
					},
				}

				kubeConverter.EXPECT().
					SecretToConfig(secret).
					Return(clusterName, &kube.ConvertedConfigs{
						RestConfig: restConfig,
						ApiConfig: &api.Config{
							CurrentContext: "current-context",
						},
					}, nil)

				receiver.EXPECT().ClusterAdded(gomock.Any(), clusterName).Return(nil)
				resync, err := cmh.AddMemberMeshPlatform(ctx, secret)
				Expect(resync).To(BeFalse(), "resync should be false")
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("delete cluster", func() {
			It("will return an error if the receiver is called and errors", func() {
				secret := &k8s_core_types.Secret{
					Type: k8s_core_types.SecretTypeOpaque,
					ObjectMeta: k8s_meta_types.ObjectMeta{
						Name: clusterName,
					},
					Data: map[string][]byte{
						clusterName: byteConfig,
					},
				}
				kubeConverter.EXPECT().
					SecretToConfig(secret).
					Return(clusterName, &kube.ConvertedConfigs{
						RestConfig: restConfig,
						ApiConfig: &api.Config{
							CurrentContext: "current-context",
						},
					}, nil)

				receiver.EXPECT().ClusterAdded(gomock.Any(), clusterName).Return(nil)
				resync, err := cmh.AddMemberMeshPlatform(ctx, secret)
				Expect(resync).To(BeFalse(), "resync should be false")
				Expect(err).NotTo(HaveOccurred())

				receiver.EXPECT().ClusterRemoved(clusterName).Return(eris.New("this is an error"))
				resync, err = cmh.DeleteMemberMeshPlatform(ctx, &k8s_core_types.Secret{
					Type: k8s_core_types.SecretTypeOpaque,
					ObjectMeta: k8s_meta_types.ObjectMeta{
						Name: clusterName,
					},
				})
				Expect(resync).To(BeFalse())
				Expect(err).To(HaveInErrorChain(PlatformDeletionError(eris.New("hello"), clusterName)))
			})

			It("will return nil and delete cluster if return is nil", func() {
				secret := &k8s_core_types.Secret{
					Type: k8s_core_types.SecretTypeOpaque,
					ObjectMeta: k8s_meta_types.ObjectMeta{
						Name: clusterName,
					},
					Data: map[string][]byte{
						clusterName: byteConfig,
					},
				}
				kubeConverter.EXPECT().
					SecretToConfig(secret).
					Return(clusterName, &kube.ConvertedConfigs{
						RestConfig: restConfig,
						ApiConfig: &api.Config{
							CurrentContext: "current-context",
						},
					}, nil)

				receiver.EXPECT().ClusterAdded(gomock.Any(), clusterName).Return(nil)
				resync, err := cmh.AddMemberMeshPlatform(ctx, secret)
				Expect(resync).To(BeFalse(), "resync should be false")
				Expect(err).NotTo(HaveOccurred())

				receiver.EXPECT().ClusterRemoved(clusterName).Return(nil)
				resync, err = cmh.DeleteMemberMeshPlatform(ctx, &k8s_core_types.Secret{
					Type: k8s_core_types.SecretTypeOpaque,
					ObjectMeta: k8s_meta_types.ObjectMeta{
						Name: clusterName,
					},
				})
				Expect(resync).To(BeFalse(), "resync should be false")
				Expect(err).NotTo(HaveOccurred())
			})

			It("will handle adding new REST API mesh platform", func() {
				secret := &k8s_core_types.Secret{
					Type: aws_creds.AWSSecretType,
					ObjectMeta: k8s_meta_types.ObjectMeta{
						Name: clusterName,
					},
					Data: map[string][]byte{
						clusterName: byteConfig,
					},
				}
				awsAPICredsHandler.EXPECT().RestAPIAdded(ctx, secret).Return(nil)
				resync, err := cmh.AddMemberMeshPlatform(ctx, secret)
				Expect(resync).To(BeFalse())
				Expect(err).NotTo(HaveOccurred())
			})

			It("will handle adding new REST API mesh platform and resync if error occurred", func() {
				secret := &k8s_core_types.Secret{
					Type: aws_creds.AWSSecretType,
					ObjectMeta: k8s_meta_types.ObjectMeta{
						Name: clusterName,
					},
					Data: map[string][]byte{
						clusterName: byteConfig,
					},
				}
				err := eris.New("some error")
				awsAPICredsHandler.EXPECT().RestAPIAdded(ctx, secret).Return(err)
				resync, err := cmh.AddMemberMeshPlatform(ctx, secret)
				Expect(resync).To(BeFalse())
				Expect(err).To(HaveInErrorChain(PlatformAddError(err, secret.GetName())))
			})

			It("will handle removing REST API mesh platform", func() {
				secret := &k8s_core_types.Secret{
					Type: aws_creds.AWSSecretType,
					ObjectMeta: k8s_meta_types.ObjectMeta{
						Name: clusterName,
					},
					Data: map[string][]byte{
						clusterName: byteConfig,
					},
				}
				awsAPICredsHandler.EXPECT().RestAPIRemoved(ctx, secret).Return(nil)
				resync, err := cmh.DeleteMemberMeshPlatform(ctx, secret)
				Expect(resync).To(BeFalse())
				Expect(err).NotTo(HaveOccurred())
			})

			It("will handle removing REST API mesh platform and resync if error occurred", func() {
				secret := &k8s_core_types.Secret{
					Type: aws_creds.AWSSecretType,
					ObjectMeta: k8s_meta_types.ObjectMeta{
						Name: clusterName,
					},
					Data: map[string][]byte{
						clusterName: byteConfig,
					},
				}
				err := eris.New("some error")
				awsAPICredsHandler.EXPECT().RestAPIRemoved(ctx, secret).Return(err)
				resync, err := cmh.DeleteMemberMeshPlatform(ctx, secret)
				Expect(resync).To(BeFalse())
				Expect(err).To(HaveInErrorChain(PlatformDeletionError(err, secret.GetName())))
			})
		})
	})
})
