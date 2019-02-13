package consul

import (
	"context"
	"fmt"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/go-utils/testutils"
	"github.com/solo-io/solo-kit/test/helpers"
	"github.com/solo-io/supergloo/pkg/install/helm"
	"github.com/solo-io/supergloo/test/utils"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
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
	chart := supportedConsulVersions[version].chartPath
	manifests, err := helm.RenderManifests(
		context.TODO(),
		chart,
		"",
		releaseName(ns, version),
		ns,
		"",
		true,
	)
	Expect(err).NotTo(HaveOccurred())
	helm.DeleteFromManifests(context.TODO(), ns, manifests)
}

func assertPodExists(namespace, withPrefix string, podLabels map[string]string) {
	kube := utils.MustKubeClient()
	Eventually(func() (*v1.Pod, error) {
		pods, err := kube.CoreV1().Pods(namespace).List(metav1.ListOptions{
			LabelSelector: labels.SelectorFromSet(podLabels).String(),
		})
		if err != nil {
			return nil, err
		}
		for _, p := range pods.Items {
			if strings.HasPrefix(p.Name, withPrefix) {
				return &p, nil
			}
		}
		return nil, errors.Errorf("%v not found", withPrefix)
	}).Should(Not(BeNil()))
}

var _ = Describe("InstallConsul", func() {
	type test struct {
		opts InstallOptions
	}
	table.DescribeTable("multiple node counts, enable/disable injector",
		func(t test, blocking ...bool) {
			if len(blocking) == 0 {
				blocking = []bool{true}
			}
			ns := "a" + helpers.RandString(5)
			t.opts.Namespace = ns
			defer cleanupManifest(ns, t.opts.Version, blocking[0])
			err := InstallConsul(context.TODO(), t.opts)
			Expect(err).NotTo(HaveOccurred())
			releaseName := releaseName(ns, t.opts.Version)
			assertPodExists(ns, releaseName+"-consul-", map[string]string{"component": "client"})
			if t.opts.AutoInject.Enabled {
				assertPodExists(ns, releaseName+"-consul-", map[string]string{"component": "connect-injector"})
			}
			for i := 0; i < t.opts.NodeCount; i++ {
				assertPodExists(ns, fmt.Sprintf("%v-consul-server-%v", releaseName, i), map[string]string{"component": "server"})
			}
		},

		table.Entry("1 node", test{
			opts: InstallOptions{
				Version:   ConsulVersion060,
				NodeCount: 1,
			},
		}),
		table.Entry("1 node injector enabled", test{
			opts: InstallOptions{
				Version:   ConsulVersion060,
				NodeCount: 1,
				AutoInject: AutoInjectInstallOptions{
					Enabled: true,
				},
			},
		}),
		table.Entry("2 node", test{
			opts: InstallOptions{
				Version:   ConsulVersion060,
				NodeCount: 2,
			},
		}, false),
	)
})
