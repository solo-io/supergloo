package failoverservice

import (
	"context"

	"github.com/solo-io/go-utils/contextutils"
	discoveryv1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/snapshot/input"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/reporting"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/mesh/failoverservice/validation"
	"github.com/solo-io/skv2/contrib/pkg/sets"
	networkingv1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
)

// Outputs of translating a FailoverService for a single Mesh
type Outputs struct {
	EnvoyFilter    *networkingv1alpha3.EnvoyFilter
	ServiceEntries *networkingv1alpha3.ServiceEntry
}

// The FailoverService translator translates a FailoverService for a single Mesh.
type Translator interface {
	// Translate translates the FailoverService into a ServiceEntry representing the new service and an accompanying EnvoyFilter.
	// Errors caused by invalid user config will be reported using the Reporter.
	Translate(
		in input.Snapshot,
		mesh *discoveryv1alpha2.Mesh,
		failoverService *discoveryv1alpha2.MeshStatus_AppliedFailoverService,
		reporter reporting.Reporter,
	) Outputs
}

type translator struct {
	ctx       context.Context
	validator validation.FailoverServiceValidator
}

func NewTranslator(ctx context.Context) Translator {
	return &translator{
		ctx:       ctx,
		validator: validation.NewFailoverServiceValidator(),
	}
}

func (t *translator) Translate(
	in input.Snapshot,
	mesh *discoveryv1alpha2.Mesh,
	failoverService *discoveryv1alpha2.MeshStatus_AppliedFailoverService,
	reporter reporting.Reporter,
) Outputs {
	istioMesh := mesh.Spec.GetIstio()
	if istioMesh == nil {
		contextutils.LoggerFrom(t.ctx).Debugf("ignoring non istio mesh %v %T", sets.Key(mesh), mesh.Spec.MeshType)
		return Outputs{}
	}

	// If validation fails, report the errors to the Meshes and do not translate.
	validationErr := t.validator.Validate(validation.Inputs{
		MeshServices:  in.MeshServices(),
		KubeClusters:  in.KubernetesClusters(),
		Meshes:        in.Meshes(),
		VirtualMeshes: in.VirtualMeshes(),
	}, failoverService.Spec)
	if validationErr != nil {
		for _, meshRef := range failoverService.Spec.Meshes {
			mesh, err := in.Meshes().Find(meshRef)
			if err != nil {
				continue // Mesh reference not found
			}
			reporter.ReportFailoverService(mesh, failoverService.Ref, validationErr)
		}
		return Outputs{}
	}

	// TODO(harveyxia) translate
}
