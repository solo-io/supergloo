package testfuncs

import (
	"fmt"
	"os"
	"time"

	"k8s.io/client-go/kubernetes"

	"github.com/solo-io/supergloo/cli/pkg/helpers/clients"
	sgtestutils "github.com/solo-io/supergloo/test/testutils"

	. "github.com/onsi/gomega"
	skclients "github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/supergloo/cli/test/utils"
	sgutils "github.com/solo-io/supergloo/test/e2e/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type AppMeshE2eTestParams struct {
	Kube                kubernetes.Interface
	MeshName            string
	SuperglooNamespace  string
	BasicNamespace      string
	NamespaceWithInject string
}

func RunAppMeshE2eTests(params AppMeshE2eTestParams) {

	secretName := "my-secret"

	createAWSSecret(params.Kube, secretName, params.SuperglooNamespace)

	testRegisterAppmesh(params.Kube, params.MeshName, secretName, params.BasicNamespace, params.SuperglooNamespace, params.NamespaceWithInject)
}

/*
   tests
*/
func testRegisterAppmesh(kube kubernetes.Interface, meshName, secretName, basicNamespace, superglooNamespace, namespaceWithInject string) {
	region, vnLabel := "us-east-1", "app"
	err := utils.Supergloo(fmt.Sprintf("register appmesh --name %s --namespace %s --region %s "+
		"--secret %s.%s --select-namespaces %s --virtual-node-label %s",
		meshName, basicNamespace, region, superglooNamespace, secretName, namespaceWithInject, vnLabel))
	Expect(err).NotTo(HaveOccurred())

	meshClient := clients.MustMeshClient()
	Eventually(func() error {
		_, err := meshClient.Read(basicNamespace, meshName, skclients.ReadOpts{})
		return err
	}).ShouldNot(HaveOccurred())

	err = sgtestutils.WaitUntilPodsRunning(time.Minute*4, superglooNamespace, "sidecar-injector")
	Expect(err).NotTo(HaveOccurred())

	err = sgutils.DeployTestRunner(basicNamespace)
	Expect(err).NotTo(HaveOccurred())

	err = sgutils.DeployBookInfoAppmesh(namespaceWithInject)
	Expect(err).NotTo(HaveOccurred())

	err = sgtestutils.WaitUntilPodsRunning(time.Minute*4, basicNamespace,
		"testrunner",
	)
	Expect(err).NotTo(HaveOccurred())

	err = sgtestutils.WaitUntilPodsRunning(time.Minute*2, namespaceWithInject,
		"reviews-v1",
		"reviews-v2",
		"reviews-v3",
	)
	Expect(err).NotTo(HaveOccurred())

	checkSidecarInjection(kube, secretName)

}

func checkSidecarInjection(kube kubernetes.Interface, namespaceWithInject string) {
	pods, err := kube.CoreV1().Pods(namespaceWithInject).List(metav1.ListOptions{})
	Expect(err).NotTo(HaveOccurred())

	for _, pod := range pods.Items {
		Expect(len(pod.Spec.Containers)).To(BeNumerically(">=", 2))
	}
}

func createAWSSecret(kube kubernetes.Interface, secretName, superglooNamespace string) {
	accessKeyId, secretAccessKey := os.Getenv("AWS_ACCESS_KEY_ID"), os.Getenv("AWS_SECRET_ACCESS_KEY")
	Expect(accessKeyId).NotTo(Equal(""))
	Expect(secretAccessKey).NotTo(Equal(""))
	err := utils.Supergloo(fmt.Sprintf(
		"create secret aws --name %s --access-key-id %s --secret-access-key %s",
		secretName, accessKeyId, secretAccessKey,
	))
	Expect(err).NotTo(HaveOccurred())

	secret, err := kube.CoreV1().Secrets(superglooNamespace).Get(secretName, metav1.GetOptions{})
	Expect(err).NotTo(HaveOccurred())
	Expect(secret).NotTo(BeNil())
}
