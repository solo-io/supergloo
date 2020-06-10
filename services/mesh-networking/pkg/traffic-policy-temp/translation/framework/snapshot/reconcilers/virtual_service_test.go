package reconcilers_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/service-mesh-hub/pkg/common/kube/selection"
	. "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/traffic-policy-temp/translation/framework/snapshot/reconcilers"
	mock_istio_networking_clients "github.com/solo-io/service-mesh-hub/test/mocks/clients/istio/networking/v1beta1"
	"istio.io/api/networking/v1alpha3"
	istio_networking "istio.io/client-go/pkg/apis/networking/v1alpha3"
	k8s_meta_types "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("VirtualService Reconciler", func() {
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
		_, err := NewVirtualServiceReconcilerBuilder().
			ScopedToNamespace("test-ns").
			ScopedToWholeCluster().
			Build()
		Expect(err).To(HaveOccurred())
	})

	It("can write new VirtualService to the proper namespace", func() {
		vsClient := mock_istio_networking_clients.NewMockVirtualServiceClient(ctrl)
		reconciler := NewVirtualServiceReconciler("test-ns", nil, vsClient)
		toReconcile := []*istio_networking.VirtualService{
			{ObjectMeta: k8s_meta_types.ObjectMeta{Name: "vs-1"}},
			{ObjectMeta: k8s_meta_types.ObjectMeta{Name: "vs-2"}},
		}

		vsClient.EXPECT().
			ListVirtualService(ctx, client.InNamespace("test-ns"), client.MatchingLabels(nil)).
			Return(&istio_networking.VirtualServiceList{}, nil)
		vsClient.EXPECT().
			CreateVirtualService(ctx, toReconcile[0]).
			Return(nil)
		vsClient.EXPECT().
			CreateVirtualService(ctx, toReconcile[1]).
			Return(nil)

		err := reconciler.Reconcile(ctx, toReconcile)
		Expect(err).NotTo(HaveOccurred())
	})

	It("can delete VirtualService that should not exist", func() {
		vsClient := mock_istio_networking_clients.NewMockVirtualServiceClient(ctrl)
		reconciler := NewVirtualServiceReconciler("test-ns", nil, vsClient)
		toReconcile := []*istio_networking.VirtualService{
			{ObjectMeta: k8s_meta_types.ObjectMeta{Name: "vs-1"}},
			{ObjectMeta: k8s_meta_types.ObjectMeta{Name: "vs-2"}},
		}

		vsClient.EXPECT().
			ListVirtualService(ctx, client.InNamespace("test-ns"), client.MatchingLabels(nil)).
			Return(&istio_networking.VirtualServiceList{
				Items: []istio_networking.VirtualService{
					*toReconcile[0],
					*toReconcile[1],
				},
			}, nil)
		vsClient.EXPECT().
			DeleteVirtualService(ctx, selection.ObjectMetaToObjectKey(toReconcile[0].ObjectMeta)).
			Return(nil)
		vsClient.EXPECT().
			DeleteVirtualService(ctx, selection.ObjectMetaToObjectKey(toReconcile[1].ObjectMeta)).
			Return(nil)

		err := reconciler.Reconcile(ctx, nil)
		Expect(err).NotTo(HaveOccurred())
	})

	It("can issue update actions for updated VirtualServices", func() {
		vsClient := mock_istio_networking_clients.NewMockVirtualServiceClient(ctrl)
		reconciler := NewVirtualServiceReconciler("test-ns", nil, vsClient)
		toReconcile := []*istio_networking.VirtualService{
			{ObjectMeta: k8s_meta_types.ObjectMeta{Name: "vs-1"}},
			{ObjectMeta: k8s_meta_types.ObjectMeta{Name: "vs-2"}},
		}
		copy1 := *toReconcile[0]
		copy2 := *toReconcile[1]

		copy1.Spec = v1alpha3.VirtualService{
			Hosts: []string{"updated-virtual-service"},
		}
		copy2.Spec = v1alpha3.VirtualService{
			Hosts: []string{"updated-virtual-service"},
		}

		vsClient.EXPECT().
			ListVirtualService(ctx, client.InNamespace("test-ns"), client.MatchingLabels(nil)).
			Return(&istio_networking.VirtualServiceList{
				Items: []istio_networking.VirtualService{
					copy1,
					copy2,
				},
			}, nil)
		vsClient.EXPECT().
			UpdateVirtualService(ctx, toReconcile[0]).
			Return(nil)
		vsClient.EXPECT().
			UpdateVirtualService(ctx, toReconcile[1]).
			Return(nil)

		err := reconciler.Reconcile(ctx, toReconcile)
		Expect(err).NotTo(HaveOccurred())
	})
})
