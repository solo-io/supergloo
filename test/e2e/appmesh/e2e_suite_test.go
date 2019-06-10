package appmesh_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	"k8s.io/client-go/kubernetes"

	"github.com/avast/retry-go"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/testutils/clusterlock"
	sgutils "github.com/solo-io/supergloo/cli/test/utils"
	"github.com/solo-io/supergloo/pkg/version"
	"github.com/solo-io/supergloo/test/e2e/utils"
	"github.com/solo-io/supergloo/test/testutils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestE2e(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "AWS App Mesh e2e Suite")
}

const (
	superglooNamespace  = "supergloo-system"
	basicNamespace      = "basic-namespace"
	namespaceWithInject = "namespace-with-inject"
)

var (
	kube            kubernetes.Interface
	lock            *clusterlock.TestClusterLocker
	rootCtx, cancel = context.WithCancel(context.Background())
)

var _ = BeforeSuite(func() {
	kube = testutils.MustKubeClient()

	// Get build information
	buildVersion, helmChartUrl, imageRepoPrefix, err := utils.GetBuildInformation()
	Expect(err).NotTo(HaveOccurred())

	// Set the supergloo version (will be equal to the BUILD_ID env)
	version.Version = buildVersion

	// Acquire cluster lock
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
		metav1.ObjectMeta{Name: namespaceWithInject, Labels: map[string]string{"app-mesh-injection": "enabled"}},
	)
	Expect(err).NotTo(HaveOccurred())

	// Install supergloo using the helm chart specific to this test run
	err = sgutils.Supergloo(fmt.Sprintf("init -f %s", helmChartUrl))
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
	testutils.TeardownSuperGloo(kube)
	kube.CoreV1().Namespaces().Delete(superglooNamespace, nil)
	kube.CoreV1().Namespaces().Delete(basicNamespace, nil)
	kube.CoreV1().Namespaces().Delete(namespaceWithInject, nil)

	testutils.WaitForNamespaceTeardown(superglooNamespace)
	testutils.WaitForNamespaceTeardown(basicNamespace)
	testutils.WaitForNamespaceTeardown(namespaceWithInject)
	log.Printf("done!")
}
