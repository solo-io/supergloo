package e2e_test

import (
	"context"
	"log"
	"os"
	"testing"

	gotestutils "github.com/solo-io/go-utils/testutils"
	"github.com/solo-io/supergloo/cli/pkg/helpers/clients"
	"github.com/solo-io/supergloo/test/e2e/utils"

	"github.com/avast/retry-go"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/testutils/clusterlock"
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
			Name:   namespaceWithInject,
			Labels: map[string]string{"istio-injection": "enabled"},
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
})

var _ = AfterSuite(func() {
	teardown()
})

func teardown() {
	if cancel != nil {
		cancel()
	}
	defer lock.ReleaseLock()
	testutils.TeardownSuperGloo(testutils.MustKubeClient())
	kube.CoreV1().Namespaces().Delete("istio-system", nil)
	kube.CoreV1().Namespaces().Delete(basicNamespace, nil)
	kube.CoreV1().Namespaces().Delete(namespaceWithInject, nil)
	testutils.TeardownIstio(kube)
	err := teardownPrometheus(promNamespace)
	if err != nil {
		log.Printf("failed to teardown prometheus: %v", err)
	}
	testutils.WaitForNamespaceTeardown("supergloo-system")
	testutils.WaitForNamespaceTeardown(basicNamespace)
	testutils.WaitForNamespaceTeardown(namespaceWithInject)
	log.Printf("done!")
}

func teardownPrometheus(namespace string) error {
	manifest, err := helmTemplate("--name=prometheus",
		"--namespace="+namespace,
		"--set", "rbac.create=true",
		"--set", "server.persistentVolume.enabled=false",
		"--set", "alertmanager.enabled=false",
		"files/prometheus-8.9.0.tgz")
	if err != nil {
		return err
	}

	err = utils.KubectlDelete(namespace, manifest)
	if err != nil {
		return err
	}

	err = kube.CoreV1().Namespaces().Delete(namespace, nil)
	if err != nil {
		return err
	}

	return nil
}
