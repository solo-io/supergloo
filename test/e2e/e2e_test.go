package e2e_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/solo-io/go-utils/testutils"
	"k8s.io/apimachinery/pkg/labels"

	"github.com/solo-io/supergloo/test/inputs"

	"github.com/gogo/protobuf/proto"

	skerrors "github.com/solo-io/solo-kit/pkg/errors"

	kubeerrs "k8s.io/apimachinery/pkg/api/errors"

	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/solo-kit/test/setup"
	utils3 "github.com/solo-io/supergloo/test/e2e/utils"
	v1 "k8s.io/api/core/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/supergloo/cli/pkg/helpers"
	"github.com/solo-io/supergloo/cli/test/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// remove supergloo pods
func deleteSuperglooPods() {
	// wait until pod is gone
	Eventually(func() error {
		dep, err := kube.ExtensionsV1beta1().Deployments("supergloo-system").Get("supergloo", metav1.GetOptions{})
		if err != nil {
			return err
		}
		dep.Spec.Replicas = proto.Int(0)
		_, err = kube.ExtensionsV1beta1().Deployments("supergloo-system").Update(dep)
		if err != nil {
			return err
		}
		pods, err := kube.CoreV1().Pods("supergloo-system").List(metav1.ListOptions{})
		if err != nil {
			return err
		}
		for _, p := range pods.Items {
			if strings.HasPrefix(p.Name, "supergloo") {
				return errors.Errorf("supergloo pods still exist")
			}
		}
		return nil
	}, time.Second*60).ShouldNot(HaveOccurred())

}

var _ = Describe("E2e", func() {
	It("installs upgrades and uninstalls istio", func() {
		// install discovery via cli
		// start discovery
		err := utils.Supergloo("init --release latest")
		Expect(err).NotTo(HaveOccurred())

		// TODO (ilackarms): add a flag to switch between starting supergloo locally and deploying via cli
		deleteSuperglooPods()

		meshName := "my-istio"

		err = utils.Supergloo(fmt.Sprintf("install istio --name=%v --mtls=true --auto-inject=true", meshName))
		Expect(err).NotTo(HaveOccurred())

		installClient := helpers.MustInstallClient()

		Eventually(func() (core.Status_State, error) {
			i, err := installClient.Read("supergloo-system", meshName, clients.ReadOpts{})
			if err != nil {
				return 0, err
			}
			Expect(i.Status.Reason).To(Equal(""))
			return i.Status.State, nil
		}, time.Minute*2).Should(Equal(core.Status_Accepted))

		Eventually(func() error {
			_, err := kube.CoreV1().Services("istio-system").Get("istio-pilot", metav1.GetOptions{})
			return err
		}).ShouldNot(HaveOccurred())

		meshClient := helpers.MustMeshClient()
		Eventually(func() error {
			_, err := meshClient.Read("supergloo-system", meshName, clients.ReadOpts{})
			return err
		}).ShouldNot(HaveOccurred())

		err = waitUntilPodsRunning(time.Minute*2, "istio-system",
			"grafana",
			"istio-citadel",
			"istio-galley",
			"istio-pilot",
			"istio-policy",
			"istio-sidecar-injector",
			"istio-telemetry",
			"istio-tracing",
			"prometheus",
		)
		Expect(err).NotTo(HaveOccurred())

		err = utils3.DeployTestRunner(basicNamespace)
		Expect(err).NotTo(HaveOccurred())

		// the sidecar injector might take some time to become available
		Eventually(func() error {
			return utils3.DeployTestRunner(namespaceWithInject)
		}, time.Minute*1).ShouldNot(HaveOccurred())

		err = utils3.DeployBookInfo(namespaceWithInject)
		Expect(err).NotTo(HaveOccurred())

		err = waitUntilPodsRunning(time.Minute*4, basicNamespace,
			"testrunner",
		)
		Expect(err).NotTo(HaveOccurred())

		err = waitUntilPodsRunning(time.Minute*2, namespaceWithInject,
			"testrunner",
			"reviews-v1",
			"reviews-v2",
			"reviews-v3",
		)
		Expect(err).NotTo(HaveOccurred())

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
		}, time.Minute*3).Should(Equal(inputs.RootCert))

		Expect(certChain).To(HaveSuffix(inputs.CertChain))

		// with mtls in strict mode, curl will fail from non-injected testrunner
		utils3.TestRunnerCurlEventuallyShouldRespond(rootCtx, basicNamespace, setup.CurlOpts{
			Service: "details." + namespaceWithInject + ".svc.cluster.local",
			Port:    9080,
			Path:    "/details/1",
		}, "Recv failure: Connection reset by peer", time.Minute*3)

		// with mtls enabled, curl will succeed from injected testrunner
		utils3.TestRunnerCurlEventuallyShouldRespond(rootCtx, namespaceWithInject, setup.CurlOpts{
			Service: "details." + namespaceWithInject + ".svc.cluster.local",
			Port:    9080,
			Path:    "/details/1",
		}, `"author":"William Shakespeare"`, time.Minute*3)

		//apply a traffic shifting rule, divert traffic to reviews
		err = utils.Supergloo(fmt.Sprintf("apply routingrule trafficshifting --target-mesh supergloo-system.my-istio --name hi --destination %v.%v-reviews-9080:%v", "supergloo-system", namespaceWithInject, 1))
		Expect(err).NotTo(HaveOccurred())

		utils3.TestRunnerCurlEventuallyShouldRespond(rootCtx, namespaceWithInject, setup.CurlOpts{
			Service: "details." + namespaceWithInject + ".svc.cluster.local",
			Port:    9080,
			Path:    "/reviews/1",
		}, `"reviewer": "Reviewer1",`, time.Minute*5)

		// test uninstall works
		err = utils.Supergloo("uninstall --name=my-istio")
		Expect(err).NotTo(HaveOccurred())

		err = nil
		Eventually(func() error {
			_, err = kube.CoreV1().Services("istio-system").Get("istio-pilot", metav1.GetOptions{})
			return err
		}, time.Minute*2).Should(HaveOccurred())
		Expect(kubeerrs.IsNotFound(err)).To(BeTrue())

		err = nil
		Eventually(func() bool {
			_, err = meshClient.Read("supergloo-system", meshName, clients.ReadOpts{})
			if err == nil {
				return false
			}
			return skerrors.IsNotExist(err)
		}, time.Minute*2).Should(BeTrue())
	})
})

func waitUntilPodsRunning(timeout time.Duration, namespace string, podPrefixes ...string) error {
	pods := helpers.MustKubeClient().CoreV1().Pods(namespace)
	getPodStatus := func(prefix string) (*v1.PodPhase, error) {
		list, err := pods.List(metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
		for _, pod := range list.Items {
			if strings.HasPrefix(pod.Name, prefix) {
				return &pod.Status.Phase, nil
			}
		}
		return nil, errors.Errorf("pod with prefix %v not found", prefix)
	}
	failed := time.After(timeout)
	notYetRunning := make(map[string]v1.PodPhase)
	for {
		select {
		case <-failed:
			return errors.Errorf("timed out waiting for pods to come online: %v", notYetRunning)
		case <-time.After(time.Second / 2):
			notYetRunning = make(map[string]v1.PodPhase)
			for _, prefix := range podPrefixes {
				stat, err := getPodStatus(prefix)
				if err != nil {
					return err
				}
				if *stat != v1.PodRunning {
					notYetRunning[prefix] = *stat
				}
			}
			if len(notYetRunning) == 0 {
				return nil
			}
		}

	}
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
		fmt.Sprintf("set rootcert --target-mesh supergloo-system.%v --tls-secret supergloo-system.%v", targetMesh, tlsSecret))
}

func getCerts(appLabel, namespace string) (string, string, error) {
	pods, err := helpers.MustKubeClient().CoreV1().Pods(namespace).List(metav1.ListOptions{
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
