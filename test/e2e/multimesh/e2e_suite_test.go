package multimesh_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/onsi/ginkgo/config"

	"github.com/solo-io/supergloo/pkg/version"

	sgutils "github.com/solo-io/supergloo/cli/test/utils"
	"github.com/solo-io/supergloo/test/e2e/utils"

	"github.com/avast/retry-go"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/testutils/clusterlock"
	"github.com/solo-io/supergloo/test/testutils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func TestE2e(t *testing.T) {
	RegisterFailHandler(func(message string, callerSkip ...int) {
		close(failed)
		Fail(message, callerSkip...)
	})
	RunSpecs(t, "multimesh e2e Suite")
}

var (
	kube                kubernetes.Interface
	lock                *clusterlock.TestClusterLocker
	rootCtx, cancel     = context.WithCancel(context.Background())
	smiIstioAdapterFile = utils.MustTestFile("istio-smi-adapter.yaml")
	failed              = make(chan struct{})
)

const (
	istioNamespace             = "istio-system"
	linkerdNamespace           = "linkerd"
	glooNamespace              = "gloo-system"
	superglooNamespace         = "supergloo-system"
	basicNamespace             = "basic-namespace"
	namespaceWithIstioInject   = "namespace-with-istio-inject"
	namespaceWithLinkerdInject = "namespace-with-linkerd-inject"
	namespaceWithAppmeshInject = "namespace-with-appmesh-inject"
	promNamespace              = "prometheus-test"
)

var _ = BeforeSuite(func() {
	opts := struct{ config.GinkgoConfigType }{config.GinkgoConfig}
	log.Printf("%#v", opts)
	kube = testutils.MustKubeClient()
	var err error

	// Get build information
	buildVersion, helmChartUrl, imageRepoPrefix, err := utils.GetBuildInformation()
	Expect(err).NotTo(HaveOccurred())

	// Set the supergloo version (will be equal to the BUILD_ID env)
	version.Version = buildVersion

	lock, err = clusterlock.NewTestClusterLocker(kube, clusterlock.Options{
		IdPrefix: os.ExpandEnv("superglooe2e-{$BUILD_ID}-"),
	})
	Expect(err).NotTo(HaveOccurred())
	Expect(lock.AcquireLock(retry.OnRetry(func(n uint, err error) {
		log.Printf("waiting to acquire lock with err: %v", err)
	}))).NotTo(HaveOccurred())

	// If present, delete all namespaces used in this test
	teardown()

	err = testutils.CreateNamespaces(kube,
		metav1.ObjectMeta{Name: superglooNamespace},
		metav1.ObjectMeta{Name: basicNamespace},
		metav1.ObjectMeta{Name: namespaceWithIstioInject, Labels: map[string]string{"istio-injection": "enabled"}},
		metav1.ObjectMeta{Name: namespaceWithLinkerdInject, Annotations: map[string]string{"linkerd.io/inject": "enabled"}},
		metav1.ObjectMeta{Name: namespaceWithAppmeshInject, Labels: map[string]string{"app-mesh-injection": "enabled"}},
	)
	Expect(err).NotTo(HaveOccurred())

	// Install supergloo using the helm chart specific to this test run
	superglooErr := sgutils.Supergloo(fmt.Sprintf("init -f %s", helmChartUrl))
	Expect(superglooErr).NotTo(HaveOccurred())

	err = utils.DeployPrometheus(kube, promNamespace)
	Expect(err).NotTo(HaveOccurred())

	// If env is set, run supergloo locally and delete remote pods
	if os.Getenv("E2E_RUN_PODS_LOCALLY") != "" {
		log.Println("Running supergloo locally")
		err = testutils.RunSuperglooLocally(rootCtx, kube, superglooNamespace, buildVersion, imageRepoPrefix)
		Expect(err).NotTo(HaveOccurred())

	}

})

var _ = AfterSuite(func() {
	defer lock.ReleaseLock()
	cancel()
	teardown()
})

func teardown() {
	utils.KubectlDeleteFile(smiIstioAdapterFile)
	testutils.TeardownSuperGloo(testutils.MustKubeClient())
	kube.CoreV1().Namespaces().Delete(linkerdNamespace, nil)
	kube.CoreV1().Namespaces().Delete(istioNamespace, nil)
	kube.CoreV1().Namespaces().Delete(glooNamespace, nil)
	kube.CoreV1().Namespaces().Delete(basicNamespace, nil)
	kube.CoreV1().Namespaces().Delete(namespaceWithIstioInject, nil)
	kube.CoreV1().Namespaces().Delete(namespaceWithLinkerdInject, nil)
	kube.CoreV1().Namespaces().Delete(namespaceWithAppmeshInject, nil)
	err := utils.TeardownPrometheus(kube, promNamespace)
	if err != nil {
		log.Printf("failed to teardown prometheus: %v", err)
	}
	testutils.TeardownWithPrefix(kube, "linkerd")
	testutils.TeardownWithPrefix(kube, "istio")
	testutils.TeardownWithPrefix(kube, "gloo")
	testutils.TeardownWithPrefix(kube, "gateway")
	testutils.WaitForNamespaceTeardown("supergloo-system")
	testutils.WaitForNamespaceTeardown(basicNamespace)
	testutils.WaitForNamespaceTeardown(namespaceWithIstioInject)
	testutils.WaitForNamespaceTeardown(namespaceWithLinkerdInject)
	testutils.WaitForNamespaceTeardown(namespaceWithAppmeshInject)
	testutils.WaitForNamespaceTeardown(istioNamespace)
	testutils.WaitForNamespaceTeardown(glooNamespace)
	testutils.WaitForNamespaceTeardown(promNamespace)
	log.Printf("done!")
}
