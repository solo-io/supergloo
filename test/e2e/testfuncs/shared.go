package testfuncs

import (
	"context"
	"fmt"
	"time"

	v1 "github.com/solo-io/supergloo/pkg/api/v1"

	. "github.com/onsi/gomega"
	skclients "github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/supergloo/cli/pkg/helpers/clients"
	"github.com/solo-io/supergloo/cli/test/utils"
	"github.com/solo-io/supergloo/test/testutils"
)

func testConfigureOrInstallGloo(glooName, superglooNamespace, glooNamespace, meshName string) {
	// check if the install already exists; if so, we're just going to add our mesh
	installs := clients.MustInstallClient()
	existingGlooInstall, err := installs.Read(superglooNamespace, glooName, skclients.ReadOpts{})
	if err == nil {
		Expect(existingGlooInstall.InstallType).To(BeAssignableToTypeOf(&v1.Install_Ingress{}))
		ingressInstall := existingGlooInstall.InstallType.(*v1.Install_Ingress)
		glooInstall := ingressInstall.Ingress.IngressInstallType.(*v1.MeshIngressInstall_Gloo).Gloo
		glooInstall.Meshes = append(glooInstall.Meshes, &core.ResourceRef{Namespace: superglooNamespace, Name: meshName})
		_, err := installs.Write(existingGlooInstall, skclients.WriteOpts{OverwriteExisting: true, Ctx: context.TODO()})
		Expect(err).NotTo(HaveOccurred())
		return
	}

	err = utils.Supergloo(fmt.Sprintf("install gloo --name=%s --target-meshes %s.%s --version v0.14.1",
		glooName, superglooNamespace, meshName))
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
