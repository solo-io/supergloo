package k8s_test

import (
	"context"
	"errors"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/testutils"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/discovery/mesh-workload/k8s"
	mock_apps "github.com/solo-io/service-mesh-hub/test/mocks/clients/kubernetes/apps/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("OwnerFetcher", func() {
	var (
		ctrl                 *gomock.Controller
		ctx                  context.Context
		ownerFetcher         k8s.OwnerFetcher
		mockDeploymentClient *mock_apps.MockDeploymentClient
		mockReplicaSetClient *mock_apps.MockReplicaSetClient
		t                    = true
		replicaSetName       = "replicaSetName"
		deploymentName       = "deploymentName"
		namespace            = "namespace"
		replicaSetObjKey     = client.ObjectKey{Namespace: namespace, Name: replicaSetName}
		deploymentObjKey     = client.ObjectKey{Namespace: namespace, Name: deploymentName}
		pod                  = &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: namespace,
				Name:      "podname",
				OwnerReferences: []metav1.OwnerReference{
					{
						Controller: &t,
						Name:       replicaSetName,
					},
				},
			},
			TypeMeta: metav1.TypeMeta{
				Kind: "Pod",
			},
		}
		replicaSet = &appsv1.ReplicaSet{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: namespace,
				Name:      replicaSetName,
				OwnerReferences: []metav1.OwnerReference{
					{
						Controller: &t,
						Name:       deploymentName,
					},
				},
			},
			TypeMeta: metav1.TypeMeta{
				Kind: "ReplicaSet",
			},
		}
	)
	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
		mockDeploymentClient = mock_apps.NewMockDeploymentClient(ctrl)
		mockReplicaSetClient = mock_apps.NewMockReplicaSetClient(ctrl)
		ownerFetcher = k8s.NewOwnerFetcher(mockDeploymentClient, mockReplicaSetClient)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("should get deployment", func() {
		expectedDeployment := &appsv1.Deployment{}
		mockReplicaSetClient.EXPECT().GetReplicaSet(ctx, replicaSetObjKey).Return(replicaSet, nil)
		mockDeploymentClient.EXPECT().GetDeployment(ctx, deploymentObjKey).Return(expectedDeployment, nil)
		deployment, err := ownerFetcher.GetDeployment(ctx, pod)
		Expect(err).ToNot(HaveOccurred())
		Expect(expectedDeployment).To(Equal(deployment))
	})

	It("should return error if owner ref for Pod not found", func() {
		f := false
		podWithoutOwner := *pod
		podWithoutOwner.OwnerReferences = []metav1.OwnerReference{
			{
				Controller: &f,
				Name:       replicaSetName,
			},
		}
		_, err := ownerFetcher.GetDeployment(ctx, &podWithoutOwner)
		Expect(err).To(testutils.HaveInErrorChain(
			k8s.ControllerOwnerNotFound(namespace, podWithoutOwner.Name, podWithoutOwner.TypeMeta.Kind)))
	})

	It("should return error if can't fetch ReplicaSet for Pod", func() {
		expectedErr := errors.New("can't fetch ReplicaSet")
		mockReplicaSetClient.EXPECT().GetReplicaSet(ctx, replicaSetObjKey).Return(nil, expectedErr)
		_, err := ownerFetcher.GetDeployment(ctx, pod)
		Expect(err).To(testutils.HaveInErrorChain(expectedErr))
	})

	It("should return error if owner ref for ReplicaSet not found", func() {
		f := false
		replicaSetWithoutOwner := *replicaSet
		replicaSetWithoutOwner.OwnerReferences = []metav1.OwnerReference{
			{
				Controller: &f,
				Name:       replicaSetName,
			},
		}
		mockReplicaSetClient.EXPECT().GetReplicaSet(ctx, replicaSetObjKey).Return(&replicaSetWithoutOwner, nil)
		_, err := ownerFetcher.GetDeployment(ctx, pod)
		Expect(err).To(testutils.HaveInErrorChain(
			k8s.ControllerOwnerNotFound(namespace, replicaSetWithoutOwner.Name, replicaSetWithoutOwner.TypeMeta.Kind)))
	})

	It("should return error if can't fetch Deployment for ReplicaSet", func() {
		expectedErr := errors.New("can't fetch Deployment")
		mockReplicaSetClient.EXPECT().GetReplicaSet(ctx, replicaSetObjKey).Return(replicaSet, nil)
		mockDeploymentClient.EXPECT().GetDeployment(ctx, deploymentObjKey).Return(nil, expectedErr)
		_, err := ownerFetcher.GetDeployment(ctx, pod)
		Expect(err).To(testutils.HaveInErrorChain(expectedErr))
	})
})
