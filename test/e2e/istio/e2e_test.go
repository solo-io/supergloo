package istio_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/solo-io/supergloo/cli/pkg/helpers/clients"
	"github.com/solo-io/supergloo/pkg/install/istio"
	sgtestutils "github.com/solo-io/supergloo/test/testutils"

	"github.com/solo-io/go-utils/testutils"
	"k8s.io/apimachinery/pkg/labels"

	"github.com/solo-io/supergloo/test/inputs"

	skerrors "github.com/solo-io/solo-kit/pkg/errors"

	kubeerrs "k8s.io/apimachinery/pkg/api/errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/errors"
	skclients "github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/solo-kit/test/setup"
	"github.com/solo-io/supergloo/cli/test/utils"
	sgutils "github.com/solo-io/supergloo/test/e2e/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const superglooNamespace = "supergloo-system"

var _ = Describe("istio e2e", func() {
	istioName := "my-istio"
	glooName := "gloo"

	It("it installs istio", func() {
		testInstallIstio(istioName)
	})

	It("discovers and configures smi adapter", func() {
		err := sgutils.KubectlApplyFile(smiIstioAdapterFile)
		Expect(err).NotTo(HaveOccurred())

		// give pod a short time to start with Pending
		time.Sleep(time.Millisecond * 100)

		err = sgtestutils.WaitUntilPodsRunning(time.Minute*4, "istio-system",
			"smi-adapter-istio",
		)
		Expect(err).NotTo(HaveOccurred())

		testTrafficShifting()
		err = sgutils.KubectlDeleteFile(smiIstioAdapterFile)
		Expect(err).NotTo(HaveOccurred())
	})

	It("it installs gloo", func() {
		sgutils.TestGlooInstall(glooName, superglooNamespace, glooNamespace, istioName)
	})

	It("it enforces policy", func() {
		testPolicy(istioName)
	})

	It("it configures prometheus", func() {
		testConfigurePrometheus(istioName, promNamespace)
	})

	It("it enables mtls", func() {
		testMtls()
	})

	It("it enables mtls ingress with gloo", func() {
		testGlooMtls(istioName)
	})

	It("it sets custom root ca", func() {
		testCertRotation(istioName)
	})

	It("it shifts traffic with routing rules", func() {
		testTrafficShifting()
	})

	It("it injects faults with routing rules", func() {
		testFaultInjection()
	})

	It("it uninstalls gloo", func() {
		testUninstallGloo(glooName)
	})

	It("it uninstalls istio", func() {
		testUninstallIstio(istioName)
	})
})

/*
   tests
*/
func testInstallIstio(meshName string) {
	version := istio.IstioVersion106
	if istioVersion := os.Getenv("ISTIO_VERSION"); istioVersion != "" {
		version = istioVersion
	}
	err := utils.Supergloo(fmt.Sprintf("install istio --name=%v --version=%v --mtls=true --auto-inject=true",
		meshName, version))
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
		_, err := kube.CoreV1().Services(istioNamespace).Get("istio-pilot", metav1.GetOptions{})
		return err
	}).ShouldNot(HaveOccurred())

	meshClient := clients.MustMeshClient()
	Eventually(func() error {
		_, err := meshClient.Read(superglooNamespace, meshName, skclients.ReadOpts{})
		return err
	}, time.Second*30, time.Second).ShouldNot(HaveOccurred())

	err = sgtestutils.WaitUntilPodsRunning(time.Minute*2, istioNamespace,
		"istio-citadel",
		"istio-galley",
		"istio-pilot",
		"istio-policy",
		"istio-sidecar-injector",
		"istio-telemetry",
	)
	Expect(err).NotTo(HaveOccurred())

	time.Sleep(time.Second * 15) // give the sidecar injector extra time to wake up

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

func testGlooInstall(glooName, istioName string) {
	err := utils.Supergloo(fmt.Sprintf("install gloo --name=%s --target-meshes %s.%s ",
		glooName, superglooNamespace, istioName))
	Expect(err).NotTo(HaveOccurred())

	installClient := clients.MustInstallClient()

	Eventually(func() (core.Status_State, error) {
		i, err := installClient.Read(superglooNamespace, glooName, skclients.ReadOpts{})
		if err != nil {
			return 0, err
		}
		Expect(i.Status.Reason).To(Equal(""))
		return i.Status.State, nil
	}, time.Minute*4).Should(Equal(core.Status_Accepted))

	meshIngressClient := clients.MustMeshIngressClient()
	Eventually(func() error {
		_, err := meshIngressClient.Read(superglooNamespace, glooName, skclients.ReadOpts{})
		return err
	}, time.Minute*2).ShouldNot(HaveOccurred())

	err = sgtestutils.WaitUntilPodsRunning(time.Minute*2, glooNamespace,
		"gloo",
		"gateway",
	)
	Expect(err).NotTo(HaveOccurred())
}

func testCertRotation(meshName string) {
	// create tls cert here to use as custom root cert
	certsDir, err := ioutil.TempDir("", "supergloocerts")
	Expect(err).NotTo(HaveOccurred())
	defer os.RemoveAll(certsDir)
	err = writeCerts(certsDir)
	Expect(err).NotTo(HaveOccurred())
	secretName := "rootcert"
	err = createTlsSecret(secretName, certsDir)
	Expect(err).NotTo(HaveOccurred())

	// update our mesh with the root cert
	err = setRootCert(meshName, secretName)
	Expect(err).NotTo(HaveOccurred())

	var certChain string
	Eventually(func() (string, error) {
		rootCa, cc, err := getCerts("details", namespaceWithInject)
		if err != nil {
			return "", err
		}
		certChain = cc
		return rootCa, nil
	}, time.Minute*4).Should(Equal(inputs.RootCert))

	Expect(certChain).To(HaveSuffix(inputs.CertChain))

	// check that communication still works
	// curl should succeed from injected testrunner
	sgutils.TestRunnerCurlEventuallyShouldRespond(rootCtx, namespaceWithInject, setup.CurlOpts{
		Service: "details." + namespaceWithInject + ".svc.cluster.local",
		Port:    9080,
		Path:    "/details/1",
	}, `"author":"William Shakespeare"`, time.Minute*8)
}

func testMtls() {
	// with mtls in strict mode, curl will fail from non-injected testrunner
	sgutils.TestRunnerCurlEventuallyShouldRespond(rootCtx, basicNamespace, setup.CurlOpts{
		Service: "details." + namespaceWithInject + ".svc.cluster.local",
		Port:    9080,
		Path:    "/details/1",
	}, "Recv failure: Connection reset by peer", time.Minute*3)

	// with mtls enabled, curl will succeed from injected testrunner
	sgutils.TestRunnerCurlEventuallyShouldRespond(rootCtx, namespaceWithInject, setup.CurlOpts{
		Service: "details." + namespaceWithInject + ".svc.cluster.local",
		Port:    9080,
		Path:    "/details/1",
	}, `"author":"William Shakespeare"`, time.Minute*3)
}

func testGlooMtls(istioName string) {
	service := "details"
	port := 9080
	upstreamName := fmt.Sprintf("%s-%s-%d", namespaceWithInject, service, port)
	err := utils.Supergloo(fmt.Sprintf("set upstream mtls --name %s --target-mesh %s.%s",
		upstreamName, superglooNamespace, istioName))
	Expect(err).NotTo(HaveOccurred())

	sgutils.TestGlooIngress(rootCtx, namespaceWithInject, superglooNamespace, glooNamespace, basicNamespace)
}

func testPolicy(meshName string) {
	// apply an 'identiy' security rule, disabling communication between all injected services
	err := utils.Supergloo(
		fmt.Sprintf("apply securityrule --target-mesh %v.%v --name enable-rbac ", superglooNamespace, meshName) +
			fmt.Sprintf("--source-upstreams %v.%v-testrunner-8080 ", superglooNamespace, namespaceWithInject) +
			fmt.Sprintf("--dest-upstreams %v.%v-testrunner-8080 ", superglooNamespace, namespaceWithInject),
	)

	Expect(err).NotTo(HaveOccurred())

	defer func() {
		// delete security rule
		// TODO (ilackarms): replace this with cli command when we have one for deleting routing rules
		err = clients.MustSecurityRuleClient().Delete(superglooNamespace, "enable-rbac", skclients.DeleteOpts{})
		Expect(err).NotTo(HaveOccurred())
	}()

	// test that communication is forbidden
	sgutils.TestRunnerCurlEventuallyShouldRespond(rootCtx, namespaceWithInject, setup.CurlOpts{
		Service: "reviews." + namespaceWithInject + ".svc.cluster.local",
		Port:    9080,
	}, `RBAC: access denied`, time.Minute*5)

	// update security rule to enable traffic from testrunner to reviews
	err = utils.Supergloo(
		fmt.Sprintf("apply securityrule --target-mesh %v.%v --name enable-rbac ", superglooNamespace, meshName) +
			fmt.Sprintf("--source-upstreams %v.%v-testrunner-8080 ", superglooNamespace, namespaceWithInject) +
			fmt.Sprintf("--dest-upstreams %v.%v-reviews-9080 ", superglooNamespace, namespaceWithInject),
	)
	Expect(err).NotTo(HaveOccurred())

	// test that communication is enabled
	sgutils.TestRunnerCurlEventuallyShouldRespond(rootCtx, namespaceWithInject, setup.CurlOpts{
		Service: "reviews." + namespaceWithInject + ".svc.cluster.local",
		Port:    9080,
		Path:    "/reviews/1",
	}, `"reviewer": "Reviewer1",`, time.Minute*5)
}

func testTrafficShifting() {
	// apply a traffic shifting rule, divert traffic to reviews
	err := utils.Supergloo(fmt.Sprintf("apply routingrule trafficshifting --target-mesh %v.my-istio --name hi "+
		"--destination %v.%v-reviews-9080:%v", superglooNamespace, superglooNamespace, namespaceWithInject, 1))

	Expect(err).NotTo(HaveOccurred())

	sgutils.TestRunnerCurlEventuallyShouldRespond(rootCtx, namespaceWithInject, setup.CurlOpts{
		Service: "details." + namespaceWithInject + ".svc.cluster.local",
		Port:    9080,
		Path:    "/reviews/1",
	}, `"reviewer": "Reviewer1",`, time.Minute*5)

	// delete traffic shifting rule
	// TODO (ilackarms): replace this with cli command when we have one for deleting routing rules
	err = clients.MustRoutingRuleClient().Delete(superglooNamespace, "hi", skclients.DeleteOpts{})
	Expect(err).NotTo(HaveOccurred())

	// ensure normal behavior restored
	sgutils.TestRunnerCurlEventuallyShouldRespond(rootCtx, namespaceWithInject, setup.CurlOpts{
		Service: "details." + namespaceWithInject + ".svc.cluster.local",
		Port:    9080,
		Path:    "/details/1",
	}, `{"id":1,"author":"William Shakespeare","year":1595,"type":"paperback","pages":200,"publisher":"PublisherA","language":"English","ISBN-10":"1234567890","ISBN-13":"123-1234567890"}`, time.Minute*5)
}

func testFaultInjection() {
	httpError, percent := 404, 50

	// apply a traffic shifting rule, divert traffic to reviews
	err := utils.Supergloo(fmt.Sprintf("apply rr fi a http --name one --target-mesh %v.my-istio "+
		"-p %d -s %d ", superglooNamespace, percent, httpError))

	Expect(err).NotTo(HaveOccurred())

	sgutils.TestRunnerCurlEventuallyShouldRespond(rootCtx, namespaceWithInject, setup.CurlOpts{
		Service: "reviews." + namespaceWithInject + ".svc.cluster.local",
		Port:    9080,
		Path:    "/reviews/1",
	}, "404", time.Minute*5)

	// delete routing rule
	// TODO (ilackarms): replace this with cli command when we have one for deleting routing rules
	err = clients.MustRoutingRuleClient().Delete(superglooNamespace, "one", skclients.DeleteOpts{})
	Expect(err).NotTo(HaveOccurred())

}

func testUninstallIstio(meshName string) {
	// test uninstall works
	err := utils.Supergloo("uninstall --name=" + meshName)
	Expect(err).NotTo(HaveOccurred())

	err = nil
	Eventually(func() error {
		_, err = kube.CoreV1().Services(istioNamespace).Get("istio-pilot", metav1.GetOptions{})
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

func testConfigurePrometheus(meshName, promNamespace string) {
	err := sgutils.DeployPrometheus(kube, promNamespace)
	Expect(err).NotTo(HaveOccurred())

	err = utils.Supergloo(fmt.Sprintf("set mesh stats "+
		"--target-mesh supergloo-system.%v "+
		"--prometheus-configmap %v.prometheus-server", meshName, promNamespace))
	Expect(err).NotTo(HaveOccurred())

	// assert the sample is valid
	sgutils.TestRunnerCurlEventuallyShouldRespond(rootCtx, basicNamespace, setup.CurlOpts{
		Service: "prometheus-server." + promNamespace + ".svc.cluster.local",
		Port:    80,
		Path:    `/api/v1/query?query=istio_requests_total\{\}`,
	}, `"istio_requests_total"`, time.Minute*5)
}

/*
util funcs
*/
func testUninstallGloo(meshIngressName string) {
	// test uninstall works
	err := utils.Supergloo("uninstall --name=" + meshIngressName)
	Expect(err).NotTo(HaveOccurred())

	err = nil
	Eventually(func() bool {
		_, err = clients.MustMeshClient().Read(superglooNamespace, meshIngressName, skclients.ReadOpts{})
		if err == nil {
			return false
		}
		return skerrors.IsNotExist(err)
	}, time.Minute*2).Should(BeTrue())
}

func writeCerts(dir string) error {
	secretContent := inputs.InputTlsSecret("", "")
	err := ioutil.WriteFile(filepath.Join(dir, "CaCert"), []byte(secretContent.CaCert), 0644)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filepath.Join(dir, "CaKey"), []byte(secretContent.CaKey), 0644)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filepath.Join(dir, "RootCert"), []byte(secretContent.RootCert), 0644)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filepath.Join(dir, "CertChain"), []byte(secretContent.CertChain), 0644)
	if err != nil {
		return err
	}
	return nil
}

func createTlsSecret(name, certDir string) error {
	err := utils.Supergloo(
		fmt.Sprintf("create secret tls --name %v --cacert %v --cakey %v --rootcert %v --certchain %v ", name,
			filepath.Join(certDir, "CaCert"),
			filepath.Join(certDir, "CaKey"),
			filepath.Join(certDir, "RootCert"),
			filepath.Join(certDir, "CertChain"),
		))
	if err != nil {
		return err
	}
	return nil
}

func setRootCert(targetMesh, tlsSecret string) error {
	return utils.Supergloo(
		fmt.Sprintf("set mesh rootcert --target-mesh supergloo-system.%v --tls-secret supergloo-system.%v", targetMesh, tlsSecret))
}

func getCerts(appLabel, namespace string) (string, string, error) {
	pods, err := clients.MustKubeClient().CoreV1().Pods(namespace).List(metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(map[string]string{"app": appLabel}).String(),
	})
	if err != nil {
		return "", "", err
	}
	if len(pods.Items) == 0 {
		return "", "", errors.Errorf("no pods found with label app: %v", appLabel)
	}

	// based on https://istio.io/docs/tasks/security/plugin-ca-cert/#verifying-the-new-certificates
	rootCert, err := testutils.KubectlOut("exec", "-n", namespace, pods.Items[0].Name, "-c", "istio-proxy", "/bin/cat", "/etc/certs/root-cert.pem")
	if err != nil {
		return "", "", err
	}
	certChain, err := testutils.KubectlOut("exec", "-n", namespace, pods.Items[0].Name, "-c", "istio-proxy", "/bin/cat", "/etc/certs/cert-chain.pem")
	if err != nil {
		return "", "", err
	}
	return rootCert, certChain, nil
}
