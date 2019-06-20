package testfuncs

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"k8s.io/client-go/kubernetes"

	"github.com/solo-io/supergloo/cli/pkg/helpers/clients"
	sgtestutils "github.com/solo-io/supergloo/test/testutils"

	skerrors "github.com/solo-io/solo-kit/pkg/errors"

	kubeerrs "k8s.io/apimachinery/pkg/api/errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	skclients "github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/solo-kit/test/setup"
	"github.com/solo-io/supergloo/cli/test/utils"
	sgutils "github.com/solo-io/supergloo/test/e2e/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type LinkerdE2eTestParams struct {
	Kube                kubernetes.Interface
	PromNamespace       string
	GlooNamespace       string
	SuperglooNamespace  string
	MeshName            string
	BasicNamespace      string
	NamespaceWithInject string
	LinkerdNamespace    string
	RootCtx             context.Context
	SharedLock          *sync.Mutex
}

func RunLinkerdE2eTests(params LinkerdE2eTestParams) {
	var (
		glooName = "gloo"
	)

	kube := params.Kube
	By("it installs linkerd", func() {
		testInstallLinkerd(kube, params.MeshName, params.BasicNamespace, params.NamespaceWithInject, params.SuperglooNamespace, params.LinkerdNamespace)
	})

	By("it installs gloo", func() {
		testConfigureOrInstallGloo(glooName, params.SuperglooNamespace, params.GlooNamespace, params.MeshName)
	})

	By("it configures prometheus", func() {
		testConfigurePrometheusWithLinkerd(params.RootCtx, params.SharedLock, params.MeshName, params.BasicNamespace, params.PromNamespace)
	})

	By("retries failed requests", func() {
		testLinkerdRetries(params.MeshName, params.NamespaceWithInject)
	})

	By("tests ingress with gloo", func() {
		sgutils.TestGlooIngress(params.RootCtx, params.NamespaceWithInject, params.SuperglooNamespace, params.GlooNamespace, params.BasicNamespace)
	})

	By("it uninstalls linkerd", func() {
		testUninstallLinkerd(kube, params.MeshName, params.SuperglooNamespace, params.LinkerdNamespace)
	})
}

/*
   tests
*/
func testInstallLinkerd(kube kubernetes.Interface, meshName, basicNamespace, namespaceWithInject, superglooNamespace, linkerdNamespace string) {
	err := utils.Supergloo(fmt.Sprintf("install linkerd --name=%v --mtls=true --auto-inject=true", meshName))
	Expect(err).NotTo(HaveOccurred())

	installClient := clients.MustInstallClient()

	Eventually(func() (core.Status_State, error) {
		i, err := installClient.Read(superglooNamespace, meshName, skclients.ReadOpts{})
		if err != nil {
			return 0, err
		}
		Expect(i.Status.Reason).To(Equal(""))
		return i.Status.State, nil
	}, time.Minute*5).Should(Equal(core.Status_Accepted))

	Eventually(func() error {
		_, err := kube.CoreV1().Services(linkerdNamespace).Get("linkerd-controller-api", metav1.GetOptions{})
		return err
	}).ShouldNot(HaveOccurred())

	meshClient := clients.MustMeshClient()
	Eventually(func() error {
		_, err := meshClient.Read(superglooNamespace, meshName, skclients.ReadOpts{})
		return err
	}).ShouldNot(HaveOccurred())

	err = sgtestutils.WaitUntilPodsRunning(time.Minute*2, linkerdNamespace,
		"linkerd-controller",
		"linkerd-web",
		"linkerd-prometheus",
		"linkerd-grafana",
		"linkerd-identity",
		"linkerd-proxy-injector",
		"linkerd-sp-validator",
	)
	Expect(err).NotTo(HaveOccurred())

	err = sgutils.DeployTestRunner(basicNamespace)
	Expect(err).NotTo(HaveOccurred())

	// the sidecar injector might take some time to become available
	Eventually(func() error {
		return sgutils.DeployTestRunner(namespaceWithInject)
	}, time.Minute*1).ShouldNot(HaveOccurred())

	err = sgutils.DeployBookInfoIstio(namespaceWithInject)
	Expect(err).NotTo(HaveOccurred())

	err = sgtestutils.WaitUntilPodsRunning(time.Minute*4, basicNamespace,
		"testrunner",
	)
	Expect(err).NotTo(HaveOccurred())

	err = sgtestutils.WaitUntilPodsRunning(time.Minute*2, namespaceWithInject,
		"testrunner",
		"reviews-v1",
		"reviews-v2",
		"reviews-v3",
	)
	Expect(err).NotTo(HaveOccurred())
}

func testUninstallLinkerd(kube kubernetes.Interface, meshName, superglooNamespace, linkerdNamespace string) {
	// test uninstall works
	err := utils.Supergloo("uninstall --name=" + meshName)
	Expect(err).NotTo(HaveOccurred())

	err = nil
	Eventually(func() error {
		_, err = kube.CoreV1().Services(linkerdNamespace).Get("linkerd-controller-api", metav1.GetOptions{})
		return err
	}, time.Minute*2).Should(HaveOccurred())
	Expect(kubeerrs.IsNotFound(err)).To(BeTrue())

	err = nil
	Eventually(func() bool {
		_, err = clients.MustMeshClient().Read(superglooNamespace, meshName, skclients.ReadOpts{})
		if err == nil {
			return false
		}
		return skerrors.IsNotExist(err)
	}, time.Minute*2).Should(BeTrue())
}

func testLinkerdRetries(meshName, namespaceWithInject string) {
	// deploy 50% failing test service
	err := sgutils.DeployTestService(namespaceWithInject)
	Expect(err).NotTo(HaveOccurred())

	// wait for pod ready
	err = sgtestutils.WaitUntilPodsRunning(time.Minute*2, namespaceWithInject,
		"test-service",
	)
	Expect(err).NotTo(HaveOccurred())

	err = utils.Supergloo(fmt.Sprintf("apply routingrule retries budget "+
		"--name my-retry-policy "+
		"--ratio 0.5 "+
		"--min-retries 3 "+
		"--ttl 1m "+
		"--target-mesh supergloo-system.%v", meshName))
	Expect(err).NotTo(HaveOccurred())

	// initially, the service will fail every other request
	// then when the retry policy becomes active,
	// 3 successive requests should all succeed
	curlThreeTimes := func() error {
		for i := 0; i < 3; i++ {
			resp, err := sgutils.TestRunnerCurl(namespaceWithInject, setup.CurlOpts{
				Service: "test-service." + namespaceWithInject + ".svc.cluster.local",
				Port:    8080,
				Path:    "/retry-this-route",
			})
			if err != nil {
				return err
			}
			if !strings.Contains(resp, "200 OK") {
				return fmt.Errorf("resp was not 200 OK:n\n%v", resp)
			}
		}
		return nil
	}

	Eventually(curlThreeTimes, time.Minute*3).ShouldNot(HaveOccurred())
}

func testConfigurePrometheusWithLinkerd(rootCtx context.Context, lock *sync.Mutex, meshName, basicNamespace, promNamespace string) {
	lock.Lock()
	defer lock.Unlock()
	err := utils.Supergloo(fmt.Sprintf("set mesh stats "+
		"--target-mesh supergloo-system.%v "+
		"--prometheus-configmap %v.prometheus-server", meshName, promNamespace))
	Expect(err).NotTo(HaveOccurred())

	// assert the sample is valid
	sgutils.TestRunnerCurlEventuallyShouldRespond(rootCtx, basicNamespace, setup.CurlOpts{
		Service: "prometheus-server." + promNamespace + ".svc.cluster.local",
		Port:    80,
		Path:    `/api/v1/query?query=request_total\{\}`,
	}, `"job":"linkerd-proxy"`, time.Minute*5)
}
