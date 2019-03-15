package initialize_test

import (
	"bytes"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/supergloo/cli/test/utils"
	superglootest "github.com/solo-io/supergloo/test/e2e/utils"
	"github.com/solo-io/supergloo/test/testutils"
	v12 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("init", func() {
	AfterEach(func() {
		kube := testutils.MustKubeClient()
		kube.CoreV1().Namespaces().Delete("supergloo-system", nil)
	})
	It("successfully installs supergloo to the cluster", func() {
		err := utils.Supergloo("init --release latest")
		Expect(err).NotTo(HaveOccurred())

		kube := testutils.MustKubeClient()
		Eventually(func() error {
			_, err := kube.CoreV1().Namespaces().Get("supergloo-system", v1.GetOptions{})
			return err
		}).ShouldNot(HaveOccurred())
		Eventually(func() (v12.PodPhase, error) {
			p, err := kube.CoreV1().Pods("supergloo-system").List(v1.ListOptions{})
			if err != nil {
				return v12.PodPending, err
			}
			for _, item := range p.Items {
				if strings.HasPrefix(item.Name, "supergloo") {
					return item.Status.Phase, nil
				}
			}
			return v12.PodPending, errors.Errorf("supergloo pod not found")
		}, time.Minute).Should(Equal(v12.PodRunning))

		// uninstall + test dry run
		manifest, err := utils.SuperglooOut("init --release latest --dry-run")
		Expect(err).NotTo(HaveOccurred())
		err = superglootest.Kubectl(bytes.NewBufferString(manifest), "delete", "-f", "-")
		Expect(err).NotTo(HaveOccurred())
	})
})
