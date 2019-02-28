package istio

import (
	"context"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/solo-io/go-utils/testutils"
	"github.com/solo-io/solo-kit/test/helpers"
	kubev1 "k8s.io/api/core/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
)

var _ = Describe("Installer", func() {
	var ns string
	BeforeEach(func() {
		// wait for all services in the previous namespace to be torn down
		// important because of a race caused by nodeport conflcit
		if ns != "" {
			Eventually(Expect(func() []kubev1.Service {
				svcs, err := MustKubeClient().CoreV1().Services(ns).List(metav1.ListOptions{})
				Expect(err).NotTo(HaveOccurred())
				return svcs.Items
			}), time.Second*30).Should(BeEmpty())
		}
		ns = "a" + helpers.RandString(5)
	})
	AfterEach(func() {
		testutils.TeardownKube(ns)
	})

	It("installs, upgrades, and uninstalls from an install object", func() {
		installConfig := &v1.Install_Istio_{
			Istio: &v1.Install_Istio{
				InstallationNamespace: ns,
				IstioVersion:          IstioVersion105,
			},
		}

		install := &v1.Install{
			Metadata:    core.Metadata{Name: "myinstall", Namespace: "myns"},
			Disabled:    false,
			InstallType: installConfig,
		}
		installedMesh, err := EnsureIstioInstall(context.TODO(), install)
		Expect(err).NotTo(HaveOccurred())
		Expect(installedMesh.Metadata.Name).To(Equal(install.Metadata.Name))

		// installed manifest should be set
		Expect(install.InstalledManifest).NotTo(HaveLen(0))

		// should be set by install
		Expect(install.InstalledMesh).NotTo(BeNil())
		Expect(*install.InstalledMesh).To(Equal(installedMesh.Metadata.Ref()))

		assertDeploymentExists(ns, "prometheus", false)

		Expect(installedMesh.Metadata.Name).To(Equal(install.Metadata.Name))

		// enable prometheus
		installConfig.Istio.InstallPrometheus = true

		installedMesh, err = EnsureIstioInstall(context.TODO(), install)
		Expect(err).NotTo(HaveOccurred())

		assertDeploymentExists(ns, "prometheus", true)

		// uninstall should work
		install.Disabled = true
		installedMesh, err = EnsureIstioInstall(context.TODO(), install)
		Expect(err).NotTo(HaveOccurred())
		Expect(installedMesh).To(BeNil())

		assertDeploymentExists(ns, "pilot", false)

	})
})
