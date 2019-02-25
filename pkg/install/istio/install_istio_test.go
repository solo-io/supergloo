package istio

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/kubeutils"
	"github.com/solo-io/go-utils/testutils"
	"github.com/solo-io/solo-kit/test/helpers"
	"github.com/solo-io/supergloo/pkg/install/utils/helm"
	v1 "k8s.io/api/apps/v1"
	kubeerrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// TODO(move to utils package)
func MustKubeClient() kubernetes.Interface {
	restConfig, err := kubeutils.GetConfig("", "")
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	kubeClient, err := kubernetes.NewForConfig(restConfig)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	return kubeClient
}

func cleanupManifest(ns, version string, blocking bool) {
	defer func() {
		testutils.TeardownKube(ns)
		if !blocking {
			return
		}

		kube := MustKubeClient()

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

func assertDeploymentExists(namespace, name string, exists bool) {
	kube := MustKubeClient()
	if exists {
		EventuallyWithOffset(1, func() (*v1.Deployment, error) {
			return kube.AppsV1().Deployments(namespace).Get(name, metav1.GetOptions{})
		}).Should(Not(BeNil()))
	} else {
		EventuallyWithOffset(1, func() bool {
			_, err := kube.AppsV1().Deployments(namespace).Get(name, metav1.GetOptions{})
			return kubeerrs.IsNotFound(err)
		}).Should(BeTrue())
	}
}

var _ = Describe("installIstio", func() {
	type test struct {
		opts installOptions
	}
	table.DescribeTable("multiple istio versions",
		func(t test, blocking ...bool) {
			if len(blocking) == 0 {
				blocking = []bool{true}
			}
			ns := "a" + helpers.RandString(5)
			t.opts.Namespace = ns
			defer cleanupManifest(ns, t.opts.Version, blocking[0])
			manifests, err := installIstio(context.TODO(), t.opts)
			Expect(err).NotTo(HaveOccurred())
			assertDeploymentExists(ns, "prometheus", t.opts.Observability.EnablePrometheus)
			assertDeploymentExists(ns, "grafana", t.opts.Observability.EnableGrafana)
			assertDeploymentExists(ns, "istio-tracing", t.opts.Observability.EnableJaeger)
			Expect(manifestsHaveKey(manifests, "istio/charts/pilot/templates/deployment.yaml")).To(BeTrue())
		},

		table.Entry("istio 1.0.3 enable all", test{
			opts: installOptions{
				Version: IstioVersion103,
				Mtls: mtlsInstallOptions{
					Enabled:        true,
					SelfSignedCert: true,
				},
				Observability: observabilityInstallOptions{
					EnableGrafana:    true,
					EnableJaeger:     true,
					EnablePrometheus: true,
				},
			},
		}),
		table.Entry("istio 1.0.5 enable all", test{
			opts: installOptions{
				Version: IstioVersion105,
				Mtls: mtlsInstallOptions{
					Enabled:        true,
					SelfSignedCert: true,
				},
				Observability: observabilityInstallOptions{
					EnableGrafana:    true,
					EnableJaeger:     true,
					EnablePrometheus: true,
				},
			},
		}, false),
	)
})

func manifestsHaveKey(manifests helm.Manifests, key string) bool {
	for _, man := range manifests {
		if man.Name == key {
			return true
		}
	}
	return false
}
