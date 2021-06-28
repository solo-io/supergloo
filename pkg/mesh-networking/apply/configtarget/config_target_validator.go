package configtarget

import (
	"fmt"
	"sort"

	"github.com/hashicorp/go-multierror"
	"github.com/rotisserie/eris"
	commonv1 "github.com/solo-io/gloo-mesh/pkg/api/common.mesh.gloo.solo.io/v1"
	discoveryv1sets "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1/sets"
	v1 "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1"
	v1sets "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1/sets"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/mesh/federation"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/selectorutils"
	"github.com/solo-io/skv2/contrib/pkg/sets"
	skv2corev1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	"github.com/solo-io/skv2/pkg/ezkube"
)

/*
	Validate configuration target references in networking configuration resources, and report
	any errors (i.e. references to non-existent discovery entities) to the offending resource status.
*/
type ConfigTargetValidator interface {
	// Validate Destination references declared on TrafficPolicies.
	ValidateTrafficPolicies(
		trafficPolicies v1.TrafficPolicySlice,
	)

	// Validate mesh references declared on VirtualMeshes.
	// Also validate that all referenced meshes are contained in at most one VirtualMesh.
	ValidateVirtualMeshes(
		virtualMeshes v1.VirtualMeshSlice,
	)

	// Validate Destination references declared on AccessPolicies.
	ValidateAccessPolicies(
		accessPolicies v1.AccessPolicySlice,
	)
}

type configTargetValidator struct {
	meshes       discoveryv1sets.MeshSet
	destinations discoveryv1sets.DestinationSet
}

func NewConfigTargetValidator(
	meshes discoveryv1sets.MeshSet,
	destinations discoveryv1sets.DestinationSet,
) ConfigTargetValidator {
	return &configTargetValidator{
		meshes:       meshes,
		destinations: destinations,
	}
}

func (c *configTargetValidator) ValidateVirtualMeshes(virtualMeshes v1.VirtualMeshSlice) {
	for _, virtualMesh := range virtualMeshes {
		errs := c.validateVirtualMesh(virtualMesh)
		if len(errs) == 0 {
			continue
		}
		virtualMesh.Status.State = commonv1.ApprovalState_INVALID
		virtualMesh.Status.Errors = getErrStrings(errs)
	}

	validateOneVirtualMeshPerMesh(virtualMeshes)
	c.validateVirtualMeshIngressGatewaySelectors(virtualMeshes)
}

func (c *configTargetValidator) ValidateTrafficPolicies(trafficPolicies v1.TrafficPolicySlice) {
	for _, trafficPolicy := range trafficPolicies {
		errs := c.validateDestinationReferences(trafficPolicy.Spec.DestinationSelector)
		if len(errs) == 0 {
			continue
		}
		trafficPolicy.Status.State = commonv1.ApprovalState_INVALID
		trafficPolicy.Status.Errors = getErrStrings(errs)
	}
}

func (c *configTargetValidator) ValidateAccessPolicies(accessPolicies v1.AccessPolicySlice) {
	for _, accessPolicy := range accessPolicies {
		errs := c.validateDestinationReferences(accessPolicy.Spec.DestinationSelector)
		if len(errs) == 0 {
			continue
		}
		accessPolicy.Status.State = commonv1.ApprovalState_INVALID
		accessPolicy.Status.Errors = getErrStrings(errs)
	}
}

func (c *configTargetValidator) validateMeshReferences(meshRefs []*skv2corev1.ObjectRef) []error {
	var errs []error
	for _, meshRef := range meshRefs {
		if err := validateObjectRef(meshRef); err != nil {
			errs = append(errs, eris.Wrap(err, "malformed meshRef"))
		} else if _, err := c.meshes.Find(meshRef); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

func (c *configTargetValidator) validateDestinationReferences(serviceSelectors []*commonv1.DestinationSelector) []error {
	var errs []error
	for _, destinationSelector := range serviceSelectors {
		kubeServiceRefs := destinationSelector.GetKubeServiceRefs()
		// only validate Destinations selected by direct reference
		if kubeServiceRefs == nil {
			continue
		}
		for _, ref := range kubeServiceRefs.Services {
			if err := validateClusterObjectRef(ref); err != nil {
				errs = append(errs, eris.Wrap(err, "malformed kubeServiceRef"))
			} else if !c.kubeServiceExists(ref) {
				errs = append(errs, eris.Errorf("Destination %s not found", sets.Key(ref)))
			}
		}
	}
	return errs
}

func (c *configTargetValidator) kubeServiceExists(ref *skv2corev1.ClusterObjectRef) bool {
	for _, destination := range c.destinations.List() {
		kubeService := destination.Spec.GetKubeService()
		if kubeService == nil {
			continue
		}
		if ezkube.ClusterRefsMatch(ref, kubeService.Ref) {
			return true
		}
	}
	return false
}

func (c *configTargetValidator) validateVirtualMesh(virtualMesh *v1.VirtualMesh) []error {
	var errs []error
	meshRefErrors := c.validateMeshReferences(virtualMesh.Spec.Meshes)
	if meshRefErrors != nil {
		errs = append(errs, meshRefErrors...)
	}
	return errs
}

func getErrStrings(errs []error) []string {
	var errStrings []string
	for _, err := range errs {
		errStrings = append(errStrings, err.Error())
	}
	return errStrings
}

// For each VirtualMesh, sort them by accepted date, then invalidate if it applies to a Mesh that
// is already grouped into a VirtualMesh.
func validateOneVirtualMeshPerMesh(virtualMeshes []*v1.VirtualMesh) {
	sortVirtualMeshesByAcceptedDate(virtualMeshes)

	vMeshesPerMesh := map[string]*v1.VirtualMesh{}
	invalidVirtualMeshes := v1sets.NewVirtualMeshSet()

	// track accepted index
	var acceptedIndex uint32
	// Invalidate VirtualMesh if it applies to a Mesh that already has an applied VirtualMesh.
	for _, vMesh := range virtualMeshes {
		if vMesh.Status.State != commonv1.ApprovalState_ACCEPTED {
			continue
		}
		vMesh := vMesh
		for _, mesh := range vMesh.Spec.Meshes {
			// Ignore VirtualMesh if previously invalidated.
			if invalidVirtualMeshes.Has(vMesh) {
				continue
			}
			meshKey := sets.Key(mesh)
			existingVirtualMesh, ok := vMeshesPerMesh[meshKey]
			vMesh.Status.ObservedGeneration = vMesh.Generation
			if !ok {
				vMeshesPerMesh[meshKey] = vMesh
				acceptedIndex++
			} else {
				vMesh.Status.State = commonv1.ApprovalState_INVALID
				vMesh.Status.Errors = append(
					vMesh.Status.Errors,
					fmt.Sprintf("Includes a Mesh (%s.%s) that already is grouped in a VirtualMesh (%s.%s)",
						mesh.Name, mesh.Namespace,
						existingVirtualMesh.Name, existingVirtualMesh.Namespace,
					),
				)
			}
			invalidVirtualMeshes.Insert(vMesh)
		}
	}
}

// For each VirtualMesh, if it has ingress gateway selectors, validate those selectors.
func (c *configTargetValidator) validateVirtualMeshIngressGatewaySelectors(virtualMeshes []*v1.VirtualMesh) {
	// Invalidate VirtualMesh if it applies to a Mesh that already has an applied VirtualMesh.
	for _, vMesh := range virtualMeshes {
		if vMesh.Status.State != commonv1.ApprovalState_ACCEPTED {
			continue
		}
		vMesh := vMesh

		if len(vMesh.Spec.GetFederation().GetEastWestIngressGatewaySelectors()) > 0 {
			// Validate Ingress gateway service selectors
			matchedDestinations := 0
			for _, ingressGatewayServiceSelector := range vMesh.Spec.GetFederation().GetEastWestIngressGatewaySelectors() {
				if errs := c.validateDestinationReferences(ingressGatewayServiceSelector.GetDestinationSelectors()); len(errs) != 0 {
					vMesh.Status.State = commonv1.ApprovalState_INVALID
					for _, err := range errs {
						vMesh.Status.Errors = append(
							vMesh.Status.Errors,
							fmt.Sprintf("Invalid Destination selector: %v", err),
						)
					}
					continue
				}

				gatewayTlsPortName := "tls"
				if ingressGatewayServiceSelector.GetGatewayTlsPortName() != "" {
					gatewayTlsPortName = ingressGatewayServiceSelector.GetGatewayTlsPortName()
				}
				for _, destination := range c.destinations.List() {
					if selectorutils.SelectorMatchesDestination(ingressGatewayServiceSelector.GetDestinationSelectors(), destination) {
						if len(destination.Spec.GetKubeService().GetWorkloadSelectorLabels()) != 0 {
							if len(destination.Spec.GetKubeService().GetWorkloadSelectorLabels()) == 0 {
								vMesh.Status.State = commonv1.ApprovalState_INVALID
								vMesh.Status.Errors = append(
									vMesh.Status.Errors,
									fmt.Sprintf("Attempting to select ingress gateway destination %v with no workload labels", sets.Key(destination)),
								)
								continue
							}
						}
						hasPort := false
						for _, ports := range destination.Spec.GetKubeService().GetPorts() {
							if ports.GetName() == gatewayTlsPortName {
								hasPort = true
								break
							}
						}
						if !hasPort {
							vMesh.Status.State = commonv1.ApprovalState_INVALID
							vMesh.Status.Errors = append(
								vMesh.Status.Errors,
								fmt.Sprintf("Attempting to select ingress gateway destination %v with no port named %v",
									sets.Key(destination), gatewayTlsPortName),
							)
							continue
						}
					}
					matchedDestinations++
				}
				if matchedDestinations == 0 {
					vMesh.Status.State = commonv1.ApprovalState_INVALID
					vMesh.Status.Errors = append(
						vMesh.Status.Errors,
						fmt.Sprintf("No Destinations selected as ingress gateway. At least one must be selected."),
					)
				}
			}
		} else {
			// Check the mesh specific default exists
			defaultExists := false
			for _, destination := range c.destinations.List() {
				if destination.Spec.GetKubeService().GetWorkloadSelectorLabels()["istio"] == "ingressgateway" {
					for _, ports := range destination.Spec.GetKubeService().GetPorts() {
						if ports.GetName() == federation.DefaultGatewayPortName && ports.GetPort() != 0 {
							defaultExists = true
						}
					}
				}
			}
			if !defaultExists {
				vMesh.Status.State = commonv1.ApprovalState_INVALID
				vMesh.Status.Errors = append(
					vMesh.Status.Errors,
					fmt.Sprintf("Unable to find default ingress gateway."),
				)
			}
		}
	}
}

// sort the set of VirtualMeshes in the order in which they were accepted.
// VMeshes which were accepted first and have not changed (i.e. their observedGeneration is up-to-date) take precedence.
// Next are vMeshes that were previously accepted but whose observedGeneration is out of date. This permits vmeshes which were modified but formerly correct to maintain
// their acceptance status ahead of vmeshes which were unmodified and previously rejected.
// Next will be the vmeshes which have been modified and rejected.
// Finally, vmeshes which are rejected and modified
func sortVirtualMeshesByAcceptedDate(virtualMeshes v1.VirtualMeshSlice) {
	isUpToDate := func(vm *v1.VirtualMesh) bool {
		return vm.Status.ObservedGeneration == vm.Generation
	}

	sort.SliceStable(virtualMeshes, func(i, j int) bool {
		vMesh1, vMesh2 := virtualMeshes[i], virtualMeshes[j]

		state1 := vMesh1.Status.State
		state2 := vMesh2.Status.State

		switch {
		case state1 == commonv1.ApprovalState_ACCEPTED:
			if state2 != commonv1.ApprovalState_ACCEPTED {
				// accepted comes before non accepted
				return true
			}

			if vMesh1UpToDate := isUpToDate(vMesh1); vMesh1UpToDate != isUpToDate(vMesh2) {
				// up to date is validated before modified
				return vMesh1UpToDate
			}

			return true
		case state2 == commonv1.ApprovalState_ACCEPTED:
			// accepted comes before non accepted
			return false
		default:
			// neither policy has been accepted, we can simply sort by unique key
			return sets.Key(vMesh1) < sets.Key(vMesh2)
		}
	})
}

// return error if any field in ClusterObjectRef is empty
func validateClusterObjectRef(ref *skv2corev1.ClusterObjectRef) error {
	err := validateObjectRef(&skv2corev1.ObjectRef{
		Name:      ref.Name,
		Namespace: ref.Namespace,
	})
	if ref.ClusterName == "" {
		err = multierror.Append(err, eris.New("'clusterName' must be specified'"))
	}
	return err
}

// return error if any field in ClusterObjectRef is empty
func validateObjectRef(ref *skv2corev1.ObjectRef) error {
	var err error
	if ref.Name == "" {
		err = multierror.Append(err, eris.New("'name' must be specified'"))
	}
	if ref.Namespace == "" {
		err = multierror.Append(err, eris.New("'namespace' must be specified'"))
	}
	return err
}
