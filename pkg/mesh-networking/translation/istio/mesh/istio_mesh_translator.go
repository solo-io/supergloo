package mesh

import (
	"context"

	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/istio/output"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/mesh/mtls"

	"github.com/solo-io/go-utils/contextutils"
	discoveryv1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/istio/input"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/reporting"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/mesh/access"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/mesh/failoverservice"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/mesh/federation"
	"github.com/solo-io/skv2/contrib/pkg/sets"
)

//go:generate mockgen -source ./istio_mesh_translator.go -destination mocks/istio_mesh_translator.go

// the VirtualService translator translates a Mesh into a VirtualService.
type Translator interface {
	// Translate translates the appropriate resources to apply the VirtualMesh to the given Mesh.
	// Output resources will be added to the output.Builder
	// Errors caused by invalid user config will be reported using the Reporter.
	Translate(
		in input.Snapshot,
		mesh *discoveryv1alpha2.Mesh,
		outputs output.Builder,
		reporter reporting.Reporter,
	)
}

type translator struct {
	ctx                       context.Context
	mtlsTranslator            mtls.Translator
	federationTranslator      federation.Translator
	accessTranslator          access.Translator
	failoverServiceTranslator failoverservice.Translator
}

func NewTranslator(
	ctx context.Context,
	mtlsTranslator mtls.Translator,
	federationTranslator federation.Translator,
	accessTranslator access.Translator,
	failoverServiceTranslator failoverservice.Translator,
) Translator {
	return &translator{
		ctx:                       ctx,
		mtlsTranslator:            mtlsTranslator,
		federationTranslator:      federationTranslator,
		accessTranslator:          accessTranslator,
		failoverServiceTranslator: failoverServiceTranslator,
	}
}

// translate the appropriate resources for the given Mesh.
func (t *translator) Translate(
	in input.Snapshot,
	mesh *discoveryv1alpha2.Mesh,
	outputs output.Builder,
	reporter reporting.Reporter,
) {
	istioMesh := mesh.Spec.GetIstio()
	if istioMesh == nil {
		contextutils.LoggerFrom(t.ctx).Debugf("ignoring non istio mesh %v %T", sets.Key(mesh), mesh.Spec.MeshType)
		return
	}

	// add mesh installation cluster to outputs
	outputs.AddCluster(istioMesh.Installation.GetCluster())

	appliedVirtualMesh := mesh.Status.AppliedVirtualMesh
	if appliedVirtualMesh != nil {
		t.mtlsTranslator.Translate(mesh, appliedVirtualMesh, outputs, reporter)
		t.federationTranslator.Translate(in, mesh, appliedVirtualMesh, outputs, reporter)
		t.accessTranslator.Translate(mesh, appliedVirtualMesh, outputs)
	}

	for _, failoverService := range mesh.Status.AppliedFailoverServices {
		t.failoverServiceTranslator.Translate(in, mesh, failoverService, outputs, reporter)
	}
}
