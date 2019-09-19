package utils

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/gomega"
	glootestutils "github.com/solo-io/gloo/projects/gloo/cli/pkg/testutils"
	skclients "github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/solo-kit/test/setup"
	"github.com/solo-io/supergloo/cli/pkg/helpers/clients"
	"github.com/solo-io/supergloo/cli/test/utils"
	"github.com/solo-io/supergloo/test/testutils"
)

func TestGlooInstall(glooName, superglooNamespace, glooNamespace, meshName string) {
	err := utils.Supergloo(fmt.Sprintf("install gloo --name=%s --target-meshes %s.%s --installation-namespace %v ",
		glooName, superglooNamespace, meshName, glooNamespace))
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

	err = testutils.WaitUntilPodsRunning(time.Minute*2, glooNamespace,
		"gloo",
		"gateway",
	)
	Expect(err).NotTo(HaveOccurred())
}

func TestGlooIngress(rootCtx context.Context, injectedNamespace, superglooNamespace, glooNamespace, basicNamespace string) {
	service := "details"
	port := 9080
	upstreamName := fmt.Sprintf("%s-%s-%d", injectedNamespace, service, port)
	err := glootestutils.Glooctl(fmt.Sprintf("add route --name detailspage"+
		" --namespace %s --path-prefix / --dest-name %s "+
		"--dest-namespace %s", glooNamespace, upstreamName, superglooNamespace))
	Expect(err).NotTo(HaveOccurred())

	// with mtls in strict mode, curl will succeed routing through gloo
	TestRunnerCurlEventuallyShouldRespond(rootCtx, basicNamespace, setup.CurlOpts{
		Service: "gateway-proxy-v2." + glooNamespace + ".svc.cluster.local",
		Port:    80,
		Path:    "/details/1",
	}, `"author":"William Shakespeare"`, time.Minute*3)
}
