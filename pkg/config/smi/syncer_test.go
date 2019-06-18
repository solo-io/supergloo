package smi_test

import (
	"context"

	. "github.com/solo-io/supergloo/pkg/config/linkerd"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/reporter"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/supergloo/test/inputs"

	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/translator/linkerd"
)

var _ = Describe("Syncer", func() {
	It("translates with the translator, reconciles with the reconciler, and reports on the reporter", func() {
		mt := &mockTranslator{}
		mRec := &mockReconcilers{}
		mRep := &mockReporter{}
		syncer := NewLinkerdConfigSyncer(mt, mRec, mRep)

		err := syncer.Sync(context.TODO(), &v1.ConfigSnapshot{})
		Expect(err).NotTo(HaveOccurred())
		Expect(mt.called).To(BeTrue())
		Expect(mRec.called).To(BeTrue())
		Expect(mRep.called).To(BeTrue())
	})
})

type mockTranslator struct{ called bool }

func (mt *mockTranslator) Translate(ctx context.Context, snapshot *v1.ConfigSnapshot) (map[*v1.Mesh]*linkerd.MeshConfig, reporter.ResourceErrors, error) {
	mt.called = true
	return map[*v1.Mesh]*linkerd.MeshConfig{inputs.LinkerdMesh("anynamespace", nil): {}}, reporter.ResourceErrors{}, nil
}

type mockReconcilers struct{ called bool }

func (mr *mockReconcilers) ReconcileAll(ctx context.Context, config *linkerd.MeshConfig) error {
	mr.called = true
	return nil
}

type mockReporter struct{ called bool }

func (mr *mockReporter) WriteReports(ctx context.Context, errs reporter.ResourceErrors, subresourceStatuses map[string]*core.Status) error {
	mr.called = true
	return nil
}
