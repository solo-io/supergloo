package istio

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/supergloo/pkg/install/utils/helm"
)

var _ = Describe("installOrUpdateIstio", func() {
	type test struct {
		opts installOptions
	}
	t := func() test {
		return test{
			opts: installOptions{
				namespace: "test",
				version:   IstioVersion103,
				Mtls: mtlsInstallOptions{
					Enabled:        true,
					SelfSignedCert: true,
				},
				Observability: observabilityInstallOptions{
					EnableGrafana:    false,
					EnableJaeger:     true,
					EnablePrometheus: false,
				},
			},
		}
	}
	Context("invalid opts", func() {
		It("errors on version", func() {
			t := t()
			t.opts.version = ""
			hi := helm.NewMockHelm(nil, nil, nil)
			t.opts.installer = hi
			_, err := helm.InstallOrUpdate(context.TODO(), t.opts)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("must provide istio install version"))
		})
		It("errors on version", func() {
			t := t()
			t.opts.version = "asdf"
			hi := helm.NewMockHelm(nil, nil, nil)
			t.opts.installer = hi
			_, err := helm.InstallOrUpdate(context.TODO(), t.opts)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("is not a supported istio version"))
		})
		It("errors on namespace", func() {
			t := t()
			t.opts.namespace = ""
			hi := helm.NewMockHelm(nil, nil, nil)
			t.opts.installer = hi
			_, err := helm.InstallOrUpdate(context.TODO(), t.opts)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("must provide istio install namespace"))
		})
	})
	Context("valid opts", func() {
		It("calls the helm installer with correct charts for the opts", func() {
			t := t()
			var actualManifests helm.Manifests
			hi := helm.NewMockHelm(func(ctx context.Context, namespace string, manifests helm.Manifests) error {
				actualManifests = manifests
				return nil
			}, nil, nil)
			t.opts.installer = hi

			t.opts.Observability.EnableJaeger = true
			_, err := helm.InstallOrUpdate(context.TODO(), t.opts)
			Expect(err).NotTo(HaveOccurred())
			Expect(actualManifests).NotTo(BeEmpty())

			t.opts.Observability.EnableJaeger = false
			_, err = helm.InstallOrUpdate(context.TODO(), t.opts)
			Expect(err).NotTo(HaveOccurred())
			Expect(actualManifests).NotTo(BeEmpty())
			Expect(manifestsHaveKey(actualManifests, "istio/charts/tracing/templates/service-jaeger.yaml")).To(BeFalse())
		})
		It("calls an update when a previous install is present", func() {
			t := t()
			var createManifests, updateManifests helm.Manifests
			hi := helm.NewMockHelm(func(ctx context.Context, namespace string, manifests helm.Manifests) error {
				createManifests = manifests
				return nil
			}, nil, func(ctx context.Context, namespace string, original, updated helm.Manifests, recreatePods bool) error {
				updateManifests = updated
				return nil
			})
			t.opts.installer = hi

			t.opts.Observability.EnableJaeger = true
			_, err := helm.InstallOrUpdate(context.TODO(), t.opts)
			Expect(err).NotTo(HaveOccurred())
			Expect(createManifests).NotTo(BeEmpty())
			Expect(manifestsHaveKey(createManifests, "istio/charts/tracing/templates/service-jaeger.yaml")).To(BeTrue())

			t.opts.previousInstall = createManifests

			t.opts.Observability.EnableJaeger = false
			_, err = helm.InstallOrUpdate(context.TODO(), t.opts)
			Expect(err).NotTo(HaveOccurred())
			Expect(updateManifests).NotTo(BeEmpty())
			Expect(manifestsHaveKey(updateManifests, "istio/charts/tracing/templates/service-jaeger.yaml")).To(BeFalse())
		})
		It("does not run update if the opts do not change", func() {
			t := t()
			hi := helm.NewMockHelm(func(ctx context.Context, namespace string, manifests helm.Manifests) error {
				return nil
			}, nil, func(ctx context.Context, namespace string, original, updated helm.Manifests, recreatePods bool) error {
				return fmt.Errorf("i was not expected to be called")
			})
			t.opts.installer = hi

			// perform initial install
			manifests, err := helm.InstallOrUpdate(context.TODO(), t.opts)
			Expect(err).NotTo(HaveOccurred())

			// same opts, now with a previous install
			t.opts.previousInstall = manifests

			// should be a no-op
			// mock helm will return an error here if update is called
			_, err = helm.InstallOrUpdate(context.TODO(), t.opts)
			Expect(err).NotTo(HaveOccurred())

			// change a config opt
			// now we should call update, and expect the error
			t.opts.Mtls.Enabled = !t.opts.Mtls.Enabled
			_, err = helm.InstallOrUpdate(context.TODO(), t.opts)
			Expect(err).To(HaveOccurred())
		})
	})
})

func manifestsHaveKey(manifests helm.Manifests, key string) bool {
	for _, man := range manifests {
		if man.Name == key {
			return true
		}
	}
	return false
}
