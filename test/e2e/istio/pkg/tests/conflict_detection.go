package tests

import (
	"context"

	"github.com/golang/protobuf/proto"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/external-apis/pkg/api/istio/networking.istio.io/v1alpha3"
	commonv1 "github.com/solo-io/gloo-mesh/pkg/api/common.mesh.gloo.solo.io/v1"
	discoveryv1 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1"
	networkingv1 "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/test/e2e"
	"github.com/solo-io/gloo-mesh/test/utils"
	skv2corev1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	"github.com/solo-io/skv2/pkg/ezkube"
	istionetworkingv1alpha3spec "istio.io/api/networking/v1alpha3"
	istionetworkingv1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

/*
	This test assumes conflict detection is enabled at Gloo Mesh boot time using the "--disallow-intersecting-config=true" CLI flag.
*/
func ConflictDetectionTest() {
	var (
		err                 error
		manifest            utils.Manifest
		mgmtClient          client.Client
		remoteClient        client.Client
		userVirtualService  *istionetworkingv1alpha3.VirtualService
		userDestinationRule *istionetworkingv1alpha3.DestinationRule
	)

	BeforeEach(func() {
		mgmtClient, err = client.New(e2e.GetEnv().Management.Config, client.Options{})
		Expect(err).NotTo(HaveOccurred())

		remoteClient, err = client.New(e2e.GetEnv().Remote.Config, client.Options{})
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		manifest.Cleanup(BookinfoNamespace)
	})

	It("detects conflicts with user supplied VirtualServices", func() {
		manifest, err = utils.NewManifest("conflict_detection_virtual_service_test_manifest.yaml")
		Expect(err).ToNot(HaveOccurred())

		By("creating a user-supplied VirtualService", func() {
			remoteReviewsHostname, err := getDestinationLocalFqdn(
				mgmtClient,
				&skv2corev1.ObjectRef{Name: "reviews-bookinfo-remote-cluster", Namespace: "gloo-mesh"},
			)
			Expect(err).ToNot(HaveOccurred())

			userVirtualService = &istionetworkingv1alpha3.VirtualService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "user-supplied",
					Namespace: BookinfoNamespace,
				},
				Spec: istionetworkingv1alpha3spec.VirtualService{
					Hosts: []string{remoteReviewsHostname},
					Http: []*istionetworkingv1alpha3spec.HTTPRoute{
						{
							Route: []*istionetworkingv1alpha3spec.HTTPRouteDestination{
								{
									Destination: &istionetworkingv1alpha3spec.Destination{
										Host: "reviews.bookinfo.svc.cluster.local",
									},
								},
							},
						},
					},
				},
			}

			err = createVirtualService(remoteClient, userVirtualService)
			Expect(err).NotTo(HaveOccurred())
		})

		By("creating a TrafficPolicy that translates a conflicting VirtualService should not be output", func() {
			trafficPolicy := &networkingv1.TrafficPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "conflicting-traffic-policy",
					Namespace: BookinfoNamespace,
				},
				TypeMeta: metav1.TypeMeta{
					Kind:       "TrafficPolicy",
					APIVersion: networkingv1.SchemeGroupVersion.String(),
				},
				Spec: networkingv1.TrafficPolicySpec{
					SourceSelector: nil,
					DestinationSelector: []*commonv1.DestinationSelector{
						{
							KubeServiceRefs: &commonv1.DestinationSelector_KubeServiceRefs{
								Services: []*skv2corev1.ClusterObjectRef{
									{
										Name:        "reviews",
										Namespace:   BookinfoNamespace,
										ClusterName: RemoteClusterName,
									},
								},
							},
						},
					},
					Policy: &networkingv1.TrafficPolicySpec_Policy{
						Retries: &networkingv1.TrafficPolicySpec_Policy_RetryPolicy{
							Attempts: 5,
						},
					},
				},
			}

			err = manifest.AppendResources(trafficPolicy)
			Expect(err).NotTo(HaveOccurred())
			err = manifest.KubeApply(BookinfoNamespace)
			Expect(err).NotTo(HaveOccurred())

			virtualService, err := getVirtualService(remoteClient, ezkube.MakeObjectRef(userVirtualService))
			Expect(err).NotTo(HaveOccurred())

			Expect(proto.Equal(&virtualService.Spec, &userVirtualService.Spec)).To(BeTrue())
		})

		By("cleaning up the user VirtualService", func() {
			err = deleteVirtualService(remoteClient, ezkube.MakeObjectRef(userVirtualService))
			Expect(err).NotTo(HaveOccurred())
		})
	})

	It("detects conflicts with user supplied DestinationRules", func() {
		manifest, err = utils.NewManifest("conflict_detection_destination_rule_test_manifest.yaml")
		Expect(err).ToNot(HaveOccurred())

		By("creating a user-supplied DestinationRule should cause default translated DestinationRule to be deleted", func() {
			remoteReviewsHostname, err := getDestinationLocalFqdn(
				mgmtClient,
				&skv2corev1.ObjectRef{Name: "reviews-bookinfo-remote-cluster", Namespace: "gloo-mesh"},
			)
			Expect(err).ToNot(HaveOccurred())

			userDestinationRule = &istionetworkingv1alpha3.DestinationRule{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "user-supplied",
					Namespace: BookinfoNamespace,
				},
				Spec: istionetworkingv1alpha3spec.DestinationRule{
					Host: remoteReviewsHostname,
				},
			}

			err = createDestinationRule(remoteClient, userDestinationRule)
			Expect(err).NotTo(HaveOccurred())

			// eventually the default translated DestinationRule should be deleted
			Eventually(func() bool {
				_, err = getDestinationRule(remoteClient, &skv2corev1.ObjectRef{
					Name:      "reviews",
					Namespace: BookinfoNamespace,
				})
				return errors.IsNotFound(err)
			}, "10s", "1s").Should(BeTrue())
		})

		By("cleaning up the user DestinationRule should cause default DestinationRule to be recreated", func() {
			err = deleteDestinationRule(remoteClient, ezkube.MakeObjectRef(userDestinationRule))
			Expect(err).NotTo(HaveOccurred())

			// eventually the default translated DestinationRule should be deleted
			Eventually(func() error {
				_, err = getDestinationRule(remoteClient, &skv2corev1.ObjectRef{
					Name:      "reviews",
					Namespace: BookinfoNamespace,
				})
				return err
			}, "10s", "1s").Should(BeNil())
		})
	})
}

func getDestinationLocalFqdn(mgmtClient client.Client, ref *skv2corev1.ObjectRef) (string, error) {
	destinationClient := discoveryv1.NewDestinationClient(mgmtClient)
	destination, err := destinationClient.GetDestination(context.TODO(), client.ObjectKey{
		Name:      ref.Name,
		Namespace: ref.Namespace,
	})
	if err != nil {
		return "", err
	}
	return destination.Status.LocalFqdn, nil
}

func getVirtualService(remoteClient client.Client, ref *skv2corev1.ObjectRef) (*istionetworkingv1alpha3.VirtualService, error) {
	virtualServiceClient := v1alpha3.NewVirtualServiceClient(remoteClient)
	return virtualServiceClient.GetVirtualService(context.TODO(), client.ObjectKey{
		Name:      ref.Name,
		Namespace: ref.Namespace,
	})
}

func createVirtualService(remoteClient client.Client, virtualService *istionetworkingv1alpha3.VirtualService) error {
	virtualServiceClient := v1alpha3.NewVirtualServiceClient(remoteClient)
	return virtualServiceClient.CreateVirtualService(context.TODO(), virtualService)
}

func deleteVirtualService(remoteClient client.Client, ref *skv2corev1.ObjectRef) error {
	virtualServiceClient := v1alpha3.NewVirtualServiceClient(remoteClient)
	return virtualServiceClient.DeleteVirtualService(context.TODO(), client.ObjectKey{
		Name:      ref.Name,
		Namespace: ref.Namespace,
	})
}

func getDestinationRule(remoteClient client.Client, ref *skv2corev1.ObjectRef) (*istionetworkingv1alpha3.DestinationRule, error) {
	destinationRuleClient := v1alpha3.NewDestinationRuleClient(remoteClient)
	return destinationRuleClient.GetDestinationRule(context.TODO(), client.ObjectKey{
		Name:      ref.Name,
		Namespace: ref.Namespace,
	})
}

func createDestinationRule(remoteClient client.Client, destinationRule *istionetworkingv1alpha3.DestinationRule) error {
	destinationRuleClient := v1alpha3.NewDestinationRuleClient(remoteClient)
	return destinationRuleClient.CreateDestinationRule(context.TODO(), destinationRule)
}

func deleteDestinationRule(remoteClient client.Client, ref *skv2corev1.ObjectRef) error {
	destinationRuleClient := v1alpha3.NewDestinationRuleClient(remoteClient)
	return destinationRuleClient.DeleteDestinationRule(context.TODO(), client.ObjectKey{
		Name:      ref.Name,
		Namespace: ref.Namespace,
	})
}
