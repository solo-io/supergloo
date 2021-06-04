package tests

import (
	"context"
	"fmt"
	"net/http"

	commonv1 "github.com/solo-io/gloo-mesh/pkg/api/common.mesh.gloo.solo.io/v1"
	discoveryv1 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1"
	networkingv1 "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/hostutils"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/metautils"
	"github.com/solo-io/gloo-mesh/test/data"
	"github.com/solo-io/gloo-mesh/test/e2e"
	"github.com/solo-io/gloo-mesh/test/utils"
	skv2corev1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	istionetworkingv1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func FederationTest() {
	var (
		err      error
		manifest utils.Manifest
		ctx      = context.Background()
	)

	AfterEach(func() {
		manifest.Cleanup(BookinfoNamespace)
	})

	/*
		These tests assume that federation has been established between mgmt and remote clusters.
	*/

	It("enables communication across clusters using global dns names", func() {
		manifest, err = utils.NewManifest("federation-trafficpolicies.yaml")
		Expect(err).NotTo(HaveOccurred())

		By("with federation enabled, TrafficShifts can be used for subsets across meshes ", func() {
			// create cross cluster traffic shift
			trafficShiftReviewsV3 := data.RemoteTrafficShiftPolicy("bookinfo-policy", BookinfoNamespace, &skv2corev1.ClusterObjectRef{
				Name:        "reviews",
				Namespace:   BookinfoNamespace,
				ClusterName: MgmtClusterName,
			}, RemoteClusterName, map[string]string{"version": "v3"}, 9080)

			err = manifest.AppendResources(trafficShiftReviewsV3)
			Expect(err).NotTo(HaveOccurred())
			err = manifest.KubeApply(BookinfoNamespace)
			Expect(err).NotTo(HaveOccurred())

			// ensure status is updated
			utils.AssertTrafficPolicyStatuses(ctx, e2e.GetEnv().Management.TrafficPolicyClient, BookinfoNamespace)

			// check we can eventually hit the v3 subset
			Eventually(CurlReviews, "30s", "1s").Should(ContainSubstring(`"color": "red"`))
			Consistently(CurlReviews).Should(ContainSubstring(`"color": "red"`))
		})
	})

	It("should output a VirtualService for the federated ServiceEntry", func() {
		By("fault injection should be applied when sending requests to a federated ServiceEntry", func() {
			manifest, err = utils.NewManifest("federation-trafficpolicies.yaml")
			Expect(err).NotTo(HaveOccurred())

			// Create TrafficPolicy with fault injection applied to remote cluster
			faultInjectionPolicy := &networkingv1.TrafficPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "remote-fault-injection",
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
						FaultInjection: &networkingv1.TrafficPolicySpec_Policy_FaultInjection{
							FaultInjectionType: &networkingv1.TrafficPolicySpec_Policy_FaultInjection_Abort_{
								&networkingv1.TrafficPolicySpec_Policy_FaultInjection_Abort{
									HttpStatus: http.StatusTeapot,
								},
							},
							Percentage: 100,
						},
					},
				},
			}

			err = manifest.AppendResources(faultInjectionPolicy)
			Expect(err).NotTo(HaveOccurred())
			err = manifest.KubeApply(BookinfoNamespace)
			Expect(err).NotTo(HaveOccurred())

			// ensure status is updated
			utils.AssertTrafficPolicyStatuses(ctx, e2e.GetEnv().Management.TrafficPolicyClient, BookinfoNamespace)

			Eventually(CurlRemoteReviews(hostutils.GetFederatedHostnameSuffix(&VirtualMesh.Spec)), "30s", "1s").Should(ContainSubstring("418"))

			// Delete TrafficPolicy so it doesn't interfere with subsequent tests
			manifest.KubeDelete(BookinfoNamespace)
		})

		By("traffic mirrors and shifts should use correct hostname for federated ServiceEntry", func() {
			manifest, err = utils.NewManifest("federation-trafficpolicies.yaml")
			Expect(err).NotTo(HaveOccurred())

			// Create TrafficPolicy with fault injection applied to remote cluster
			trafficPolicy := &networkingv1.TrafficPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "remote-mirror-and-shift",
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
						Mirror: &networkingv1.TrafficPolicySpec_Policy_Mirror{
							DestinationType: &networkingv1.TrafficPolicySpec_Policy_Mirror_KubeService{
								KubeService: &skv2corev1.ClusterObjectRef{
									Name:        "reviews",
									Namespace:   BookinfoNamespace,
									ClusterName: MgmtClusterName,
								},
							},
							Percentage: 50,
							Port:       9080,
						},
						TrafficShift: &networkingv1.TrafficPolicySpec_Policy_MultiDestination{
							Destinations: []*networkingv1.TrafficPolicySpec_Policy_MultiDestination_WeightedDestination{
								{
									DestinationType: &networkingv1.TrafficPolicySpec_Policy_MultiDestination_WeightedDestination_KubeService{
										KubeService: &networkingv1.TrafficPolicySpec_Policy_MultiDestination_WeightedDestination_KubeDestination{
											Name:        "reviews",
											Namespace:   BookinfoNamespace,
											ClusterName: MgmtClusterName,
										},
									},
									Weight: 50,
								},
							},
						},
					},
				},
			}

			err = manifest.AppendResources(trafficPolicy)
			Expect(err).NotTo(HaveOccurred())
			err = manifest.KubeApply(BookinfoNamespace)
			Expect(err).NotTo(HaveOccurred())

			// ensure status is updated
			utils.AssertTrafficPolicyStatuses(ctx, e2e.GetEnv().Management.TrafficPolicyClient, BookinfoNamespace)

			var getFederatedVirtualService = func() (*istionetworkingv1alpha3.VirtualService, error) {
				env := e2e.GetEnv()
				vsClient := env.Management.VirtualServiceClient
				meta := metautils.FederatedObjectMeta(
					&metav1.ObjectMeta{
						Name:        "reviews",
						Namespace:   BookinfoNamespace,
						ClusterName: RemoteClusterName,
					},
					&discoveryv1.MeshSpec_MeshInstallation{
						Namespace: "istio-system",
						Cluster:   MgmtClusterName,
					},
					nil,
				)
				return vsClient.GetVirtualService(ctx, client.ObjectKey{Name: meta.Name, Namespace: meta.Namespace})
			}

			Eventually(func() bool {
				virtualService, err := getFederatedVirtualService()
				if err != nil {
					return false
				}
				if len(virtualService.Spec.Http) > 0 {
					httpRoute := virtualService.Spec.Http[0]
					if httpRoute.GetMirror().GetHost() != fmt.Sprintf("reviews.%s.svc.cluster.local", BookinfoNamespace) {
						return false
					}
					if len(httpRoute.GetRoute()) < 1 {
						return false
					}
					if httpRoute.GetRoute()[0].GetDestination().GetHost() != fmt.Sprintf("reviews.%s.svc.cluster.local", BookinfoNamespace) {
						return false
					}
					return true
				}
				return false
			}, "30s", "1s").Should(BeTrue())
		})
	})
}
