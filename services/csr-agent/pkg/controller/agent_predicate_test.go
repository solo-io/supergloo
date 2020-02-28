package csr_agent_controller_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/contextutils"
	core_types "github.com/solo-io/mesh-projects/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/mesh-projects/pkg/api/security.zephyr.solo.io/v1alpha1"
	security_types "github.com/solo-io/mesh-projects/pkg/api/security.zephyr.solo.io/v1alpha1/types"
	csr_agent_controller "github.com/solo-io/mesh-projects/services/csr-agent/pkg/controller"
	test_logging "github.com/solo-io/mesh-projects/test/logging"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

var _ = Describe("predicate", func() {
	var (
		ctx          context.Context
		csrPredicate predicate.Predicate
		testLogger   *test_logging.TestLogger
	)

	BeforeEach(func() {
		testLogger = test_logging.NewTestLogger()
		ctx = contextutils.WithExistingLogger(context.TODO(), testLogger.Logger())
		csrPredicate = csr_agent_controller.CsrAgentPredicateProvider(ctx)
	})

	It("will return false to update if type is not a MeshGroupCertificateSigningRequest", func() {
		updateEvent := event.UpdateEvent{
			MetaNew:   &metav1.ObjectMeta{},
			ObjectNew: &corev1.Pod{},
		}
		Expect(csrPredicate.Update(updateEvent)).To(BeFalse())
	})

	It("will return false if statuses are equal", func() {
		updateEvent := event.UpdateEvent{
			ObjectNew: &v1alpha1.MeshGroupCertificateSigningRequest{
				Spec: security_types.MeshGroupCertificateSigningRequestSpec{},
				Status: security_types.MeshGroupCertificateSigningRequestStatus{
					ComputedStatus: &core_types.ComputedStatus{
						Message: "hello",
					},
				},
			},
			ObjectOld: &v1alpha1.MeshGroupCertificateSigningRequest{
				Spec: security_types.MeshGroupCertificateSigningRequestSpec{},
				Status: security_types.MeshGroupCertificateSigningRequestStatus{
					ComputedStatus: &core_types.ComputedStatus{
						Message: "hello",
					},
				},
			},
		}
		Expect(csrPredicate.Update(updateEvent)).To(BeFalse())
	})

	It("will return false if cert len == 0", func() {
		updateEvent := event.UpdateEvent{
			ObjectNew: &v1alpha1.MeshGroupCertificateSigningRequest{
				Spec: security_types.MeshGroupCertificateSigningRequestSpec{},
				Status: security_types.MeshGroupCertificateSigningRequestStatus{
					Response: &security_types.MeshGroupCertificateSigningResponse{},
				},
			},
			ObjectOld: &v1alpha1.MeshGroupCertificateSigningRequest{},
		}
		Expect(csrPredicate.Update(updateEvent)).To(BeFalse())
	})

	It("will return true if cert len > 0 && statuses are not equal", func() {
		updateEvent := event.UpdateEvent{
			ObjectNew: &v1alpha1.MeshGroupCertificateSigningRequest{
				Spec: security_types.MeshGroupCertificateSigningRequestSpec{},
				Status: security_types.MeshGroupCertificateSigningRequestStatus{
					Response: &security_types.MeshGroupCertificateSigningResponse{
						CaCertificate:   []byte("ca-certs"),
						RootCertificate: []byte("root-certs"),
					},
				},
			},
			ObjectOld: &v1alpha1.MeshGroupCertificateSigningRequest{},
		}
		Expect(csrPredicate.Update(updateEvent)).To(BeTrue())
	})
})
