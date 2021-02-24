package mesh

import (
	"context"

	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/output/istio"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/output/local"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/mesh/mtls"

	discoveryv1 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/input"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/reporting"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/mesh/access"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/mesh/federation"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/skv2/contrib/pkg/sets"
)

//go:generate mockgen -source ./istio_mesh_translator.go -destination mocks/istio_mesh_translator.go

// the VirtualService translator translates a Mesh into a VirtualService.
type Translator interface {
	// Translate translates the appropriate resources to apply the VirtualMesh to the given Mesh.
	// Output resources will be added to the istio.Builder
	// Errors caused by invalid user config will be reported using the Reporter.
	Translate(
		in input.LocalSnapshot,
		mesh *discoveryv1.Mesh,
		istioOutputs istio.Builder,
		localOutputs local.Builder,
		reporter reporting.Reporter,
	)
}

type translator struct {
	ctx                  context.Context
	mtlsTranslator       mtls.Translator
	federationTranslator federation.Translator
	accessTranslator     access.Translator
}

func NewTranslator(
	ctx context.Context,
	mtlsTranslator mtls.Translator,
	federationTranslator federation.Translator,
	accessTranslator access.Translator,
) Translator {
	return &translator{
		ctx:                  ctx,
		mtlsTranslator:       mtlsTranslator,
		federationTranslator: federationTranslator,
		accessTranslator:     accessTranslator,
	}
}

// translate the appropriate resources for the given Mesh.
func (t *translator) Translate(
	in input.LocalSnapshot,
	mesh *discoveryv1.Mesh,
	istioOutputs istio.Builder,
	localOutputs local.Builder,
	reporter reporting.Reporter,
) {
	istioMesh := mesh.Spec.GetIstio()
	if istioMesh == nil {
		contextutils.LoggerFrom(t.ctx).Debugf("ignoring non istio mesh %v %T", sets.Key(mesh), mesh.Spec.Type)
		return
	}

	// add mesh installation cluster to outputs
	istioOutputs.AddCluster(istioMesh.Installation.GetCluster())

	appliedVirtualMesh := mesh.Status.AppliedVirtualMesh
	if appliedVirtualMesh != nil {
		t.mtlsTranslator.Translate(mesh, appliedVirtualMesh, istioOutputs, localOutputs, reporter)
		t.federationTranslator.Translate(in, mesh, appliedVirtualMesh, istioOutputs, reporter)
		t.accessTranslator.Translate(mesh, appliedVirtualMesh, istioOutputs)
	}
}
