package internal_test

import (
	"context"
	"fmt"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	"github.com/solo-io/service-mesh-hub/cli/pkg/tree/check/healthcheck/internal"
	healthcheck_types "github.com/solo-io/service-mesh-hub/cli/pkg/tree/check/healthcheck/types"
	"github.com/solo-io/service-mesh-hub/pkg/env"
	mock_kubernetes_core "github.com/solo-io/service-mesh-hub/test/mocks/clients/kubernetes/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("SMH components check", func() {
	var (
		ctrl *gomock.Controller
		ctx  context.Context
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("reports a generic error if client fails", func() {
		testErr := eris.New("test-err")
		podClient := mock_kubernetes_core.NewMockPodClient(ctrl)
		podClient.EXPECT().
			ListPod(ctx, client.InNamespace(env.GetWriteNamespace())).
			Return(nil, testErr)

		check := internal.NewSmhComponentsHealthCheck()
		clients := healthcheck_types.Clients{
			PodClient: podClient,
		}
		runFailure, checkApplies := check.Run(ctx, env.GetWriteNamespace(), clients)

		Expect(checkApplies).To(BeTrue())
		Expect(runFailure).NotTo(BeNil())
		Expect(runFailure.ErrorMessage).To(Equal(internal.GenericCheckFailed(testErr).Error()))
	})

	It("reports an error if SMH is not installed", func() {
		podClient := mock_kubernetes_core.NewMockPodClient(ctrl)
		podClient.EXPECT().
			ListPod(ctx, client.InNamespace(env.GetWriteNamespace())).
			Return(&v1.PodList{}, nil)

		check := internal.NewSmhComponentsHealthCheck()
		clients := healthcheck_types.Clients{
			PodClient: podClient,
		}
		runFailure, checkApplies := check.Run(ctx, env.GetWriteNamespace(), clients)

		Expect(checkApplies).To(BeTrue())
		Expect(runFailure).NotTo(BeNil())
		Expect(runFailure.ErrorMessage).To(Equal(internal.NoServiceMeshHubComponentsExist.Error()))
		Expect(runFailure.Hint).To(Equal("you can install Service Mesh Hub with `meshctl install`"))
	})

	It("reports an error if a pod is failing, and provides a helpful hint", func() {
		podClient := mock_kubernetes_core.NewMockPodClient(ctrl)
		podClient.EXPECT().
			ListPod(ctx, client.InNamespace(env.GetWriteNamespace())).
			Return(&v1.PodList{
				Items: []v1.Pod{{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "mesh-discovery-random-uuid",
						Namespace: env.GetWriteNamespace(),
					},
					Status: v1.PodStatus{
						ContainerStatuses: []v1.ContainerStatus{{
							Name: "mesh-discovery",
							State: v1.ContainerState{
								Waiting: &v1.ContainerStateWaiting{
									// taken from a real pod failure
									Reason:  "CrashLoopBackoff",
									Message: "back-off 1m20s",
								},
							},
						}},
					},
				}},
			}, nil)

		check := internal.NewSmhComponentsHealthCheck()
		clients := healthcheck_types.Clients{
			PodClient: podClient,
		}
		runFailure, checkApplies := check.Run(ctx, env.GetWriteNamespace(), clients)

		Expect(checkApplies).To(BeTrue())
		Expect(runFailure).NotTo(BeNil())
		Expect(runFailure.ErrorMessage).To(Equal(eris.New(fmt.Sprintf("Container %s in pod %s is waiting: CrashLoopBackoff (back-off 1m20s)", "mesh-discovery", "mesh-discovery-random-uuid")).Error()))
		Expect(runFailure.Hint).To(Equal(fmt.Sprintf("try running either `kubectl -n %s describe pod %s` or `kubectl -n %s logs %s`", env.GetWriteNamespace(), "mesh-discovery-random-uuid", env.GetWriteNamespace(), "mesh-discovery-random-uuid")))
	})
})
