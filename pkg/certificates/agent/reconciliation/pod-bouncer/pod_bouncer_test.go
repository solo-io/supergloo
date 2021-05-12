package pod_bouncer_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1client "github.com/solo-io/external-apis/pkg/api/k8s/core/v1/mocks"
	corev1sets "github.com/solo-io/external-apis/pkg/api/k8s/core/v1/sets"
	certificatesv1 "github.com/solo-io/gloo-mesh/pkg/api/certificates.mesh.gloo.solo.io/v1"
	. "github.com/solo-io/gloo-mesh/pkg/certificates/agent/reconciliation/pod-bouncer"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("PodBouncer", func() {
	var (
		ctrl *gomock.Controller
		ctx  context.Context

		podClientMock *corev1client.MockPodClient
	)

	BeforeEach(func() {
		ctrl, ctx = gomock.WithContext(context.Background(), GinkgoT())
		podClientMock = corev1client.NewMockPodClient(ctrl)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("can signal that bounced pods are not ready", func() {
		// Don't mock this dependency as we want to test together
		podBouncer := NewPodBouncer(podClientMock, NewSecretRootCertMatcher())

		pbd := &certificatesv1.PodBounceDirective{
			Spec: certificatesv1.PodBounceDirectiveSpec{
				PodsToBounce: []*certificatesv1.PodBounceDirectiveSpec_PodSelector{
					{
						WaitForReplicas: 1,
						Labels:          map[string]string{"app": "gloo"},
					},
				},
			},
			Status: certificatesv1.PodBounceDirectiveStatus{
				PodsBounced: []*certificatesv1.PodBounceDirectiveStatus_BouncedPodSet{
					{
						BouncedPods: []string{
							"pod1-nejrthoiw-shfdsa",
						},
					},
				},
			},
		}

		pods := corev1sets.NewPodSet(
			&corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name: "pod1-nejrthoiw-shfdsa",
				},
			},
			&corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "pod2-ahfidsha-asdasd",
					Labels: map[string]string{"app": "hello"},
				},
			},
		)

		wait, err := podBouncer.BouncePods(ctx, pbd, pods, nil, nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(wait).To(BeTrue())
	})
})
