package install_test

import (
	"context"
	"fmt"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/solo-io/supergloo/cli/pkg/helpers"
	"github.com/solo-io/supergloo/cli/pkg/helpers/clients"
	"github.com/solo-io/supergloo/cli/pkg/helpers/mocks"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/install/istio"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	skclients "github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/supergloo/cli/test/utils"
	"github.com/solo-io/supergloo/test/inputs"
)

var _ = Describe("Install", func() {

	var (
		ctrl               *gomock.Controller
		mockKubectl        *mocks.MockKubectl
		superglooNamespace = "supergloo-system"
	)

	BeforeEach(func() {
		clients.UseMemoryClients()

		ctrl = gomock.NewController(T)
		mockKubectl = mocks.NewMockKubectl(ctrl)
		helpers.SetKubectlMock(mockKubectl)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	getInstall := func(name, namespace string) *v1.Install {
		in, err := clients.MustInstallClient().Read(namespace, name, skclients.ReadOpts{})
		ExpectWithOffset(1, err).NotTo(HaveOccurred())
		return in
	}

	Context("non-interactive", func() {

		Describe("should create the expected install", func() {

			var (
				ctx    context.Context
				cancel context.CancelFunc
			)

			BeforeEach(func() {
				ctx, cancel = context.WithCancel(context.Background())
			})

			AfterEach(func() {
				cancel()
			})

			installAndVerifyIstio := func(
				ctx context.Context,
				name,
				namespace,
				version string,
				mtls,
				autoInject,
				ingress,
				egress,
				prometheus,
				jaeger,
				grafana,
				smiInstall bool) {

				expectedKubectlCalls := 0
				if smiInstall {
					expectedKubectlCalls = 1

					// If the SMI adapter needs to be installed, supergloo will block until the install transitions to
					// the 'accepted' status. We will perform this transition manually to simulate a successful install
					go func() {
						defer GinkgoRecover()
						completeInstall(ctx, core.ResourceRef{Namespace: superglooNamespace, Name: name})
					}()
				}
				mockKubectl.EXPECT().ApplyManifest(gomock.Any()).Times(expectedKubectlCalls)

				err := utils.Supergloo("install istio " +
					fmt.Sprintf("--name=%v ", name) +
					fmt.Sprintf("--installation-namespace istio ") +
					fmt.Sprintf("--version=%v ", version) +
					fmt.Sprintf("--mtls=%v ", mtls) +
					fmt.Sprintf("--auto-inject=%v ", autoInject) +
					fmt.Sprintf("--ingress=%v ", ingress) +
					fmt.Sprintf("--egress=%v ", egress) +
					fmt.Sprintf("--grafana=%v ", grafana) +
					fmt.Sprintf("--prometheus=%v ", prometheus) +
					fmt.Sprintf("--jaeger=%v ", jaeger) +
					fmt.Sprintf("--smi-install=%v ", smiInstall))

				if version == "badver" {
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("is not a supported istio version"))
					return
				}

				Expect(err).NotTo(HaveOccurred())

				install := getInstall(name, superglooNamespace)

				istio := MustIstioInstallType(install)
				Expect(istio.Istio.Version).To(Equal(version))
				Expect(istio.Istio.EnableMtls).To(Equal(mtls))
				Expect(istio.Istio.EnableAutoInject).To(Equal(autoInject))
				Expect(istio.Istio.EnableIngress).To(Equal(ingress))
				Expect(istio.Istio.EnableEgress).To(Equal(egress))
				Expect(istio.Istio.InstallPrometheus).To(Equal(prometheus))
				Expect(istio.Istio.InstallJaeger).To(Equal(jaeger))
				Expect(istio.Istio.InstallGrafana).To(Equal(grafana))
			}

			It("work with a valid option set", func() {
				installAndVerifyIstio(ctx, "a1a", "ns", "1.0.3", true, true, true, true, true, true, true, false)
			})

			It("work with another valid option set", func() {
				installAndVerifyIstio(ctx, "b1a", "ns", "1.0.5", false, false, false, false, false, false, false, true)
			})

			It("fails with an invalid option set", func() {
				installAndVerifyIstio(ctx, "c1a", "ns", "badver", false, false, false, false, false, false, false, false)
			})
		})
		It("should create installation namespace if it does not exist to begin with", func() {
			name := "input"
			namespace := "ns"
			installNs := "istio-system"

			err := utils.Supergloo("install istio " +
				fmt.Sprintf("--name=%v ", name) +
				fmt.Sprintf("--installation-namespace %s ", installNs) +
				fmt.Sprintf("--namespace=%v ", namespace))
			Expect(err).NotTo(HaveOccurred())

			kube := clients.MustKubeClient()
			ns, err := kube.CoreV1().Namespaces().Get(installNs, metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())
			Expect(ns.Name).To(Equal(installNs))
		})
		It("should enable an existing + disabled install", func() {
			name := "input"
			namespace := "ns"
			inst := inputs.IstioInstall(name, namespace, "any", "1.0.5", true)
			ic := clients.MustInstallClient()
			_, err := ic.Write(inst, skclients.WriteOpts{})
			Expect(err).NotTo(HaveOccurred())

			err = utils.Supergloo("install istio " +
				fmt.Sprintf("--name=%v ", name) +
				fmt.Sprintf("--namespace=%v ", namespace))
			Expect(err).NotTo(HaveOccurred())

			updatedInstall, err := ic.Read(namespace, name, skclients.ReadOpts{})
			Expect(err).NotTo(HaveOccurred())
			Expect(updatedInstall.Disabled).To(BeFalse())

		})
		It("should error enable on existing enabled install", func() {
			name := "input"
			namespace := "ns"
			inst := inputs.IstioInstall(name, namespace, "any", "1.0.5", false)
			ic := clients.MustInstallClient()
			_, err := ic.Write(inst, skclients.WriteOpts{})
			Expect(err).NotTo(HaveOccurred())

			err = utils.Supergloo("install istio " +
				fmt.Sprintf("--name=%v ", name) +
				fmt.Sprintf("--namespace=%v ", namespace))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("already installed and enabled"))
		})
		It("should update existing enabled install if --update set to true", func() {
			name := "input"
			namespace := "ns"
			inst := inputs.IstioInstall(name, namespace, "istio-system", "1.0.5", false)
			Expect(inst.InstallType).To(BeAssignableToTypeOf(&v1.Install_Mesh{}))
			ic := clients.MustInstallClient()
			_, err := ic.Write(inst, skclients.WriteOpts{})
			Expect(err).NotTo(HaveOccurred())

			err = utils.Supergloo("install istio " +
				fmt.Sprintf("--name=%v ", name) +
				fmt.Sprintf("--namespace=%v ", namespace) +
				"--mtls=true " +
				"--update=true ")
			Expect(err).NotTo(HaveOccurred())

			updatedInstall, err := ic.Read(namespace, name, skclients.ReadOpts{})
			Expect(err).NotTo(HaveOccurred())
			Expect(*updatedInstall).To(Equal(v1.Install{
				Metadata: core.Metadata{
					Name:            name,
					Namespace:       namespace,
					ResourceVersion: updatedInstall.Metadata.ResourceVersion,
				},
				InstallationNamespace: "istio-system",
				Disabled:              false,
				InstallType: &v1.Install_Mesh{
					Mesh: &v1.MeshInstall{
						MeshInstallType: &v1.MeshInstall_Istio{
							Istio: &v1.IstioInstall{
								Version:          istio.IstioVersion106,
								EnableAutoInject: true,
								EnableMtls:       true,
							},
						},
					},
				},
			}))
		})
	})
})

func MustIstioInstallType(install *v1.Install) *v1.MeshInstall_Istio {
	Expect(install.InstallType).To(BeAssignableToTypeOf(&v1.Install_Mesh{}))
	mesh := install.InstallType.(*v1.Install_Mesh)
	Expect(mesh.Mesh.MeshInstallType).To(BeAssignableToTypeOf(&v1.MeshInstall_Istio{}))
	istioMesh := mesh.Mesh.MeshInstallType.(*v1.MeshInstall_Istio)
	return istioMesh
}

func completeInstall(ctx context.Context, installRef core.ResourceRef) {
	err := helpers.WaitForInstallStatus(ctx, installRef, core.Status_Pending, 5*time.Second)
	Expect(err).NotTo(HaveOccurred())

LOOP:
	for {
		select {
		case <-time.After(500 * time.Millisecond):
			break LOOP
		case <-ctx.Done():
			return
		}
	}

	utils.CompleteInstall(ctx, clients.MustInstallClient(), installRef, 0)
}
