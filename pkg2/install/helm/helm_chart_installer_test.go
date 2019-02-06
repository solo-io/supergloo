package helm_test

import (
	"context"

	"github.com/solo-io/go-utils/kubeutils"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/solo-io/supergloo/pkg2/install/helm"
)

var istioCrd = apiextensions.CustomResourceDefinition{}

var _ = Describe("HelmChartInstaller", func() {
	It("installs from a helm chart", func() {
		ns := "test"
		values := `
security:
  enabled: true #should install policy crds

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
		err = ApplyManifests(context.TODO(), ns, manifests)
		Expect(err).NotTo(HaveOccurred())

		cfg, err := kubeutils.GetConfig("", "")
		Expect(err).NotTo(HaveOccurred())

		apiexts, err := clientset.NewForConfig(cfg)
		Expect(err).NotTo(HaveOccurred())

		_, err = apiexts.ApiextensionsV1beta1().CustomResourceDefinitions().Get("policies.authentication.istio.io", v1.GetOptions{})
		Expect(err).NotTo(HaveOccurred())

		err = DeleteManifests(context.TODO(), ns, manifests)
		Expect(err).NotTo(HaveOccurred())
	})
	It("handles value overrides correctly", func() {
		ns := "test2"
		values := `
security:
  enabled: false #should not install policy crds

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
		err = ApplyManifests(context.TODO(), ns, manifests)
		Expect(err).NotTo(HaveOccurred())

		cfg, err := kubeutils.GetConfig("", "")
		Expect(err).NotTo(HaveOccurred())

		apiexts, err := clientset.NewForConfig(cfg)
		Expect(err).NotTo(HaveOccurred())
		_, err = apiexts.ApiextensionsV1beta1().CustomResourceDefinitions().Get("policies.authentication.istio.io", v1.GetOptions{})
		Expect(err).To(HaveOccurred())

		err = DeleteManifests(context.TODO(), ns, manifests)
		Expect(err).NotTo(HaveOccurred())
	})
})
