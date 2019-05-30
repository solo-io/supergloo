package linkerd_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/solo-io/supergloo/pkg/version"

	"github.com/solo-io/supergloo/pkg/constants"

	"github.com/solo-io/supergloo/test/e2e/utils"

	"github.com/avast/retry-go"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/testutils/clusterlock"
	"github.com/solo-io/supergloo/cli/pkg/helpers/clients"
	sgutils "github.com/solo-io/supergloo/cli/test/utils"
	mdsetup "github.com/solo-io/supergloo/pkg/meshdiscovery/setup"
	"github.com/solo-io/supergloo/pkg/setup"
	"github.com/solo-io/supergloo/test/testutils"
	kubev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func TestE2e(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Linkerd e2e Suite")
}

var (
	kube                                kubernetes.Interface
	lock                                *clusterlock.TestClusterLocker
	rootCtx                             context.Context
	cancel                              func()
	basicNamespace, namespaceWithInject string
	promNamespace                       = "prometheus-test"
	chartUrl                            string
)

const (
	linkerdNamesapce = "linkerd"
	glooNamespace    = "gloo-system"
)

var _ = BeforeSuite(func() {
	kube = testutils.MustKubeClient()
	var err error

	// Get build information
	buildVersion, helmChartUrl, imageRepoPrefix, err := utils.GetBuildInformation()
	Expect(err).NotTo(HaveOccurred())

	// Set the supergloo version (will be equal to the BUILD_ID env)
	version.Version = buildVersion
	chartUrl = helmChartUrl

	lock, err = clusterlock.NewTestClusterLocker(kube, clusterlock.Options{
		IdPrefix: os.ExpandEnv("superglooe2e-{$BUILD_ID}-"),
	})
	Expect(err).NotTo(HaveOccurred())
	Expect(lock.AcquireLock(retry.OnRetry(func(n uint, err error) {
		log.Printf("waiting to acquire lock with err: %v", err)
	}))).NotTo(HaveOccurred())

	basicNamespace, namespaceWithInject = "basic-namespace", "namespace-with-inject"

	teardown()

	kube = clients.MustKubeClient()
	_, err = kube.CoreV1().Namespaces().Create(&kubev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: basicNamespace,
		},
	})
	Expect(err).NotTo(HaveOccurred())

	_, err = kube.CoreV1().Namespaces().Create(&kubev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:        namespaceWithInject,
			Annotations: map[string]string{"linkerd.io/inject": "enabled"},
		},
	})
	Expect(err).NotTo(HaveOccurred())

	_, err = kube.CoreV1().Namespaces().Create(&kubev1.Namespace{
		// create sg ns
		ObjectMeta: metav1.ObjectMeta{Name: "supergloo-system"},
	})
	Expect(err).NotTo(HaveOccurred())

	Expect(os.Setenv(constants.PodNamespaceEnvName, superglooNamespace)).NotTo(HaveOccurred())
	image := fmt.Sprintf("%s/%s:%s", imageRepoPrefix, constants.SidecarInjectorImageName, buildVersion)
	Expect(os.Setenv(constants.SidecarInjectorImageNameEnvName, image)).NotTo(HaveOccurred())
	Expect(os.Setenv(constants.SidecarInjectorImagePullPolicyEnvName, "Always")).NotTo(HaveOccurred())
	// start supergloo (requires setting the two envs if running locally)
	rootCtx, cancel = context.WithCancel(context.TODO())
	go func() {
		defer GinkgoRecover()
		err := setup.Main(rootCtx, func(e error) {
			defer GinkgoRecover()
			return
			// TODO: assert errors here
			Expect(e).NotTo(HaveOccurred())
		})
		Expect(err).NotTo(HaveOccurred())
	}()

	// start mesh discovery
	go func() {
		defer GinkgoRecover()
		err := mdsetup.Main(rootCtx, func(e error) {
			defer GinkgoRecover()
			return
			// TODO: assert errors here
			Expect(e).NotTo(HaveOccurred())
		}, nil)
		Expect(err).NotTo(HaveOccurred())
	}()

	// Install supergloo using the helm chart specific to this test run
	superglooErr := sgutils.Supergloo(fmt.Sprintf("init -f %s", chartUrl))
	Expect(superglooErr).NotTo(HaveOccurred())

	// TODO (ilackarms): add a flag to switch between starting supergloo locally and deploying via cli
	testutils.DeleteSuperglooPods(kube, superglooNamespace)
})

var _ = AfterSuite(func() {
	defer lock.ReleaseLock()
	teardown()
})

func teardown() {
	if cancel != nil {
		cancel()
	}
	testutils.TeardownSuperGloo(testutils.MustKubeClient())
	kube.CoreV1().Namespaces().Delete(linkerdNamesapce, nil)
	kube.CoreV1().Namespaces().Delete(glooNamespace, nil)
	kube.CoreV1().Namespaces().Delete(basicNamespace, nil)
	kube.CoreV1().Namespaces().Delete(namespaceWithInject, nil)
	err := utils.TeardownPrometheus(kube, promNamespace)
	if err != nil {
		log.Printf("failed to teardown prometheus: %v", err)
	}
	testutils.TeardownWithPrefix(kube, "linkerd")
	testutils.TeardownWithPrefix(kube, "gloo")
	testutils.WaitForNamespaceTeardown("supergloo-system")
	testutils.WaitForNamespaceTeardown(basicNamespace)
	testutils.WaitForNamespaceTeardown(namespaceWithInject)
	testutils.WaitForNamespaceTeardown(linkerdNamesapce)
	testutils.WaitForNamespaceTeardown(glooNamespace)
	log.Printf("done!")
}
