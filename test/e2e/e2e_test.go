package e2e_test

import (
	"context"
	"io/ioutil"
	"path/filepath"
	"time"

	"github.com/solo-io/go-utils/testutils"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha2"

	"github.com/solo-io/service-mesh-hub/test/e2e"
	v1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"

	"github.com/solo-io/service-mesh-hub/test/data"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/solo-io/skv2/codegen/util"

	"github.com/solo-io/skv2/codegen/model"
	"github.com/solo-io/skv2/codegen/render"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	policyName     = "bookinfo-policy"
	policyService  = "reviews"
	appNamespace   = "bookinfo"
	policyCluster  = "master-cluster"
	policyManifest = "test/e2e/bookinfo-policies.yaml"

	trafficShiftReviewsV2 = data.TrafficShiftPolicy(policyName, appNamespace, &v1.ClusterObjectRef{
		Name:        policyService,
		Namespace:   appNamespace,
		ClusterName: policyCluster,
	}, map[string]string{"version": "v2"}, 9080)

	dynamicClient client.Client

	curlReviews = func() string {
		env := e2e.GetEnv()
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute/3)
		defer cancel()
		out := env.Management.GetPod(appNamespace, "productpage").Curl(ctx, "http://reviews:9080/reviews/1", "-v")
		GinkgoWriter.Write([]byte(out))
		return out
	}
)

// Must run `make generated-code` before running this test
var _ = Describe("SMH E2e", func() {
	BeforeEach(func() {

		err := writeTestManifest(policyManifest)
		Expect(err).NotTo(HaveOccurred())

		dynamicClient, err = client.New(e2e.GetEnv().Management.Config, client.Options{})
		Expect(err).NotTo(HaveOccurred())

	})

	AfterEach(func() {
		err := testutils.Kubectl("delete", "-f", policyManifest)
		Expect(err).NotTo(HaveOccurred())
	})

	It("applies TrafficShift policies to local subsets", func() {

		// first check that we can hit both subsets
		Eventually(curlReviews, "1m", "1s").Should(ContainSubstring(`"color": "black"`))
		Eventually(curlReviews, "1m", "1s").Should(ContainSubstring(`"color": "red"`))

		err := testutils.Kubectl("apply", "-n="+appNamespace, "-f="+policyManifest)
		Expect(err).NotTo(HaveOccurred())

		// ensure status is updated
		assertTrafficPolicyStatuses()

		// check we consistently hit the v2 subset
		Consistently(curlReviews, "10s", "0.1s").Should(ContainSubstring(`"color": "black"`))
	})
})

func assertCrdStatuses() {

	err := testutils.Kubectl("apply", "-n="+appNamespace, "-f="+policyManifest)
	Expect(err).NotTo(HaveOccurred())

	assertTrafficPolicyStatuses()
}

func assertTrafficPolicyStatuses() {
	ctx := context.Background()
	trafficPolicy := v1alpha2.NewTrafficPolicyClient(dynamicClient)

	EventuallyWithOffset(1, func() bool {
		list, err := trafficPolicy.ListTrafficPolicy(ctx, client.InNamespace(appNamespace))
		ExpectWithOffset(1, err).NotTo(HaveOccurred())
		ExpectWithOffset(1, list.Items).To(HaveLen(1))
		for _, policy := range list.Items {
			if policy.Status.ObservedGeneration == 0 {
				return false
			}
		}
		return true
	}, time.Second*20).Should(BeTrue())
}

func writeTestManifest(manifestFile string) error {
	objs := []metav1.Object{
		trafficShiftReviewsV2,
	}
	return writeResourcesToManifest(objs, manifestFile)
}

func writeResourcesToManifest(resources []metav1.Object, filename string) error {
	// use skv2 libraries to write the resources as yaml
	manifest, err := render.ManifestsRenderer{
		AppName: "bookinfo-policies",
		ResourceFuncs: map[render.OutFile]render.MakeResourceFunc{
			render.OutFile{}: func(group render.Group) []metav1.Object {
				return resources
			},
		},
	}.RenderManifests(model.Group{RenderManifests: true})
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(filepath.Join(filepath.Dir(util.GoModPath()), filename), []byte(manifest[0].Content), 0644); err != nil {
		return err
	}

	return nil
}
