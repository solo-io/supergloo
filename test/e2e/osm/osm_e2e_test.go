package osm_test

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	commonv1 "github.com/solo-io/gloo-mesh/pkg/api/common.mesh.gloo.solo.io/v1"
	networkingv1 "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/test/e2e"
	"github.com/solo-io/gloo-mesh/test/utils"
	skv2corev1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Shared test vars
var (
	bookThief = "bookthief"
	bookStore = "bookstore"

	mgmtClusterName = "mgmt-cluster"

	curlBookstore = func() string {
		return curlFromBookthief("http://bookstore.bookstore")
	}

	curlFromBookthief = func(url string) string {
		env := e2e.GetSingleClusterEnv()
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute/3)
		defer cancel()
		out := env.Management.GetPod(ctx, bookThief, bookThief).Curl(ctx, url, "-v")
		GinkgoWriter.Write([]byte(out))
		return out
	}
)

var _ = Describe("OsmE2e", func() {

	It("works", func() {
		manifest, err := utils.NewManifest("osm_test.yaml")
		Expect(err).ToNot(HaveOccurred())

		trafficPolicy := &networkingv1.TrafficPolicy{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-policy",
			},
			TypeMeta: metav1.TypeMeta{
				Kind:       "TrafficPolicy",
				APIVersion: networkingv1.SchemeGroupVersion.String(),
			},
			Spec: networkingv1.TrafficPolicySpec{
				DestinationSelector: []*commonv1.DestinationSelector{
					{
						KubeServiceRefs: &commonv1.DestinationSelector_KubeServiceRefs{
							Services: []*skv2corev1.ClusterObjectRef{
								{
									Name:        bookStore,
									Namespace:   bookStore,
									ClusterName: mgmtClusterName,
								},
							},
						},
					},
				},
				Policy: &networkingv1.TrafficPolicySpec_Policy{
					TrafficShift: &networkingv1.TrafficPolicySpec_Policy_MultiDestination{
						Destinations: []*networkingv1.TrafficPolicySpec_Policy_MultiDestination_WeightedDestination{
							{
								DestinationType: &networkingv1.TrafficPolicySpec_Policy_MultiDestination_WeightedDestination_KubeService{
									KubeService: &networkingv1.TrafficPolicySpec_Policy_MultiDestination_WeightedDestination_KubeDestination{
										Name:        fmt.Sprintf("%s-v1", bookStore),
										Namespace:   bookStore,
										ClusterName: mgmtClusterName,
									},
								},
								Weight: 50,
							},
							{
								DestinationType: &networkingv1.TrafficPolicySpec_Policy_MultiDestination_WeightedDestination_KubeService{
									KubeService: &networkingv1.TrafficPolicySpec_Policy_MultiDestination_WeightedDestination_KubeDestination{
										Name:        fmt.Sprintf("%s-v2", bookStore),
										Namespace:   bookStore,
										ClusterName: mgmtClusterName,
									},
								},
								Weight: 50,
							},
						},
					},
				},
			},
		}
		accessPolicy := &networkingv1.AccessPolicy{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-policy",
			},
			TypeMeta: metav1.TypeMeta{
				Kind:       "AccessPolicy",
				APIVersion: networkingv1.SchemeGroupVersion.String(),
			},
			Spec: networkingv1.AccessPolicySpec{
				SourceSelector: []*commonv1.IdentitySelector{
					{
						KubeServiceAccountRefs: &commonv1.IdentitySelector_KubeServiceAccountRefs{
							ServiceAccounts: []*skv2corev1.ClusterObjectRef{
								{
									Name:        bookThief,
									Namespace:   bookThief,
									ClusterName: mgmtClusterName,
								},
							},
						},
					},
				},
				DestinationSelector: []*commonv1.DestinationSelector{
					{
						KubeServiceRefs: &commonv1.DestinationSelector_KubeServiceRefs{
							Services: []*skv2corev1.ClusterObjectRef{
								{
									Name:        fmt.Sprintf("%s-v1", bookStore),
									Namespace:   bookStore,
									ClusterName: mgmtClusterName,
								},
								{
									Name:        fmt.Sprintf("%s-v2", bookStore),
									Namespace:   bookStore,
									ClusterName: mgmtClusterName,
								},
							},
						},
					},
				},
			},
		}

		err = manifest.AppendResources(trafficPolicy, accessPolicy)
		Expect(err).NotTo(HaveOccurred())

		err = manifest.KubeApply("gloo-mesh")
		Expect(err).NotTo(HaveOccurred())

		// check basic success
		Eventually(curlBookstore, "5m", "1s").Should(ContainSubstring("200 OK"))
		// check v1
		Eventually(curlBookstore, "5m", "1s").Should(ContainSubstring("identity: bookstore-v1"))
		// check v2
		Eventually(curlBookstore, "5m", "1s").Should(ContainSubstring("identity: bookstore-v2"))

	})

})
