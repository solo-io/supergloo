package shared_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/utils/kubeutils"
	"github.com/solo-io/solo-kit/test/helpers"
	"github.com/solo-io/solo-kit/test/setup"
	. "github.com/solo-io/supergloo/pkg/install/shared"
	"github.com/solo-io/supergloo/test/utils"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var _ = Describe("InstallKubeManifest", func() {
	var namespace string
	BeforeEach(func() {
		namespace = "install-kube-manifest-" + helpers.RandString(8)
		err := setup.SetupKubeForTest(namespace)
		Expect(err).NotTo(HaveOccurred())
	})
	AfterEach(func() {
		setup.TeardownKube(namespace)
	})
	It("installs arbitrary kube manifests", func() {
		err := deployBookinfo(namespace)
		Expect(err).NotTo(HaveOccurred())

		cfg, err := kubeutils.GetConfig("", "")
		Expect(err).NotTo(HaveOccurred())
		kube, err := kubernetes.NewForConfig(cfg)
		Expect(err).NotTo(HaveOccurred())

		svcs, err := kube.CoreV1().Services(namespace).List(v1.ListOptions{})
		Expect(err).NotTo(HaveOccurred())
		deployments, err := kube.ExtensionsV1beta1().Deployments(namespace).List(v1.ListOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(svcs.Items).To(HaveLen(4))
		Expect(deployments.Items).To(HaveLen(6))

	})
})

func deployBookinfo(namespace string) error {
	cfg, err := kubeutils.GetConfig("", "")
	if err != nil {
		return err
	}
	kube, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return err
	}

	apiext, err := clientset.NewForConfig(cfg)
	if err != nil {
		return err
	}

	installer := NewKubeInstaller(kube, apiext, namespace)

	kubeObjs, err := ParseKubeManifest(utils.IstioBookinfoYaml)
	if err != nil {
		return err
	}

	for _, kubeOjb := range kubeObjs {
		if err := installer.Create(kubeOjb); err != nil {
			return err
		}
	}
	return nil
}
