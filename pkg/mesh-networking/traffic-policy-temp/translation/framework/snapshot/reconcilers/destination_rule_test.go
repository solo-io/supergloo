package reconcilers_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/service-mesh-hub/pkg/common/kube/selection"
	. "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/traffic-policy-temp/translation/framework/snapshot/reconcilers"
	mock_istio_networking_clients "github.com/solo-io/service-mesh-hub/test/mocks/clients/istio/networking/v1beta1"
	"istio.io/api/networking/v1alpha3"
	istio_networking "istio.io/client-go/pkg/apis/networking/v1alpha3"
	k8s_meta_types "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("DestinationRule Reconciler", func() {
	var (
		ctx  context.Context
		ctrl *gomock.Controller
	)

	BeforeEach(func() {
		ctx = context.TODO()
		ctrl = gomock.NewController(GinkgoT())
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("errors in the builder if conflicting values set", func() {
		_, err := NewDestinationRuleReconcilerBuilder().
			ScopedToNamespace("test-ns").
			ScopedToWholeCluster().
			Build()
		Expect(err).To(HaveOccurred())
	})

	It("can write new DestinationRules to the proper namespace", func() {
		drClient := mock_istio_networking_clients.NewMockDestinationRuleClient(ctrl)
		reconciler := NewDestinationRuleReconciler("test-ns", nil, drClient)
		rulesToReconcile := []*istio_networking.DestinationRule{
			{ObjectMeta: k8s_meta_types.ObjectMeta{Name: "dr-1"}},
			{ObjectMeta: k8s_meta_types.ObjectMeta{Name: "dr-2"}},
		}

		drClient.EXPECT().
			ListDestinationRule(ctx, client.InNamespace("test-ns"), client.MatchingLabels(nil)).
			Return(&istio_networking.DestinationRuleList{}, nil)
		drClient.EXPECT().
			CreateDestinationRule(ctx, rulesToReconcile[0]).
			Return(nil)
		drClient.EXPECT().
			CreateDestinationRule(ctx, rulesToReconcile[1]).
			Return(nil)

		err := reconciler.Reconcile(ctx, rulesToReconcile)
		Expect(err).NotTo(HaveOccurred())
	})

	It("can delete DestinationRules that should not exist", func() {
		drClient := mock_istio_networking_clients.NewMockDestinationRuleClient(ctrl)
		reconciler := NewDestinationRuleReconciler("test-ns", nil, drClient)
		rulesToReconcile := []*istio_networking.DestinationRule{
			{ObjectMeta: k8s_meta_types.ObjectMeta{Name: "dr-1"}},
			{ObjectMeta: k8s_meta_types.ObjectMeta{Name: "dr-2"}},
		}

		drClient.EXPECT().
			ListDestinationRule(ctx, client.InNamespace("test-ns"), client.MatchingLabels(nil)).
			Return(&istio_networking.DestinationRuleList{
				Items: []istio_networking.DestinationRule{
					*rulesToReconcile[0],
					*rulesToReconcile[1],
				},
			}, nil)
		drClient.EXPECT().
			DeleteDestinationRule(ctx, selection.ObjectMetaToObjectKey(rulesToReconcile[0].ObjectMeta)).
			Return(nil)
		drClient.EXPECT().
			DeleteDestinationRule(ctx, selection.ObjectMetaToObjectKey(rulesToReconcile[1].ObjectMeta)).
			Return(nil)

		err := reconciler.Reconcile(ctx, nil)
		Expect(err).NotTo(HaveOccurred())
	})

	It("can issue update actions for updated DestinationRules", func() {
		drClient := mock_istio_networking_clients.NewMockDestinationRuleClient(ctrl)
		reconciler := NewDestinationRuleReconciler("test-ns", nil, drClient)
		rulesToReconcile := []*istio_networking.DestinationRule{
			{ObjectMeta: k8s_meta_types.ObjectMeta{Name: "dr-1"}},
			{ObjectMeta: k8s_meta_types.ObjectMeta{Name: "dr-2"}},
		}
		dr1Copy := *rulesToReconcile[0]
		dr2Copy := *rulesToReconcile[1]

		dr1Copy.Spec = v1alpha3.DestinationRule{
			Host: "updated-destination-rule",
		}
		dr2Copy.Spec = v1alpha3.DestinationRule{
			Host: "updated-destination-rule",
		}

		drClient.EXPECT().
			ListDestinationRule(ctx, client.InNamespace("test-ns"), client.MatchingLabels(nil)).
			Return(&istio_networking.DestinationRuleList{
				Items: []istio_networking.DestinationRule{
					dr1Copy,
					dr2Copy,
				},
			}, nil)
		drClient.EXPECT().
			UpdateDestinationRule(ctx, rulesToReconcile[0]).
			Return(nil)
		drClient.EXPECT().
			UpdateDestinationRule(ctx, rulesToReconcile[1]).
			Return(nil)

		err := reconciler.Reconcile(ctx, rulesToReconcile)
		Expect(err).NotTo(HaveOccurred())
	})
})
