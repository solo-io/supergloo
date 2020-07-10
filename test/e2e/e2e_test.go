package e2e_test

import (
	"context"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	v1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	"github.com/solo-io/smh/test/e2e"
	"io/ioutil"
	"path/filepath"
	"time"

	"github.com/solo-io/smh/test/data"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/solo-io/skv2/codegen/util"

	"github.com/solo-io/skv2/codegen/model"
	"github.com/solo-io/skv2/codegen/render"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	policyName      = "bookinfo-policy"
	policyService   = "reviews"
	policyNamespace = "bookinfo"
	policyCluster   = "management-cluster-1"
	policyManifest  = "test/e2e/bookinfo-policies.yaml"
	username        = "petlover"
	password        = "ilovepets"

	trafficShiftReviewsV2 = data.TrafficShiftPolicy(policyName, policyNamespace, &v1.ClusterObjectRef{
		Name:        policyService,
		Namespace:   policyNamespace,
		ClusterName: policyCluster,
	}, map[string]string{"version": "v2"}, 9080)

	dynamicClient client.Client

	curlReviews = func() string {
		env := e2e.GetEnv()
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute/3)
		defer cancel()
		out := env.Management.GetPod("default", "productpage").Curl(ctx, "http://reviews:9080/reviews/1", "-v")
		GinkgoWriter.Write([]byte(out))
		return out
	}
)

// Must run `make generated-code` before running this test
var _ = Describe("SMH E2e", func() {
	BeforeEach(func() {

		_ = util.Kubectl(nil, "create", "ns", policyNamespace)

		err := writeTestManifest(policyManifest)
		Expect(err).NotTo(HaveOccurred())

		err = util.Kubectl(nil, "apply", "-n="+policyNamespace, "-f="+policyManifest)
		Expect(err).NotTo(HaveOccurred())

		cfg, err := e2e.GetEnv().Management.Config.ClientConfig()
		Expect(err).NotTo(HaveOccurred())

		dynamicClient, err = client.New(cfg, client.Options{})
		Expect(err).NotTo(HaveOccurred())

	})

	AfterEach(func() {
		err := util.Kubectl(nil, "delete", "ns", policyNamespace)
		Expect(err).NotTo(HaveOccurred())
	})

	It("sets status properly on all API resources", func() {
		assertCrdStatuses() // TODO all crd types
	})

	It("applies TrafficShift policies to local subsets", func() {

		// first check that we have a response to reduce flakiness
		Eventually(curlReviews, "1m", "1s").Should(ContainSubstring(`"color": "black"`))
	})
})

func assertCrdStatuses() {
	assertTrafficPolicyStatuses()
}

func assertTrafficPolicyStatuses() {
	ctx := context.Background()
	trafficPolicy := v1alpha1.NewTrafficPolicyClient(dynamicClient)

	EventuallyWithOffset(2, func() bool {
		list, err := trafficPolicy.ListTrafficPolicy(ctx, client.InNamespace(policyNamespace))
		ExpectWithOffset(2, err).NotTo(HaveOccurred())
		ExpectWithOffset(2, list.Items).To(HaveLen(2))
		for _, apiProduct := range list.Items {
			if apiProduct.Status.ObservedGeneration == 0 {
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
