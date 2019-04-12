package linkerd

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/installutils/kubeinstall/mocks"
)

var _ = Describe("values", func() {
	ns := "space"
	It("renders helm values correctly", func() {
		testValues := func(enableMtls, enableAutoInject bool) {
			o := newInstallOpts(
				Version_stable221,
				ns,
				enableMtls,
				enableAutoInject,
			)
			values, err := o.values()
			Expect(err).NotTo(HaveOccurred())
			Expect(values).To(ContainSubstring(fmt.Sprintf("EnableTLS: %v", enableMtls)))
			Expect(values).To(ContainSubstring(fmt.Sprintf("ProxyAutoInjectEnabled: %v", enableAutoInject)))
		}

		testValues(false, false)
		testValues(false, true)
		testValues(true, false)
		testValues(true, true)
		testValues(false, true)
	})
})

var _ = Describe("install", func() {
	It("passes the expected resources to the installer", func() {
		kubeInstaller := &mocks.MockKubeInstaller{}
		o := newInstallOpts(
			Version_stable221,
			"ns",
			true,
			true,
		)
		err := o.install(context.TODO(), kubeInstaller, nil)
		Expect(err).NotTo(HaveOccurred())
		calledWith := kubeInstaller.ReconcileCalledWith
		Expect(calledWith.InstallNamespace).To(Equal("ns"))
		Expect(calledWith.Resources).To(HaveLen(31))

	})
})
