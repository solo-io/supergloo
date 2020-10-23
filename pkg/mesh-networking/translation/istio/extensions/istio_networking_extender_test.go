package extensions_test

import (
	"context"

	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/extensions"
	mock_extensions "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/extensions/mocks"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	discoveryv1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/extensions/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/output/istio"
	mock_istio_extensions "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/extensions/mocks"
	v1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	networkingv1alpha3spec "istio.io/api/networking/v1alpha3"
	istionetworkingv1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/extensions"
)

//go:generate mockgen -destination mocks/mock_extensions_client.go -package mock_extensions github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/extensions/v1alpha1 NetworkingExtensionsClient,NetworkingExtensions_WatchPushNotificationsClient

var _ = Describe("IstioNetworkingExtender", func() {
	var (
		ctl         *gomock.Controller
		client      *mock_istio_extensions.MockNetworkingExtensionsClient
		mockClients extensions.Clients
		clientset   *mock_extensions.MockClientset
		ctx         = context.TODO()
		exts        IstioExtender
	)
	BeforeEach(func() {
		ctl = gomock.NewController(GinkgoT())
		client = mock_istio_extensions.NewMockNetworkingExtensionsClient(ctl)
		clientset = mock_extensions.NewMockClientset(ctl)
		exts = NewIstioExtensions(clientset)
		mockClients = extensions.Clients{client}
	})
	AfterEach(func() {
		ctl.Finish()
	})

	It("applies patches to traffic target outputs", func() {
		tt := &discoveryv1alpha2.TrafficTarget{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "traffictarget",
				Namespace: "namespace",
			},
			Spec: discoveryv1alpha2.TrafficTargetSpec{
				Type: &discoveryv1alpha2.TrafficTargetSpec_KubeService_{
					KubeService: &discoveryv1alpha2.TrafficTargetSpec_KubeService{
						Ref: &v1.ClusterObjectRef{
							Name:        "foo",
							Namespace:   "bar",
							ClusterName: "cluster",
						},
					},
				},
			},
		}

		outputs := istio.NewBuilder(ctx, "test")
		outputs.AddVirtualServices(&istionetworkingv1alpha3.VirtualService{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "foo",
				Namespace: "bar",
			},
		})
		expectedOutputs := outputs.Clone()
		// modify
		expectedOutputs.AddVirtualServices(&istionetworkingv1alpha3.VirtualService{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "foo",
				Namespace: "bar",
			},
			Spec: networkingv1alpha3spec.VirtualService{
				Hosts: []string{"added-a-host"},
			},
		})
		// add
		expectedOutputs.AddDestinationRules(&istionetworkingv1alpha3.DestinationRule{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "foo",
				Namespace: "bar",
			},
		})

		clientset.EXPECT().GetClients().Return(mockClients)
		client.EXPECT().GetTrafficTargetPatches(ctx, &v1alpha1.TrafficTargetPatchRequest{
			TrafficTarget: &v1alpha1.TrafficTargetResource{
				Metadata: extensions.ObjectMetaToProto(tt.ObjectMeta),
				Spec:     &tt.Spec,
				Status:   &tt.Status,
			},
			GeneratedResources: MakeGeneratedResources(outputs),
		}).Return(&v1alpha1.PatchList{
			PatchedResources: MakeGeneratedResources(expectedOutputs),
		}, nil)

		// sanity check
		Expect(outputs).NotTo(Equal(expectedOutputs))

		err := exts.PatchTrafficTargetOutputs(ctx, tt, outputs)
		Expect(err).NotTo(HaveOccurred())

		// expect patches to be applied
		Expect(outputs).To(Equal(expectedOutputs))

	})
	It("applies patches to workload outputs", func() {
		tt := &discoveryv1alpha2.Workload{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "workload",
				Namespace: "namespace",
			},
		}

		outputs := istio.NewBuilder(ctx, "test")
		outputs.AddVirtualServices(&istionetworkingv1alpha3.VirtualService{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "foo",
				Namespace: "bar",
			},
		})
		expectedOutputs := outputs.Clone()
		// modify
		expectedOutputs.AddVirtualServices(&istionetworkingv1alpha3.VirtualService{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "foo",
				Namespace: "bar",
			},
			Spec: networkingv1alpha3spec.VirtualService{
				Hosts: []string{"added-a-host"},
			},
		})
		// add
		expectedOutputs.AddDestinationRules(&istionetworkingv1alpha3.DestinationRule{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "foo",
				Namespace: "bar",
			},
		})

		clientset.EXPECT().GetClients().Return(mockClients)
		client.EXPECT().GetWorkloadPatches(ctx, &v1alpha1.WorkloadPatchRequest{
			Workload: &v1alpha1.WorkloadResource{
				Metadata: extensions.ObjectMetaToProto(tt.ObjectMeta),
				Spec:     &tt.Spec,
				Status:   &tt.Status,
			},
			GeneratedResources: MakeGeneratedResources(outputs),
		}).Return(&v1alpha1.PatchList{
			PatchedResources: MakeGeneratedResources(expectedOutputs),
		}, nil)

		// sanity check
		Expect(outputs).NotTo(Equal(expectedOutputs))

		err := exts.PatchWorkloadOutputs(ctx, tt, outputs)
		Expect(err).NotTo(HaveOccurred())

		// expect patches to be applied
		Expect(outputs).To(Equal(expectedOutputs))

	})
	It("applies patches to mesh outputs", func() {
		tt := &discoveryv1alpha2.Mesh{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "mesh",
				Namespace: "namespace",
			},
		}

		outputs := istio.NewBuilder(ctx, "test")
		outputs.AddVirtualServices(&istionetworkingv1alpha3.VirtualService{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "foo",
				Namespace: "bar",
			},
		})
		expectedOutputs := outputs.Clone()
		// modify
		expectedOutputs.AddVirtualServices(&istionetworkingv1alpha3.VirtualService{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "foo",
				Namespace: "bar",
			},
			Spec: networkingv1alpha3spec.VirtualService{
				Hosts: []string{"added-a-host"},
			},
		})
		// add
		expectedOutputs.AddDestinationRules(&istionetworkingv1alpha3.DestinationRule{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "foo",
				Namespace: "bar",
			},
		})

		clientset.EXPECT().GetClients().Return(mockClients)
		client.EXPECT().GetMeshPatches(ctx, &v1alpha1.MeshPatchRequest{
			Mesh: &v1alpha1.MeshResource{
				Metadata: extensions.ObjectMetaToProto(tt.ObjectMeta),
				Spec:     &tt.Spec,
				Status:   &tt.Status,
			},
			GeneratedResources: MakeGeneratedResources(outputs),
		}).Return(&v1alpha1.PatchList{
			PatchedResources: MakeGeneratedResources(expectedOutputs),
		}, nil)

		// sanity check
		Expect(outputs).NotTo(Equal(expectedOutputs))

		err := exts.PatchMeshOutputs(ctx, tt, outputs)
		Expect(err).NotTo(HaveOccurred())

		// expect patches to be applied
		Expect(outputs).To(Equal(expectedOutputs))

	})
})
