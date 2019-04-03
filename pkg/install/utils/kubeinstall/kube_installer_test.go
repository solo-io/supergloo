package kubeinstall_test

import (
	"context"

	"github.com/solo-io/supergloo/pkg/install/utils/kuberesource"
	appsv1 "k8s.io/api/apps/v1"
	appsv1beta2 "k8s.io/api/apps/v1beta2"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/solo-io/supergloo/pkg/install/utils/kubeinstall"

	"github.com/solo-io/go-utils/kubeutils"
	"github.com/solo-io/supergloo/pkg/install/utils/helmchart"

	superglootest "github.com/solo-io/supergloo/test/testutils"

	"github.com/solo-io/go-utils/testutils"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	// Needed to run tests in GKE
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var istioCrd = apiextensions.CustomResourceDefinition{}

var (
	kubeClient kubernetes.Interface
)

var _ = Describe("KubeInstaller", func() {
	var (
		ns string
	)
	BeforeEach(func() {
		kubeClient = superglootest.MustKubeClient()
		// wait for all services in the previous namespace to be torn down
		// important because of a race caused by nodeport conflcit
		if ns != "" {
			superglootest.WaitForIstioTeardown(ns)
		}
		ns = "test" + testutils.RandString(5)
		testutils.SetupKubeForTest(ns)
	})
	AfterEach(func() {
		testutils.TeardownKube(ns)
		superglootest.TeardownWithPrefix(kubeClient, "istio")
		superglootest.TeardownWithPrefix(kubeClient, "prometheus")
		superglootest.WaitForIstioTeardown(ns)
	})
	Context("create manifest", func() {
		It("creates resources from a helm chart", func() {
			values := `
mixer:
  enabled: true #should install mixer

`
			manifests, err := helmchart.RenderManifests(
				context.TODO(),
				"https://s3.amazonaws.com/supergloo.solo.io/istio-1.0.3.tgz",
				values,
				"aaa",
				ns,
				"",
				true,
			)
			Expect(err).NotTo(HaveOccurred())

			restCfg, err := kubeutils.GetConfig("", "")
			Expect(err).NotTo(HaveOccurred())
			cache := NewCache()
			err = cache.Init(context.TODO(), restCfg)
			Expect(err).NotTo(HaveOccurred())
			inst, err := NewKubeInstaller(restCfg, cache)
			Expect(err).NotTo(HaveOccurred())

			resources, err := manifests.ResourceList()
			Expect(err).NotTo(HaveOccurred())

			uniqueLabels := map[string]string{"unique": "setoflabels"}
			err = inst.ReconcilleResources(context.TODO(), ns, resources, uniqueLabels)
			Expect(err).NotTo(HaveOccurred())

			genericClient, err := client.New(restCfg, client.Options{})
			Expect(err).NotTo(HaveOccurred())
			// expect each resource to exist
			for _, resource := range resources {
				err := genericClient.Get(context.TODO(), client.ObjectKey{resource.GetNamespace(), resource.GetName()}, resource)
				Expect(err).NotTo(HaveOccurred())
				if resource.Object["kind"] == "Deployment" {
					// ensure all deployments have at least one ready replica
					deployment, err := kuberesource.ConvertUnstructured(resource)
					Expect(err).NotTo(HaveOccurred())
					switch dep := deployment.(type) {
					case *appsv1.Deployment:
						Expect(dep.Status.ReadyReplicas).To(BeNumerically(">=", 1))
					case *extensionsv1beta1.Deployment:
						Expect(dep.Status.ReadyReplicas).To(BeNumerically(">=", 1))
					case *appsv1beta2.Deployment:
						Expect(dep.Status.ReadyReplicas).To(BeNumerically(">=", 1))
					}
				}
			}
			// expect the mixer deployments to be created
			_, err = kubeClient.AppsV1().Deployments(ns).Get("istio-policy", v1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())
			_, err = kubeClient.AppsV1().Deployments(ns).Get("istio-telemetry", v1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())

			err = inst.PurgeResources(context.TODO(), uniqueLabels)
			Expect(err).NotTo(HaveOccurred())

			// uninstalled
			_, err = kubeClient.AppsV1().Deployments(ns).Get("istio-policy", v1.GetOptions{})
			Expect(err).To(HaveOccurred())
			_, err = kubeClient.AppsV1().Deployments(ns).Get("istio-telemetry", v1.GetOptions{})
			Expect(err).To(HaveOccurred())
		})
	})
})
