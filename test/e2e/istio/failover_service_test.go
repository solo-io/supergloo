package istio_test

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	networkingv1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/test/e2e"
	"github.com/solo-io/service-mesh-hub/test/utils"
	v1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("FailoverService", func() {
	var (
		err                    error
		manifest               utils.Manifest
		ctx                    = context.Background()
		failoverServiceObjMeta = metav1.ObjectMeta{
			Name:      "reviews-failover",
			Namespace: BookinfoNamespace,
		}
	)

	AfterEach(func() {
		manifest.Cleanup(BookinfoNamespace)
		// Ensure restoring bookinfo containers if test fails.
		env := e2e.GetEnv()
		env.Management.EnableContainer(ctx, BookinfoNamespace, "reviews-v1")
		env.Management.EnableContainer(ctx, BookinfoNamespace, "reviews-v2")
		env.Management.WaitForRollout(ctx, BookinfoNamespace, "reviews-v1")
		env.Management.WaitForRollout(ctx, BookinfoNamespace, "reviews-v2")
	})

	It("should create a failover service", func() {
		manifest, err = utils.NewManifest("failover_service_test_manifest.yaml")
		Expect(err).ToNot(HaveOccurred())
		env := e2e.GetEnv()

		failoverServiceHostname := "reviews-failover.bookinfo.global"
		curlFailoverService := func() string {
			return curlFromProductpage(fmt.Sprintf("http://%s:9080/reviews/1", failoverServiceHostname))
		}

		By("creating a new FailoverService with the prerequisite TrafficPolicy and VirtualMesh", func() {
			trafficPolicy := &networkingv1alpha2.TrafficPolicy{
				TypeMeta: metav1.TypeMeta{
					Kind:       "TrafficPolicy",
					APIVersion: networkingv1alpha2.SchemeGroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "reviews-outlier-detection",
					Namespace: BookinfoNamespace,
				},
				Spec: networkingv1alpha2.TrafficPolicySpec{
					DestinationSelector: []*networkingv1alpha2.ServiceSelector{
						{
							KubeServiceRefs: &networkingv1alpha2.ServiceSelector_KubeServiceRefs{
								Services: []*v1.ClusterObjectRef{
									{
										Name:        "reviews",
										Namespace:   BookinfoNamespace,
										ClusterName: mgmtClusterName,
									},
									{
										Name:        "reviews",
										Namespace:   BookinfoNamespace,
										ClusterName: remoteClusterName,
									},
								},
							},
						},
					},
					OutlierDetection: &networkingv1alpha2.TrafficPolicySpec_OutlierDetection{
						ConsecutiveErrors: 1,
					},
				},
			}
			failoverService := &networkingv1alpha2.FailoverService{
				TypeMeta: metav1.TypeMeta{
					Kind:       "FailoverService",
					APIVersion: networkingv1alpha2.SchemeGroupVersion.String(),
				},
				ObjectMeta: failoverServiceObjMeta,
				Spec: networkingv1alpha2.FailoverServiceSpec{
					Hostname: failoverServiceHostname,
					Port: &networkingv1alpha2.FailoverServiceSpec_Port{
						Number:   9080,
						Protocol: "http",
					},
					Meshes: []*v1.ObjectRef{
						masterMesh,
					},
					BackingServices: []*networkingv1alpha2.FailoverServiceSpec_BackingService{
						{
							BackingServiceType: &networkingv1alpha2.FailoverServiceSpec_BackingService_KubeService{
								KubeService: &v1.ClusterObjectRef{
									Name:        "reviews",
									Namespace:   BookinfoNamespace,
									ClusterName: mgmtClusterName,
								},
							},
						},
						{
							BackingServiceType: &networkingv1alpha2.FailoverServiceSpec_BackingService_KubeService{
								KubeService: &v1.ClusterObjectRef{
									Name:        "reviews",
									Namespace:   BookinfoNamespace,
									ClusterName: remoteClusterName,
								},
							},
						},
					},
				},
			}

			err := manifest.AppendResources(trafficPolicy)
			Expect(err).NotTo(HaveOccurred())
			err = manifest.KubeApply(BookinfoNamespace)
			Expect(err).NotTo(HaveOccurred())
			// Wait for TrafficPolicy with outlier detection to be processed before creating FailoverService.
			utils.AssertTrafficPolicyStatuses(dynamicClient, BookinfoNamespace)

			err = manifest.AppendResources(failoverService)
			Expect(err).NotTo(HaveOccurred())
			err = manifest.KubeApply(BookinfoNamespace)
			Expect(err).NotTo(HaveOccurred())

			// Make it failover to remote cluster with reviews-v3, disable all master cluster reviews pods to prove
			// that request is being served by remote cluster
			env.Management.DisableContainer(ctx, BookinfoNamespace, "reviews-v1", "reviews")
			env.Management.DisableContainer(ctx, BookinfoNamespace, "reviews-v2", "reviews")
			env.Management.WaitForRollout(ctx, BookinfoNamespace, "reviews-v1")
			env.Management.WaitForRollout(ctx, BookinfoNamespace, "reviews-v2")

			// first check that we have a response to reduce flakiness
			Eventually(curlFailoverService, "1m", "1s").Should(ContainSubstring("200 OK"))
		})

		By("setting a TrafficShift to redirect traffic to the FailoverService", func() {
			trafficPolicy := &networkingv1alpha2.TrafficPolicy{
				TypeMeta: metav1.TypeMeta{
					Kind:       "TrafficPolicy",
					APIVersion: networkingv1alpha2.SchemeGroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "reviews-shift-failover",
					Namespace: BookinfoNamespace,
				},
				Spec: networkingv1alpha2.TrafficPolicySpec{
					DestinationSelector: []*networkingv1alpha2.ServiceSelector{
						{
							KubeServiceRefs: &networkingv1alpha2.ServiceSelector_KubeServiceRefs{
								Services: []*v1.ClusterObjectRef{
									{
										Name:        "reviews",
										Namespace:   BookinfoNamespace,
										ClusterName: mgmtClusterName,
									},
								},
							},
						},
					},
					TrafficShift: &networkingv1alpha2.TrafficPolicySpec_MultiDestination{
						Destinations: []*networkingv1alpha2.TrafficPolicySpec_MultiDestination_WeightedDestination{
							{
								DestinationType: &networkingv1alpha2.TrafficPolicySpec_MultiDestination_WeightedDestination_FailoverService{
									FailoverService: &networkingv1alpha2.TrafficPolicySpec_MultiDestination_WeightedDestination_FailoverServiceDestination{
										Name:      failoverServiceObjMeta.Name,
										Namespace: failoverServiceObjMeta.Namespace,
									},
								},
							},
						},
					},
				},
			}
			err := manifest.AppendResources(trafficPolicy)
			Expect(err).NotTo(HaveOccurred())
			err = manifest.KubeApply(BookinfoNamespace)
			Expect(err).NotTo(HaveOccurred())
			utils.AssertTrafficPolicyStatuses(dynamicClient, BookinfoNamespace)

			// reviews-v3 is only deployed on remote cluster, so receiving a response proves that the FailoverService is working
			Eventually(curlReviews, "1m", "1s").Should(ContainSubstring(`"color": "red"`))
		})

		By("re-enable management-plane reviews deployments", func() {
			err := manifest.KubeDelete(BookinfoNamespace)
			Expect(err).NotTo(HaveOccurred())

			env.Management.EnableContainer(ctx, BookinfoNamespace, "reviews-v1")
			env.Management.EnableContainer(ctx, BookinfoNamespace, "reviews-v2")
			env.Management.WaitForRollout(ctx, BookinfoNamespace, "reviews-v1")
			env.Management.WaitForRollout(ctx, BookinfoNamespace, "reviews-v2")
			Eventually(curlReviews, "1m", "1s").Should(ContainSubstring("200 OK"))
		})
	})
})
