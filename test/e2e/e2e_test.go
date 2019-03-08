package e2e_test

import (
	"strings"
	"time"

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

var _ = Describe("E2e", func() {
	It("installs upgrades and uninstalls istio", func() {
		err := utils.Supergloo("install istio --name=my-istio --mtls=true --auto-inject=true")
		Expect(err).NotTo(HaveOccurred())

		installClient := helpers.MustInstallClient()

		Eventually(func() (core.Status_State, error) {
			i, err := installClient.Read("supergloo-system", "my-istio", clients.ReadOpts{})
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
			_, err := meshClient.Read("supergloo-system", "my-istio", clients.ReadOpts{})
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

		err = waitUntilPodsRunning(time.Minute*2, basicNamespace,
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

		// with mtls in strict mode, curl will fail from non-injected testrunner
		utils3.TestRunnerCurlEventuallyShouldRespond(rootCtx, basicNamespace, setup.CurlOpts{
			Service: "details." + namespaceWithInject + ".svc.cluster.local",
			Port:    9080,
			Path:    "/details/1",
		}, "Recv failure: Connection reset by peer", time.Minute*2)

		// with mtls enabled, curl will succedd from injected testrunner
		utils3.TestRunnerCurlEventuallyShouldRespond(rootCtx, namespaceWithInject, setup.CurlOpts{
			Service: "details." + namespaceWithInject + ".svc.cluster.local",
			Port:    9080,
			Path:    "/details/1",
		}, `"author":"William Shakespeare"`, time.Minute*2)

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
		Eventually(func() error {
			_, err = meshClient.Read("supergloo-system", "my-istio", clients.ReadOpts{})
			return err
		}, time.Minute*2).Should(HaveOccurred())
		Expect(kubeerrs.IsNotFound(err)).To(BeTrue())

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
