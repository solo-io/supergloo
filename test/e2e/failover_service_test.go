package e2e_test

import (
	"context"
	"fmt"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	networkingv1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/test/data"
	"github.com/solo-io/service-mesh-hub/test/e2e"
	"github.com/solo-io/service-mesh-hub/test/utils"
	v1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("FailoverService", func() {
	var (
		err      error
		manifest utils.Manifest
		ctx      = context.Background()
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
			virtualMesh := data.SelfSignedVirtualMesh(
				"bookinfo-federation",
				BookinfoNamespace,
				[]*v1.ObjectRef{
					masterMesh,
					remoteMesh,
				})
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
										ClusterName: masterClusterName,
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
				ObjectMeta: metav1.ObjectMeta{
					Name:      "reviews-failover",
					Namespace: BookinfoNamespace,
				},
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
									ClusterName: masterClusterName,
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
			err = manifest.AppendResources(virtualMesh)
			Expect(err).NotTo(HaveOccurred())
			err = manifest.KubeApply(BookinfoNamespace)
			Expect(err).NotTo(HaveOccurred())

			// ensure status is updated
			assertVirtualMeshStatuses()
			// check we can hit the remote service
			// give 5 minutes because the workflow depends on restarting pods
			// which can take several minutes
			Eventually(curlRemoteReviews, "10m", "2s").Should(ContainSubstring("200 OK"))

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
			Eventually(curlFailoverService, "1m", "1s").Should(ContainSubstring(`"color": "red"`))
			// now check that it is consistent 10 times in a row
			Eventually(func() bool {
				for i := 0; i < 5; i++ {
					if !strings.Contains(curlFailoverService(), `"color": "red"`) {
						return false
					}
					time.Sleep(2 * time.Second)
				}
				return true
			}, "5m", "2s").Should(BeTrue())
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
