package translation

import (
	"bytes"
	"context"
	"fmt"

	appmeshv1beta2 "github.com/aws/aws-app-mesh-controller-for-k8s/apis/appmesh/v1beta2"
	smiaccessv1alpha2 "github.com/servicemeshinterface/smi-sdk-go/pkg/apis/access/v1alpha2"
	smispecsv1alpha3 "github.com/servicemeshinterface/smi-sdk-go/pkg/apis/specs/v1alpha3"
	smisplitv1alpha2 "github.com/servicemeshinterface/smi-sdk-go/pkg/apis/split/v1alpha2"
	certificatesv1 "github.com/solo-io/gloo-mesh/pkg/api/certificates.mesh.gloo.solo.io/v1"
	discoveryv1 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/input"
	appmeshoutput "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/output/appmesh"
	istiooutput "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/output/istio"
	localoutput "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/output/local"
	smioutput "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/output/smi"
	networkingv1 "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1"
	xdsv1beta1 "github.com/solo-io/gloo-mesh/pkg/api/xds.agent.enterprise.mesh.gloo.solo.io/v1beta1"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/reporting"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/appmesh"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/osm"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/metautils"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/skv2/contrib/pkg/output"
	"github.com/solo-io/skv2/pkg/ezkube"
	"github.com/solo-io/skv2/pkg/multicluster"
	"istio.io/client-go/pkg/apis/networking/v1alpha3"
	"istio.io/client-go/pkg/apis/security/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// The networking translator translates an input networking snapshot to an output snapshot of mesh specific config resources.
// The output snapshot represents all output resources currently enforced by the system given the input events processed thus far.
type Translator interface {
	// errors reflect an internal translation error and should never happen
	Translate(
		ctx context.Context,
		eventObjs []ezkube.ResourceId,
		in input.LocalSnapshot,
		userSupplied input.RemoteSnapshot,
		reporter reporting.Reporter,
	) (*Outputs, error)
}

type translator struct {
	totalTranslates   int // TODO(ilackarms): metric
	istioTranslator   istio.Translator
	appmeshTranslator appmesh.Translator
	osmTranslator     osm.Translator

	outputs *Outputs
}

func NewTranslator(
	istioTranslator istio.Translator,
	appmeshTranslator appmesh.Translator,
	osmTranslator osm.Translator,
) Translator {
	return &translator{
		istioTranslator:   istioTranslator,
		appmeshTranslator: appmeshTranslator,
		osmTranslator:     osmTranslator,
	}
}

// Translate all input objects into corresponding output objects.
// eventObjs is the set of objects for which events have occurred since the last invocation of Translate,
// which is used to limit processing to only what's relevant given the changed input objects.
func (t *translator) Translate(
	ctx context.Context,
	eventObjs []ezkube.ResourceId,
	in input.LocalSnapshot,
	userSupplied input.RemoteSnapshot,
	reporter reporting.Reporter,
) (*Outputs, error) {
	t.totalTranslates++
	ctx = contextutils.WithLogger(ctx, fmt.Sprintf("translation-%v", t.totalTranslates))

	istioOutputs := istiooutput.NewBuilder(ctx, fmt.Sprintf("networking-istio-%v", t.totalTranslates))
	appmeshOutputs := appmeshoutput.NewBuilder(ctx, fmt.Sprintf("networking-appmesh-%v", t.totalTranslates))
	smiOutputs := smioutput.NewBuilder(ctx, fmt.Sprintf("networking-smi-%v", t.totalTranslates))
	localOutputs := localoutput.NewBuilder(ctx, fmt.Sprintf("networking-local-%v", t.totalTranslates))

	t.istioTranslator.Translate(ctx, eventObjs, in, userSupplied, istioOutputs, localOutputs, reporter)
	t.appmeshTranslator.Translate(ctx, in, appmeshOutputs, reporter)
	t.osmTranslator.Translate(ctx, in, smiOutputs, reporter)

	// first translation, initialize outputs
	if t.outputs == nil {
		t.outputs = &Outputs{
			Istio:   istioOutputs,
			Appmesh: appmeshOutputs,
			Smi:     smiOutputs,
			Local:   localOutputs,
		}
		return t.outputs, nil
	}

	// update outputs
	t.outputs = updateOutputs(ctx, in, istioOutputs, appmeshOutputs, smiOutputs, localOutputs, t.outputs)

	return t.outputs, nil
}

// Update outputs by the following procedure:
//  1. insert all newly translated objects
//  2. insert any objects translated from previous translations, but filter out any objects that should be garbage collected
func updateOutputs(
	ctx context.Context,
	in input.LocalSnapshot,
	istioOutputs istiooutput.Builder,
	appmeshOutputs appmeshoutput.Builder,
	smiOutputs smioutput.Builder,
	localOutputs localoutput.Builder,
	oldOutputs *Outputs,
) *Outputs {
	// initialize new outputs with recently translated objects
	updatedOutputs := &Outputs{
		Istio:   istioOutputs,
		Appmesh: appmeshOutputs,
		Smi:     smiOutputs,
		Local:   localOutputs,
	}

	// insert outputs from previous translations that haven't been updated and don't require garbage collection
	// NOTE: the following block must be maintained with all output types.
	// TODO: leverage code gen?

	// Istio
	// AuthorizationPolicies
	oldAuthorizationPolicies := oldOutputs.Istio.GetAuthorizationPolicies().List(func(obj *v1beta1.AuthorizationPolicy) bool {
		return updatedOutputs.Istio.GetAuthorizationPolicies().Has(obj) || shouldGarbageCollect(ctx, in, obj)
	})
	updatedOutputs.Istio.AddAuthorizationPolicies(oldAuthorizationPolicies...)
	// DestinationRules
	oldDestinationRules := oldOutputs.Istio.GetDestinationRules().List(func(obj *v1alpha3.DestinationRule) bool {
		return updatedOutputs.Istio.GetDestinationRules().Has(obj) || shouldGarbageCollect(ctx, in, obj)
	})
	updatedOutputs.Istio.AddDestinationRules(oldDestinationRules...)
	// EnvoyFilters
	oldEnvoyFilters := oldOutputs.Istio.GetEnvoyFilters().List(func(obj *v1alpha3.EnvoyFilter) bool {
		return updatedOutputs.Istio.GetEnvoyFilters().Has(obj) || shouldGarbageCollect(ctx, in, obj)
	})
	updatedOutputs.Istio.AddEnvoyFilters(oldEnvoyFilters...)
	// Gateways
	oldGateways := oldOutputs.Istio.GetGateways().List(func(obj *v1alpha3.Gateway) bool {
		return updatedOutputs.Istio.GetGateways().Has(obj) || shouldGarbageCollect(ctx, in, obj)
	})
	updatedOutputs.Istio.AddGateways(oldGateways...)
	// ServiceEntries
	oldServiceEntries := oldOutputs.Istio.GetServiceEntries().List(func(obj *v1alpha3.ServiceEntry) bool {
		return updatedOutputs.Istio.GetServiceEntries().Has(obj) || shouldGarbageCollect(ctx, in, obj)
	})
	updatedOutputs.Istio.AddServiceEntries(oldServiceEntries...)
	// VirtualServices
	oldVirtualServices := oldOutputs.Istio.GetVirtualServices().List(func(obj *v1alpha3.VirtualService) bool {
		return updatedOutputs.Istio.GetVirtualServices().Has(obj) || shouldGarbageCollect(ctx, in, obj)
	})
	updatedOutputs.Istio.AddVirtualServices(oldVirtualServices...)
	// IssuedCertificates
	oldIssuedCertificates := oldOutputs.Istio.GetIssuedCertificates().List(func(obj *certificatesv1.IssuedCertificate) bool {
		return updatedOutputs.Istio.GetIssuedCertificates().Has(obj) || shouldGarbageCollect(ctx, in, obj)
	})
	updatedOutputs.Istio.AddIssuedCertificates(oldIssuedCertificates...)
	// PodBounceDirectives
	oldPodBounceDirectives := oldOutputs.Istio.GetPodBounceDirectives().List(func(obj *certificatesv1.PodBounceDirective) bool {
		return updatedOutputs.Istio.GetPodBounceDirectives().Has(obj) || shouldGarbageCollect(ctx, in, obj)
	})
	updatedOutputs.Istio.AddPodBounceDirectives(oldPodBounceDirectives...)
	// XdsConfigs
	oldXdsConfigs := oldOutputs.Istio.GetXdsConfigs().List(func(obj *xdsv1beta1.XdsConfig) bool {
		return updatedOutputs.Istio.GetXdsConfigs().Has(obj) || shouldGarbageCollect(ctx, in, obj)
	})
	updatedOutputs.Istio.AddXdsConfigs(oldXdsConfigs...)

	// AppMesh
	// AppMesh VirtualServices
	oldAppMeshVirtualServices := oldOutputs.Appmesh.GetVirtualServices().List(func(obj *appmeshv1beta2.VirtualService) bool {
		return updatedOutputs.Appmesh.GetVirtualServices().Has(obj) || shouldGarbageCollect(ctx, in, obj)
	})
	updatedOutputs.Appmesh.AddVirtualServices(oldAppMeshVirtualServices...)
	// AppMesh VirtualNodes
	oldAppMeshVirtualNodes := oldOutputs.Appmesh.GetVirtualNodes().List(func(obj *appmeshv1beta2.VirtualNode) bool {
		return updatedOutputs.Appmesh.GetVirtualNodes().Has(obj) || shouldGarbageCollect(ctx, in, obj)
	})
	updatedOutputs.Appmesh.AddVirtualNodes(oldAppMeshVirtualNodes...)
	// AppMesh VirtualRouters
	oldAppMeshVirtualRouters := oldOutputs.Appmesh.GetVirtualRouters().List(func(obj *appmeshv1beta2.VirtualRouter) bool {
		return updatedOutputs.Appmesh.GetVirtualRouters().Has(obj) || shouldGarbageCollect(ctx, in, obj)
	})
	updatedOutputs.Appmesh.AddVirtualRouters(oldAppMeshVirtualRouters...)

	// SMI
	// TrafficTargets
	oldSmiTrafficTargets := oldOutputs.Smi.GetTrafficTargets().List(func(obj *smiaccessv1alpha2.TrafficTarget) bool {
		return updatedOutputs.Smi.GetTrafficTargets().Has(obj) || shouldGarbageCollect(ctx, in, obj)
	})
	updatedOutputs.Smi.AddTrafficTargets(oldSmiTrafficTargets...)
	// HttpRouteGroups
	oldSmiHttpRouteGroups := oldOutputs.Smi.GetHTTPRouteGroups().List(func(obj *smispecsv1alpha3.HTTPRouteGroup) bool {
		return updatedOutputs.Smi.GetHTTPRouteGroups().Has(obj) || shouldGarbageCollect(ctx, in, obj)
	})
	updatedOutputs.Smi.AddHTTPRouteGroups(oldSmiHttpRouteGroups...)
	// TrafficSplits
	oldSmiTrafficSplits := oldOutputs.Smi.GetTrafficSplits().List(func(obj *smisplitv1alpha2.TrafficSplit) bool {
		return updatedOutputs.Smi.GetTrafficSplits().Has(obj) || shouldGarbageCollect(ctx, in, obj)
	})
	updatedOutputs.Smi.AddTrafficSplits(oldSmiTrafficSplits...)

	// Local outputs
	oldSecrets := oldOutputs.Local.GetSecrets().List(func(obj *corev1.Secret) bool {
		return updatedOutputs.Local.GetSecrets().Has(obj) || shouldGarbageCollect(ctx, in, obj)
	})
	updatedOutputs.Local.AddSecrets(oldSecrets...)

	return updatedOutputs
}

// Return true if the object should be garbage collected (i.e. deleted from k8s storage).
// Specifically, garbage collect the object if any of its parents no longer exist.
func shouldGarbageCollect(
	ctx context.Context,
	in input.LocalSnapshot,
	obj client.Object,
) bool {
	for parentGvk, resourceIds := range metautils.RetrieveParents(ctx, obj) {

		// NOTE: This block must be maintained with all relevant parent GVK's.
		switch parentGvk {

		// Discovery parents
		case discoveryv1.WorkloadGVK:
			for _, resourceId := range resourceIds {
				if !in.Workloads().Has(resourceId) {
					return true
				}
			}
		case discoveryv1.DestinationGVK:
			for _, resourceId := range resourceIds {
				if !in.Destinations().Has(resourceId) {
					return true
				}
			}
		case discoveryv1.MeshGVK:
			for _, resourceId := range resourceIds {
				if !in.Meshes().Has(resourceId) {
					return true
				}
			}
		// Networking parents
		case networkingv1.TrafficPolicyGVK:
			// TODO(harveyxia) does this generalize?
			// Because TrafficPolicies are merged, only garbage collect output if *all* parent TrafficPolicies no longer exist
			for _, resourceId := range resourceIds {
				if in.TrafficPolicies().Has(resourceId) {
					return false
				}
			}
			return true
		case networkingv1.AccessPolicyGVK:
			// Because AccessPolicies are merged, only garbage collect output if *all* parent AccessPolicies no longer exist
			for _, resourceId := range resourceIds {
				if in.AccessPolicies().Has(resourceId) {
					return false
				}
			}
			return true
		case networkingv1.VirtualMeshGVK:
			for _, resourceId := range resourceIds {
				if !in.VirtualMeshes().Has(resourceId) {
					return true
				}
			}
		}
	}

	return false
}

type Outputs struct {
	Istio   istiooutput.Builder
	Appmesh appmeshoutput.Builder
	Smi     smioutput.Builder
	Local   localoutput.Builder
}

func (t *Outputs) snapshots() (outputSnapshots, error) {
	istioSnapshot, err := t.Istio.BuildSinglePartitionedSnapshot(metautils.TranslatedObjectLabels())
	if err != nil {
		return outputSnapshots{}, err
	}

	appmeshSnapshot, err := t.Appmesh.BuildSinglePartitionedSnapshot(metautils.TranslatedObjectLabels())
	if err != nil {
		return outputSnapshots{}, err
	}

	smiSnapshot, err := t.Smi.BuildSinglePartitionedSnapshot(metautils.TranslatedObjectLabels())
	if err != nil {
		return outputSnapshots{}, err
	}

	localSnapshot, err := t.Local.BuildSinglePartitionedSnapshot(metautils.TranslatedObjectLabels())
	if err != nil {
		return outputSnapshots{}, err
	}

	return outputSnapshots{
		istio:   istioSnapshot,
		appmesh: appmeshSnapshot,
		smi:     smiSnapshot,
		local:   localSnapshot,
	}, nil
}

func (t *Outputs) MarshalJSON() ([]byte, error) {
	snaps, err := t.snapshots()
	if err != nil {
		return nil, err
	}
	return snaps.MarshalJSON()
}

func (t *Outputs) ApplyMultiCluster(
	ctx context.Context,
	clusterClient client.Client,
	multiClusterClient multicluster.Client,
	errHandler output.ErrorHandler,
) error {
	snaps, err := t.snapshots()
	if err != nil {
		return err
	}
	// Apply mesh resources to registered clusters
	snaps.istio.ApplyMultiCluster(ctx, multiClusterClient, errHandler)
	snaps.appmesh.ApplyMultiCluster(ctx, multiClusterClient, errHandler)
	snaps.smi.ApplyMultiCluster(ctx, multiClusterClient, errHandler)
	// Apply local resources only to management cluster
	snaps.local.ApplyLocalCluster(ctx, clusterClient, errHandler)

	return nil
}

type outputSnapshots struct {
	istio   istiooutput.Snapshot
	appmesh appmeshoutput.Snapshot
	smi     smioutput.Snapshot
	local   localoutput.Snapshot
}

func (t outputSnapshots) MarshalJSON() ([]byte, error) {

	istioByt, err := t.istio.MarshalJSON()
	if err != nil {
		return nil, err
	}
	appmeshByt, err := t.appmesh.MarshalJSON()
	if err != nil {
		return nil, err
	}
	smiByt, err := t.smi.MarshalJSON()
	if err != nil {
		return nil, err
	}
	localByt, err := t.local.MarshalJSON()
	if err != nil {
		return nil, err
	}
	return bytes.Join([][]byte{istioByt, appmeshByt, smiByt, localByt}, []byte("\n")), nil
}
