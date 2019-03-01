package istio

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/supergloo/pkg/install/utils/helm"
)

var _ = Describe("installIstio", func() {
	type test struct {
		opts installOptions
	}
	t := func() test {
		return test{
			opts: installOptions{
				Namespace: "test",
				Version:   IstioVersion103,
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
			t.opts.Version = ""
			hi := newMockHelm(nil, nil, nil)
			installer := &defaultIstioInstaller{helmInstaller: hi}
			_, err := installer.installIstio(context.TODO(), t.opts)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("must provide istio install version"))
		})
		It("errors on version", func() {
			t := t()
			t.opts.Version = "asdf"
			hi := newMockHelm(nil, nil, nil)
			installer := &defaultIstioInstaller{helmInstaller: hi}
			_, err := installer.installIstio(context.TODO(), t.opts)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("is not a supported istio version"))
		})
		It("errors on namespace", func() {
			t := t()
			t.opts.Namespace = ""
			hi := newMockHelm(nil, nil, nil)
			installer := &defaultIstioInstaller{helmInstaller: hi}
			_, err := installer.installIstio(context.TODO(), t.opts)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("must provide istio install namespace"))
		})
	})
	Context("valid opts", func() {
		It("calls the helm installer with correct charts for the opts", func() {
			t := t()
			var actualManifests helm.Manifests
			hi := newMockHelm(func(ctx context.Context, namespace string, manifests helm.Manifests) error {
				actualManifests = manifests
				return nil
			}, nil, nil)
			installer := &defaultIstioInstaller{helmInstaller: hi}

			t.opts.Observability.EnableJaeger = true
			_, err := installer.installIstio(context.TODO(), t.opts)
			Expect(err).NotTo(HaveOccurred())
			Expect(actualManifests).NotTo(BeEmpty())

			t.opts.Observability.EnableJaeger = false
			_, err = installer.installIstio(context.TODO(), t.opts)
			Expect(err).NotTo(HaveOccurred())
			Expect(actualManifests).NotTo(BeEmpty())
			Expect(manifestsHaveKey(actualManifests, "istio/charts/tracing/templates/service-jaeger.yaml")).To(BeFalse())
		})
		It("calls an update when a previous install is present", func() {
			t := t()
			var actualManifests, updateManifests helm.Manifests
			hi := newMockHelm(func(ctx context.Context, namespace string, manifests helm.Manifests) error {
				actualManifests = manifests
				return nil
			}, nil, func(ctx context.Context, namespace string, original, updated helm.Manifests, recreatePods bool) error {
				updateManifests = updated
				return nil
			})
			installer := &defaultIstioInstaller{helmInstaller: hi}

			t.opts.Observability.EnableJaeger = true
			_, err := installer.installIstio(context.TODO(), t.opts)
			Expect(err).NotTo(HaveOccurred())
			Expect(actualManifests).NotTo(BeEmpty())

			t.opts.previousInstall = actualManifests

			t.opts.Observability.EnableJaeger = false
			_, err = installer.installIstio(context.TODO(), t.opts)
			Expect(err).NotTo(HaveOccurred())
			Expect(updateManifests).NotTo(BeEmpty())
			Expect(manifestsHaveKey(updateManifests, "istio/charts/tracing/templates/service-jaeger.yaml")).To(BeFalse())
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

type mockHelm struct {
	create func(ctx context.Context, namespace string, manifests helm.Manifests) error
	delete func(ctx context.Context, namespace string, manifests helm.Manifests) error
	update func(ctx context.Context, namespace string, original, updated helm.Manifests, recreatePods bool) error
}

func newMockHelm(create func(ctx context.Context, namespace string, manifests helm.Manifests) error,
	delete func(ctx context.Context, namespace string, manifests helm.Manifests) error,
	update func(ctx context.Context, namespace string, original, updated helm.Manifests, recreatePods bool) error) *mockHelm {
	return &mockHelm{
		create: create,
		delete: delete,
		update: update,
	}
}

func (h *mockHelm) CreateFromManifests(ctx context.Context, namespace string, manifests helm.Manifests) error {
	return h.create(ctx, namespace, manifests)
}

func (h *mockHelm) DeleteFromManifests(ctx context.Context, namespace string, manifests helm.Manifests) error {
	return h.delete(ctx, namespace, manifests)
}

func (h *mockHelm) UpdateFromManifests(ctx context.Context, namespace string, original, updated helm.Manifests, recreatePods bool) error {
	return h.update(ctx, namespace, original, updated, recreatePods)
}
