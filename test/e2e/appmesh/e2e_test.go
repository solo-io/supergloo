package appmesh_test

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	skclients "github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/supergloo/cli/pkg/helpers/clients"
	"github.com/solo-io/supergloo/install/helm/supergloo/generate"
	sgutils "github.com/solo-io/supergloo/test/e2e/utils"
	sgtestutils "github.com/solo-io/supergloo/test/testutils"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/supergloo/cli/test/utils"
)

const superglooNamespace = "supergloo-system"

var _ = Describe("E2e", func() {
	It("registers and tests appmesh", func() {
		// install discovery via cli
		// start discovery
		var superglooErr error
		projectRoot := filepath.Join(os.Getenv("GOPATH"), "src", os.Getenv("PROJECT_ROOT"))
		err := generate.Run("dev", "Always", projectRoot)
		if err == nil {
			superglooErr = utils.Supergloo(fmt.Sprintf("init --release latest --values %s", filepath.Join(projectRoot, generate.ValuesOutput)))
		} else {
			superglooErr = utils.Supergloo("init --release latest")
		}
		Expect(superglooErr).NotTo(HaveOccurred())

		// TODO (ilackarms): add a flag to switch between starting supergloo locally and deploying via cli
		sgtestutils.DeleteSuperglooPods(kube, superglooNamespace)
		appmeshName := "appmesh"
		secretName := "my-secret"

		createAWSSecret(secretName)

		testRegisterAppmesh(appmeshName, secretName)

		testUnregisterAppmesh(appmeshName)
	})
})

/*
   tests
*/
func testRegisterAppmesh(meshName, secretName string) {
	region, vnLabel := "us-east-1", "vn"
	err := utils.Supergloo(fmt.Sprintf("register appmesh --name %s --region %s "+
		"--secret %s --select-namespaces %s --virtual-node-label %s", meshName, region, secretName, namespaceWithInject, vnLabel))
	Expect(err).NotTo(HaveOccurred())

	meshClient := clients.MustMeshClient()
	Eventually(func() error {
		_, err := meshClient.Read(superglooNamespace, meshName, skclients.ReadOpts{})
		return err
	}).ShouldNot(HaveOccurred())

	err = sgutils.DeployTestRunner(basicNamespace)
	Expect(err).NotTo(HaveOccurred())

	// the sidecar injector might take some time to become available
	Eventually(func() error {
		return sgutils.DeployTestRunner(namespaceWithInject)
	}, time.Minute*1).ShouldNot(HaveOccurred())

	err = sgutils.DeployBookInfo(namespaceWithInject)
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

func testUnregisterAppmesh(meshName string) {

}

func createAWSSecret(secretName string) {
	accessKeyId, secretAccessKey := os.Getenv("AWS_ACCESS_KEY_ID"), os.Getenv("AWS_SECRET_ACCESS_KEY")
	Expect(accessKeyId).NotTo(Equal(""))
	Expect(secretAccessKey).NotTo(Equal(""))
	err := utils.Supergloo(fmt.Sprintf(
		"create secret aws --name %s --access-key-id %s --secret-access-key %s",
		secretName, accessKeyId, secretAccessKey,
	))
	Expect(err).NotTo(HaveOccurred())

	secret, err := kube.CoreV1().Secrets(superglooNamespace).Get(secretName, v1.GetOptions{})
	Expect(err).NotTo(HaveOccurred())
	Expect(secret).NotTo(BeNil())
}
