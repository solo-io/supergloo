package linkerd_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"

	sgutils "github.com/solo-io/supergloo/cli/test/utils"
	"github.com/solo-io/supergloo/install/helm/supergloo/generate"

	"github.com/avast/retry-go"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	gotestutils "github.com/solo-io/go-utils/testutils"
	"github.com/solo-io/go-utils/testutils/clusterlock"
	"github.com/solo-io/supergloo/cli/pkg/helpers/clients"
	"github.com/solo-io/supergloo/pkg/setup"
	"github.com/solo-io/supergloo/test/testutils"
	kubev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func TestE2e(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "E2e Suite")
}

var (
	kube                                kubernetes.Interface
	lock                                *clusterlock.TestClusterLocker
	rootCtx                             context.Context
	cancel                              func()
	basicNamespace, namespaceWithInject string
	promNamespace                       = "prometheus-test" + gotestutils.RandString(4)
)

const (
	linkerdNamesapce = "linkerd"
	glooNamespace    = "gloo-system"
)

var _ = BeforeSuite(func() {
	kube = testutils.MustKubeClient()
	var err error

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

	rootCtx, cancel = context.WithCancel(context.TODO())
	// create sg ns
	_, err = kube.CoreV1().Namespaces().Create(&kubev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: "supergloo-system"},
	})
	Expect(err).NotTo(HaveOccurred())

	// start supergloo
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

	// install discovery via cli
	// start discovery
	var superglooErr error
	projectRoot := filepath.Join(os.Getenv("GOPATH"), "src", os.Getenv("PROJECT_ROOT"))
	err = generate.Run("dev", "Always", projectRoot)
	if err == nil {
		superglooErr = sgutils.Supergloo(fmt.Sprintf("init --release latest --values %s", filepath.Join(projectRoot, generate.ValuesOutput)))
	} else {
		superglooErr = sgutils.Supergloo("init --release latest")
	}
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
	testutils.TeardownWithPrefix(kube, "linkerd")
	testutils.WaitForNamespaceTeardown("supergloo-system")
	testutils.WaitForNamespaceTeardown(basicNamespace)
	testutils.WaitForNamespaceTeardown(namespaceWithInject)
	testutils.WaitForNamespaceTeardown(linkerdNamesapce)
	testutils.WaitForNamespaceTeardown(glooNamespace)
	log.Printf("done!")
}
