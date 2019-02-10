package istio

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/testutils"
	"github.com/solo-io/solo-kit/test/helpers"
	"github.com/solo-io/supergloo/pkg/install/helm"
	"github.com/solo-io/supergloo/test/utils"
	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func cleanupManifest(ns, version string, blocking bool) {
	defer func() {
		testutils.TeardownKube(ns)
		if !blocking {
			return
		}

		kube := utils.MustKubeClient()
		// wait for ns to be removed
		Eventually(func() error {
			_, err := kube.CoreV1().Namespaces().Get(ns, metav1.GetOptions{})
			return err
		}, time.Minute).Should(Not(BeNil()))
	}()
	chart := supportedIstioVersions[version].chartPath
	manifests, err := helm.RenderManifests(
		context.TODO(),
		chart,
		"",
		"test",
		ns,
		"",
		true,
	)
	Expect(err).NotTo(HaveOccurred())
	helm.DeleteFromManifests(context.TODO(), ns, manifests)
}

func assertDeploymentExists(namespace, name string) {
	kube := utils.MustKubeClient()
	Eventually(func() (*v1.Deployment, error) {
		return kube.AppsV1().Deployments(namespace).Get(name, metav1.GetOptions{})
	}).Should(Not(BeNil()))
}

var _ = Describe("InstallIstio", func() {
	type test struct {
		opts InstallOptions
	}
	table.DescribeTable("multiple istio versions",
		func(t test, blocking ...bool) {
			if len(blocking) == 0 {
				blocking = []bool{true}
			}
			ns := "a" + helpers.RandString(5)
			t.opts.Namespace = ns
			defer cleanupManifest(ns, t.opts.Version, blocking[0])
			err := InstallIstio(context.TODO(), t.opts)
			Expect(err).NotTo(HaveOccurred())
			if t.opts.Observability.EnablePrometheus {
				assertDeploymentExists(ns, "prometheus")
			}
			if t.opts.Observability.EnableGrafana {
				assertDeploymentExists(ns, "grafana")
			}
			if t.opts.Observability.EnableJaeger {
				assertDeploymentExists(ns, "istio-tracing")
			}
		},

		table.Entry("istio 1.0.3 enable all", test{
			opts: InstallOptions{
				Version: IstioVersion103,
				Mtls: MtlsInstallOptions{
					Enabled:        true,
					SelfSignedCert: true,
				},
				Observability: ObservabilityInstallOptions{
					EnableGrafana:    true,
					EnableJaeger:     true,
					EnablePrometheus: true,
				},
			},
		}),
		table.Entry("istio 1.0.5 enable all", test{
			opts: InstallOptions{
				Version: IstioVersion105,
				Mtls: MtlsInstallOptions{
					Enabled:        true,
					SelfSignedCert: true,
				},
				Observability: ObservabilityInstallOptions{
					EnableGrafana:    true,
					EnableJaeger:     true,
					EnablePrometheus: true,
				},
			},
		}, false),
	)
})
