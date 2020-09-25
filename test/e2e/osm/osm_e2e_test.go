package osm_test

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/test/e2e"
	"github.com/solo-io/service-mesh-hub/test/utils"
	v1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
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

		trafficPolicy := &v1alpha2.TrafficPolicy{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-policy",
			},
			TypeMeta: metav1.TypeMeta{
				Kind:       "TrafficPolicy",
				APIVersion: v1alpha2.SchemeGroupVersion.String(),
			},
			Spec: v1alpha2.TrafficPolicySpec{
				DestinationSelector: []*v1alpha2.TrafficTargetSelector{
					{
						KubeServiceRefs: &v1alpha2.TrafficTargetSelector_KubeServiceRefs{
							Services: []*v1.ClusterObjectRef{
								{
									Name:        bookStore,
									Namespace:   bookStore,
									ClusterName: mgmtClusterName,
								},
							},
						},
					},
				},
				TrafficShift: &v1alpha2.TrafficPolicySpec_MultiDestination{
					Destinations: []*v1alpha2.TrafficPolicySpec_MultiDestination_WeightedDestination{
						{
							DestinationType: &v1alpha2.TrafficPolicySpec_MultiDestination_WeightedDestination_KubeService{
								KubeService: &v1alpha2.TrafficPolicySpec_MultiDestination_WeightedDestination_KubeDestination{
									Name:        fmt.Sprintf("%s-v1", bookStore),
									Namespace:   bookStore,
									ClusterName: mgmtClusterName,
								},
							},
							Weight: 50,
						},
						{
							DestinationType: &v1alpha2.TrafficPolicySpec_MultiDestination_WeightedDestination_KubeService{
								KubeService: &v1alpha2.TrafficPolicySpec_MultiDestination_WeightedDestination_KubeDestination{
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
		}
		accessPolicy := &v1alpha2.AccessPolicy{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-policy",
			},
			TypeMeta: metav1.TypeMeta{
				Kind:       "AccessPolicy",
				APIVersion: v1alpha2.SchemeGroupVersion.String(),
			},
			Spec: v1alpha2.AccessPolicySpec{
				SourceSelector: []*v1alpha2.IdentitySelector{
					{
						KubeServiceAccountRefs: &v1alpha2.IdentitySelector_KubeServiceAccountRefs{
							ServiceAccounts: []*v1.ClusterObjectRef{
								{
									Name:        bookThief,
									Namespace:   bookThief,
									ClusterName: mgmtClusterName,
								},
							},
						},
					},
				},
				DestinationSelector: []*v1alpha2.TrafficTargetSelector{
					{
						KubeServiceRefs: &v1alpha2.TrafficTargetSelector_KubeServiceRefs{
							Services: []*v1.ClusterObjectRef{
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

		err = manifest.KubeApply("service-mesh-hub")
		Expect(err).NotTo(HaveOccurred())

		// check basic success
		Eventually(curlBookstore, "5m", "1s").Should(ContainSubstring("200 OK"))
		// check v1
		Eventually(curlBookstore, "5m", "1s").Should(ContainSubstring("identity: bookstore-v1"))
		// check v2
		Eventually(curlBookstore, "5m", "1s").Should(ContainSubstring("identity: bookstore-v2"))

	})

})
