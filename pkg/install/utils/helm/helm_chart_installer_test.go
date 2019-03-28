package helm_test

import (
	"context"
	"strings"

	superglootest "github.com/solo-io/supergloo/test/testutils"

	"github.com/solo-io/go-utils/testutils"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	// Needed to run tests in GKE
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/solo-io/supergloo/pkg/install/utils/helm"
)

var istioCrd = apiextensions.CustomResourceDefinition{}

var (
	kubeClient kubernetes.Interface
)

var _ = Describe("HelmChartInstaller", func() {
	var (
		ns   string
		inst = NewHelmInstaller()
	)
	BeforeEach(func() {
		kubeClient = superglootest.MustKubeClient()
		// wait for all services in the previous namespace to be torn down
		// important because of a race caused by nodeport conflcit
		if ns != "" {
			superglootest.WaitForIstioTeardown(ns)
		}
		ns = "test" + testutils.RandString(5)
	})
	AfterEach(func() {
		testutils.TeardownKube(ns)
		superglootest.TeardownWithPrefix(kubeClient, "istio")
		superglootest.WaitForIstioTeardown(ns)
	})
	Context("create manifest", func() {
		It("creates resources from a helm chart", func() {
			values := `
mixer:
  enabled: true #should install mixer

`
			manifests, err := RenderManifests(
				context.TODO(),
				"https://s3.amazonaws.com/supergloo.solo.io/istio-1.0.3.tgz",
				values,
				"yella",
				ns,
				"",
				true,
			)
			defer inst.DeleteFromManifests(context.TODO(), ns, manifests)
			Expect(err).NotTo(HaveOccurred())
			err = inst.CreateFromManifests(context.TODO(), ns, manifests)
			Expect(err).NotTo(HaveOccurred())

			// yes mixer
			_, err = kubeClient.AppsV1().Deployments(ns).Get("istio-policy", v1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())
			_, err = kubeClient.AppsV1().Deployments(ns).Get("istio-telemetry", v1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())

			// does not error on crds already existing
			crdManifests, _ := manifests.SplitByCrds()
			err = inst.CreateFromManifests(context.TODO(), ns, crdManifests)
			Expect(err).NotTo(HaveOccurred())

			err = inst.DeleteFromManifests(context.TODO(), ns, manifests)
			Expect(err).NotTo(HaveOccurred())
		})
		It("handles value overrides correctly", func() {
			values := `
mixer:
  enabled: false #should not install mixer

`
			manifests, err := RenderManifests(
				context.TODO(),
				"https://s3.amazonaws.com/supergloo.solo.io/istio-1.0.3.tgz",
				values,
				"yella",
				ns,
				"",
				true,
			)
			Expect(err).NotTo(HaveOccurred())

			for _, man := range manifests {
				// no security crds
				Expect(man.Content).NotTo(ContainSubstring("policies.authentication.istio.io"))

				// no mixer-policy
				Expect(man.Content).NotTo(ContainSubstring(`apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: istio-policy`))
				// no mixer-telemetry
				Expect(man.Content).NotTo(ContainSubstring(`apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: istio-telemetry`))
			}

		})
	})
	Context("update manifest", func() {
		It("updates the existing deployed resources correctly", func() {
			values := `
mixer:
  enabled: true #should install mixer

`
			manifests, err := RenderManifests(
				context.TODO(),
				"https://s3.amazonaws.com/supergloo.solo.io/istio-1.0.3.tgz",
				values,
				"yella",
				ns,
				"",
				true,
			)
			defer inst.DeleteFromManifests(context.TODO(), ns, manifests)
			Expect(err).NotTo(HaveOccurred())
			err = inst.CreateFromManifests(context.TODO(), ns, manifests)
			Expect(err).NotTo(HaveOccurred())

			var foundMixerPolicy, foundMixerTelemetry bool
			for _, man := range manifests {
				// yes mixer-policy
				if strings.Contains(man.Content, `apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: istio-policy`) {
					foundMixerPolicy = true
				}
				if strings.Contains(man.Content, `apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: istio-telemetry`) {
					foundMixerTelemetry = true
				}
			}

			Expect(foundMixerPolicy).To(BeTrue())
			Expect(foundMixerTelemetry).To(BeTrue())

		})
	})
})
