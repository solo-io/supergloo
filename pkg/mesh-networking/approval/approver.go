package approval

import (
	"context"
	"sort"

	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation"

	"github.com/solo-io/go-utils/contextutils"
	discoveryv1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/istio/input"
	networkingv1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/utils/selectorutils"
	"github.com/solo-io/skv2/contrib/pkg/sets"
	"github.com/solo-io/skv2/pkg/ezkube"
)

// the approver validates user-applied configuration
// and produces a snapshot that is ready for translation (i.e. with accepted policies applied to all the Status of all targeted MeshServices)
// Note that the Approver also updates the statuses of objects contained in the input Snapshot.
// The Input Snapshot's SyncStatuses method should usually be called
// after running the approver.
type Approver interface {
	Approve(ctx context.Context, input input.Snapshot)
}

type approver struct {
	// the approver runs the networking translator in order to detect & report translation errors
	translator translation.Translator
}

func NewApprover(
	translator translation.Translator,
) Approver {
	return &approver{
		translator: translator,
	}
}

func (v *approver) Approve(ctx context.Context, input input.Snapshot) {
	ctx = contextutils.WithLogger(ctx, "validation")
	reporter := newApprovalReporter()

	virtualMeshes := input.VirtualMeshes().List()
	validateOneVirtualMeshPerMesh(virtualMeshes)

	for _, meshService := range input.MeshServices().List() {
		meshService.Status.AppliedTrafficPolicies = getAppliedTrafficPolicies(input.TrafficPolicies().List(), meshService)
		meshService.Status.AppliedAccessPolicies = getAppliedAccessPolicies(input.AccessPolicies().List(), meshService)
	}
	for _, mesh := range input.Meshes().List() {
		mesh.Status.AppliedVirtualMesh = getAppliedVirtualMesh(virtualMeshes, mesh)
		mesh.Status.AppliedFailoverServices = getAppliedFailoverServices(input.FailoverServices().List(), mesh)
	}

	_, err := v.translator.Translate(ctx, input, reporter)
	if err != nil {
		// should never happen
		contextutils.LoggerFrom(ctx).DPanicf("internal error: failed to run translator: %v", err)
	}

	// initialize traffic policy statuses
	for _, trafficPolicy := range input.TrafficPolicies().List() {
		trafficPolicy.Status = networkingv1alpha2.TrafficPolicyStatus{
			State:              networkingv1alpha2.ApprovalState_ACCEPTED,
			ObservedGeneration: trafficPolicy.Generation,
			MeshServices:       map[string]*networkingv1alpha2.ApprovalStatus{},
		}
	}

	// initialize access policy statuses
	for _, accessPolicy := range input.AccessPolicies().List() {
		accessPolicy.Status = networkingv1alpha2.AccessPolicyStatus{
			State:              networkingv1alpha2.ApprovalState_ACCEPTED,
			ObservedGeneration: accessPolicy.Generation,
			MeshServices:       map[string]*networkingv1alpha2.ApprovalStatus{},
		}
	}

	// initialize FailoverService statuses
	for _, failoverService := range input.FailoverServices().List() {
		failoverService.Status = networkingv1alpha2.FailoverServiceStatus{
			State:              networkingv1alpha2.ApprovalState_ACCEPTED,
			ObservedGeneration: failoverService.Generation,
			Meshes:             map[string]*networkingv1alpha2.ApprovalStatus{},
		}
	}

	// By this point, VirtualMeshes have already undergone pre-translation validation.
	// If VirtualMesh is accepted by previous validation, reset its status so it can be populated by translation.
	for _, virtualMesh := range virtualMeshes {
		if virtualMesh.Status.State != networkingv1alpha2.ApprovalState_ACCEPTED {
			continue
		}
		virtualMesh.Status.ObservedGeneration = virtualMesh.Generation
		virtualMesh.Status.Meshes = map[string]*networkingv1alpha2.ApprovalStatus{}
	}

	// update meshservice, trafficpolicy, and accesspolicy statuses
	for _, meshService := range input.MeshServices().List() {
		meshService.Status.ObservedGeneration = meshService.Generation
		meshService.Status.AppliedTrafficPolicies = validateAndReturnApprovedTrafficPolicies(ctx, input, reporter, meshService)
		meshService.Status.AppliedAccessPolicies = validateAndReturnApprovedAccessPolicies(ctx, input, reporter, meshService)
	}

	for _, mesh := range input.Meshes().List() {
		mesh.Status.ObservedGeneration = mesh.Generation
		mesh.Status.AppliedFailoverServices = validateAndReturnApprovedFailoverServices(ctx, input, reporter, mesh)
		mesh.Status.AppliedVirtualMesh = validateAndReturnVirtualMesh(ctx, input, reporter, mesh)
	}

	// TODO(ilackarms): validate meshworkloads
}

// this function both validates the status of TrafficPolicies (sets error or accepted state)
// as well as returns a list of accepted traffic policies for the mesh service status
func validateAndReturnApprovedTrafficPolicies(ctx context.Context, input input.Snapshot, reporter *approvalReporter, meshService *discoveryv1alpha2.MeshService) []*discoveryv1alpha2.MeshServiceStatus_AppliedTrafficPolicy {
	var validatedTrafficPolicies []*discoveryv1alpha2.MeshServiceStatus_AppliedTrafficPolicy

	// track accepted index
	var acceptedIndex uint32
	for _, appliedTrafficPolicy := range meshService.Status.AppliedTrafficPolicies {
		errsForTrafficPolicy := reporter.getTrafficPolicyErrors(meshService, appliedTrafficPolicy.Ref)

		trafficPolicy, err := input.TrafficPolicies().Find(appliedTrafficPolicy.Ref)
		if err != nil {
			// should never happen
			contextutils.LoggerFrom(ctx).Errorf("internal error: failed to look up applied traffic policy %v: %v", appliedTrafficPolicy.Ref, err)
			continue
		}

		if len(errsForTrafficPolicy) == 0 {
			trafficPolicy.Status.MeshServices[sets.Key(meshService)] = &networkingv1alpha2.ApprovalStatus{
				AcceptanceOrder: acceptedIndex,
				State:           networkingv1alpha2.ApprovalState_ACCEPTED,
			}
			validatedTrafficPolicies = append(validatedTrafficPolicies, appliedTrafficPolicy)
			acceptedIndex++
		} else {
			var errMsgs []string
			for _, tpErr := range errsForTrafficPolicy {
				errMsgs = append(errMsgs, tpErr.Error())
			}
			trafficPolicy.Status.MeshServices[sets.Key(meshService)] = &networkingv1alpha2.ApprovalStatus{
				State:  networkingv1alpha2.ApprovalState_INVALID,
				Errors: errMsgs,
			}
			trafficPolicy.Status.State = networkingv1alpha2.ApprovalState_INVALID
		}
	}

	return validatedTrafficPolicies
}

// this function both validates the status of AccessPolicies (sets error or accepted state)
// as well as returns a list of accepted AccessPolicies for the mesh service status
func validateAndReturnApprovedAccessPolicies(
	ctx context.Context,
	input input.Snapshot,
	reporter *approvalReporter,
	meshService *discoveryv1alpha2.MeshService,
) []*discoveryv1alpha2.MeshServiceStatus_AppliedAccessPolicy {
	var validatedAccessPolicies []*discoveryv1alpha2.MeshServiceStatus_AppliedAccessPolicy

	// track accepted index
	var acceptedIndex uint32
	for _, appliedAccessPolicy := range meshService.Status.AppliedAccessPolicies {
		errsForAccessPolicy := reporter.getAccessPolicyErrors(meshService, appliedAccessPolicy.Ref)

		accessPolicy, err := input.AccessPolicies().Find(appliedAccessPolicy.Ref)
		if err != nil {
			// should never happen
			contextutils.LoggerFrom(ctx).Errorf("internal error: failed to look up applied AccessPolicy %v: %v", appliedAccessPolicy.Ref, err)
			continue
		}

		if len(errsForAccessPolicy) == 0 {
			accessPolicy.Status.MeshServices[sets.Key(meshService)] = &networkingv1alpha2.ApprovalStatus{
				AcceptanceOrder: acceptedIndex,
				State:           networkingv1alpha2.ApprovalState_ACCEPTED,
			}
			validatedAccessPolicies = append(validatedAccessPolicies, appliedAccessPolicy)
			acceptedIndex++
		} else {
			var errMsgs []string
			for _, apErr := range errsForAccessPolicy {
				errMsgs = append(errMsgs, apErr.Error())
			}
			accessPolicy.Status.MeshServices[sets.Key(meshService)] = &networkingv1alpha2.ApprovalStatus{
				State:  networkingv1alpha2.ApprovalState_INVALID,
				Errors: errMsgs,
			}
			accessPolicy.Status.State = networkingv1alpha2.ApprovalState_INVALID
		}
	}

	return validatedAccessPolicies
}

func validateAndReturnApprovedFailoverServices(
	ctx context.Context,
	input input.Snapshot,
	reporter *approvalReporter,
	mesh *discoveryv1alpha2.Mesh,
) []*discoveryv1alpha2.MeshStatus_AppliedFailoverService {
	var validatedFailoverServices []*discoveryv1alpha2.MeshStatus_AppliedFailoverService

	// track accepted index
	var acceptedIndex uint32
	for _, appliedFailoverService := range mesh.Status.AppliedFailoverServices {
		errsForFailoverService := reporter.getFailoverServiceErrors(mesh, appliedFailoverService.Ref)

		failoverService, err := input.FailoverServices().Find(appliedFailoverService.Ref)
		if err != nil {
			// should never happen
			contextutils.LoggerFrom(ctx).Errorf("internal error: failed to look up applied FailoverService %v: %v", appliedFailoverService.Ref, err)
			continue
		}

		if len(errsForFailoverService) == 0 {
			failoverService.Status.Meshes[sets.Key(mesh)] = &networkingv1alpha2.ApprovalStatus{
				AcceptanceOrder: acceptedIndex,
				State:           networkingv1alpha2.ApprovalState_ACCEPTED,
			}
			validatedFailoverServices = append(validatedFailoverServices, appliedFailoverService)
			acceptedIndex++
		} else {
			var errMsgs []string
			for _, fsErr := range errsForFailoverService {
				errMsgs = append(errMsgs, fsErr.Error())
			}
			failoverService.Status.Meshes[sets.Key(mesh)] = &networkingv1alpha2.ApprovalStatus{
				State:  networkingv1alpha2.ApprovalState_INVALID,
				Errors: errMsgs,
			}
			failoverService.Status.State = networkingv1alpha2.ApprovalState_INVALID
		}
	}

	return validatedFailoverServices
}

func validateAndReturnVirtualMesh(
	ctx context.Context,
	input input.Snapshot,
	reporter *approvalReporter,
	mesh *discoveryv1alpha2.Mesh,
) *discoveryv1alpha2.MeshStatus_AppliedVirtualMesh {
	appliedVirtualMesh := mesh.Status.AppliedVirtualMesh
	if appliedVirtualMesh == nil {
		return nil
	}
	errsForVirtualMesh := reporter.getVirtualMeshErrors(mesh, appliedVirtualMesh.Ref)

	virtualMesh, err := input.VirtualMeshes().Find(appliedVirtualMesh.Ref)
	if err != nil {
		// should never happen
		contextutils.LoggerFrom(ctx).Errorf("internal error: failed to look up applied VirtualMesh %v: %v", appliedVirtualMesh.Ref, err)
		return nil
	}

	if len(errsForVirtualMesh) == 0 {
		virtualMesh.Status.Meshes[sets.Key(mesh)] = &networkingv1alpha2.ApprovalStatus{
			State: networkingv1alpha2.ApprovalState_ACCEPTED,
		}
		return appliedVirtualMesh
	} else {
		var errMsgs []string
		for _, fsErr := range errsForVirtualMesh {
			errMsgs = append(errMsgs, fsErr.Error())
		}
		virtualMesh.Status.Meshes[sets.Key(mesh)] = &networkingv1alpha2.ApprovalStatus{
			State:  networkingv1alpha2.ApprovalState_INVALID,
			Errors: errMsgs,
		}
		virtualMesh.Status.State = networkingv1alpha2.ApprovalState_INVALID
		return nil
	}
}

// the approval reporter validates individual policies and reports any encountered errors
type approvalReporter struct {
	// NOTE(ilackarms): map access should be synchronous (called in a single context),
	// so locking should not be necessary.
	unapprovedTrafficPolicies  map[*discoveryv1alpha2.MeshService]map[string][]error
	unapprovedAccessPolicies   map[*discoveryv1alpha2.MeshService]map[string][]error
	unapprovedFailoverServices map[*discoveryv1alpha2.Mesh]map[string][]error
	unapprovedVirtualMeshes    map[*discoveryv1alpha2.Mesh]map[string][]error
	invalidFailoverServices    map[string][]error
}

func newApprovalReporter() *approvalReporter {
	return &approvalReporter{
		unapprovedTrafficPolicies:  map[*discoveryv1alpha2.MeshService]map[string][]error{},
		unapprovedAccessPolicies:   map[*discoveryv1alpha2.MeshService]map[string][]error{},
		unapprovedFailoverServices: map[*discoveryv1alpha2.Mesh]map[string][]error{},
		unapprovedVirtualMeshes:    map[*discoveryv1alpha2.Mesh]map[string][]error{},
		invalidFailoverServices:    map[string][]error{},
	}
}

// mark the policy with an error; will be used to filter the policy out of
// the accepted status later
func (v *approvalReporter) ReportTrafficPolicyToMeshService(meshService *discoveryv1alpha2.MeshService, trafficPolicy ezkube.ResourceId, err error) {
	invalidTrafficPoliciesForMeshService := v.unapprovedTrafficPolicies[meshService]
	if invalidTrafficPoliciesForMeshService == nil {
		invalidTrafficPoliciesForMeshService = map[string][]error{}
	}
	key := sets.Key(trafficPolicy)
	errs := invalidTrafficPoliciesForMeshService[key]
	errs = append(errs, err)
	invalidTrafficPoliciesForMeshService[key] = errs
	v.unapprovedTrafficPolicies[meshService] = invalidTrafficPoliciesForMeshService
}

func (v *approvalReporter) ReportAccessPolicyToMeshService(meshService *discoveryv1alpha2.MeshService, accessPolicy ezkube.ResourceId, err error) {
	invalidAccessPoliciesForMeshService := v.unapprovedAccessPolicies[meshService]
	if invalidAccessPoliciesForMeshService == nil {
		invalidAccessPoliciesForMeshService = map[string][]error{}
	}
	key := sets.Key(accessPolicy)
	errs := invalidAccessPoliciesForMeshService[key]
	errs = append(errs, err)
	invalidAccessPoliciesForMeshService[key] = errs
	v.unapprovedAccessPolicies[meshService] = invalidAccessPoliciesForMeshService
}

func (v *approvalReporter) ReportVirtualMeshToMesh(mesh *discoveryv1alpha2.Mesh, virtualMesh ezkube.ResourceId, err error) {
	invalidVirtualMeshesForMesh := v.unapprovedVirtualMeshes[mesh]
	if invalidVirtualMeshesForMesh == nil {
		invalidVirtualMeshesForMesh = map[string][]error{}
	}
	key := sets.Key(virtualMesh)
	errs := invalidVirtualMeshesForMesh[key]
	errs = append(errs, err)
	invalidVirtualMeshesForMesh[key] = errs
	v.unapprovedVirtualMeshes[mesh] = invalidVirtualMeshesForMesh
}

func (v *approvalReporter) ReportFailoverServiceToMesh(mesh *discoveryv1alpha2.Mesh, failoverService ezkube.ResourceId, err error) {
	invalidFailoverServicesForMesh := v.unapprovedFailoverServices[mesh]
	if invalidFailoverServicesForMesh == nil {
		invalidFailoverServicesForMesh = map[string][]error{}
	}
	key := sets.Key(failoverService)
	errs := invalidFailoverServicesForMesh[key]
	errs = append(errs, err)
	invalidFailoverServicesForMesh[key] = errs
	v.unapprovedFailoverServices[mesh] = invalidFailoverServicesForMesh
}

func (v *approvalReporter) ReportFailoverService(failoverService ezkube.ResourceId, newErrs []error) {
	key := sets.Key(failoverService)
	errs := v.invalidFailoverServices[key]
	if errs == nil {
		errs = []error{}
	}
	errs = append(errs, newErrs...)
	v.invalidFailoverServices[key] = errs
}

func (v *approvalReporter) getTrafficPolicyErrors(meshService *discoveryv1alpha2.MeshService, trafficPolicy ezkube.ResourceId) []error {
	invalidTrafficPoliciesForMeshService, ok := v.unapprovedTrafficPolicies[meshService]
	if !ok {
		return nil
	}
	tpErrors, ok := invalidTrafficPoliciesForMeshService[sets.Key(trafficPolicy)]
	if !ok {
		return nil
	}
	return tpErrors
}

func (v *approvalReporter) getAccessPolicyErrors(meshService *discoveryv1alpha2.MeshService, accessPolicy ezkube.ResourceId) []error {
	invalidAccessPoliciesForMeshService, ok := v.unapprovedAccessPolicies[meshService]
	if !ok {
		return nil
	}
	apErrors, ok := invalidAccessPoliciesForMeshService[sets.Key(accessPolicy)]
	if !ok {
		return nil
	}
	return apErrors
}

func (v *approvalReporter) getFailoverServiceErrors(mesh *discoveryv1alpha2.Mesh, failoverService ezkube.ResourceId) []error {
	var errs []error
	// Mesh-dependent errors
	invalidAccessPoliciesForMeshService, ok := v.unapprovedFailoverServices[mesh]
	if ok {
		fsErrors, ok := invalidAccessPoliciesForMeshService[sets.Key(failoverService)]
		if ok {
			errs = append(errs, fsErrors...)
		}
	}

	// Mesh-independent errors
	fsErrs := v.invalidFailoverServices[sets.Key(failoverService)]
	if fsErrs != nil {
		errs = append(errs, fsErrs...)
	}
	return errs
}

func (v *approvalReporter) getVirtualMeshErrors(mesh *discoveryv1alpha2.Mesh, virtualMesh ezkube.ResourceId) []error {
	var errs []error
	// Mesh-dependent errors
	invalidAccessPoliciesForMeshService, ok := v.unapprovedVirtualMeshes[mesh]
	if ok {
		fsErrors, ok := invalidAccessPoliciesForMeshService[sets.Key(virtualMesh)]
		if ok {
			errs = append(errs, fsErrors...)
		}
	}

	return errs
}

func getAppliedTrafficPolicies(
	trafficPolicies networkingv1alpha2.TrafficPolicySlice,
	meshService *discoveryv1alpha2.MeshService,
) []*discoveryv1alpha2.MeshServiceStatus_AppliedTrafficPolicy {
	var matchingTrafficPolicies networkingv1alpha2.TrafficPolicySlice
	for _, policy := range trafficPolicies {
		if selectorutils.SelectorMatchesService(policy.Spec.DestinationSelector, meshService) {
			matchingTrafficPolicies = append(matchingTrafficPolicies, policy)
		}
	}

	sortTrafficPoliciesByAcceptedDate(meshService, matchingTrafficPolicies)

	var appliedPolicies []*discoveryv1alpha2.MeshServiceStatus_AppliedTrafficPolicy
	for _, policy := range matchingTrafficPolicies {
		policy := policy // pike
		appliedPolicies = append(appliedPolicies, &discoveryv1alpha2.MeshServiceStatus_AppliedTrafficPolicy{
			Ref:                ezkube.MakeObjectRef(policy),
			Spec:               &policy.Spec,
			ObservedGeneration: policy.Generation,
		})
	}
	return appliedPolicies
}

// sort the set of traffic policies in the order in which they were accepted.
// Traffic policies which were accepted first and have not changed (i.e. their observedGeneration is up-to-date) take precedence.
// Next are policies that were previously accepted but whose observedGeneration is out of date. This permits policies which were modified but formerly correct to maintain
// their acceptance status ahead of policies which were unomdified and previously rejected.
// Next will be the policies which have been modified and rejected.
// Finally, policies which are rejected and modified
func sortTrafficPoliciesByAcceptedDate(meshService *discoveryv1alpha2.MeshService, trafficPolicies networkingv1alpha2.TrafficPolicySlice) {
	isUpToDate := func(tp *networkingv1alpha2.TrafficPolicy) bool {
		return tp.Status.ObservedGeneration == tp.Generation
	}

	sort.SliceStable(trafficPolicies, func(i, j int) bool {
		tp1, tp2 := trafficPolicies[i], trafficPolicies[j]

		status1 := tp1.Status.MeshServices[sets.Key(meshService)]
		status2 := tp2.Status.MeshServices[sets.Key(meshService)]

		if status2 == nil {
			// if status is not set, the traffic policy is "pending" for this mesh service
			// and should get sorted after an accepted status.
			return status1 != nil
		} else if status1 == nil {
			return true
		}

		switch {
		case status1.State == networkingv1alpha2.ApprovalState_ACCEPTED:
			if status2.State != networkingv1alpha2.ApprovalState_ACCEPTED {
				// accepted comes before non accepted
				return true
			}

			if tp1UpToDate := isUpToDate(tp1); tp1UpToDate != isUpToDate(tp2) {
				// up to date is validated before modified
				return tp1UpToDate
			}

			// sort by the previous acceptance order
			return status1.AcceptanceOrder < status2.AcceptanceOrder
		case status2.State == networkingv1alpha2.ApprovalState_ACCEPTED:
			// accepted comes before non accepted
			return false
		default:
			// neither policy has been accepted, we can simply sort by unique key
			return sets.Key(tp1) < sets.Key(tp2)
		}
	})
}

// Fetch all AccessPolicies applicable to the given MeshService.
// Sorting is not needed because the additive semantics of AccessPolicies does not allow for conflicts.
func getAppliedAccessPolicies(
	accessPolicies networkingv1alpha2.AccessPolicySlice,
	meshService *discoveryv1alpha2.MeshService,
) []*discoveryv1alpha2.MeshServiceStatus_AppliedAccessPolicy {
	var appliedPolicies []*discoveryv1alpha2.MeshServiceStatus_AppliedAccessPolicy
	for _, policy := range accessPolicies {
		policy := policy // pike
		if !selectorutils.SelectorMatchesService(policy.Spec.DestinationSelector, meshService) {
			continue
		}
		appliedPolicies = append(appliedPolicies, &discoveryv1alpha2.MeshServiceStatus_AppliedAccessPolicy{
			Ref:                ezkube.MakeObjectRef(policy),
			Spec:               &policy.Spec,
			ObservedGeneration: policy.Generation,
		})
	}

	return appliedPolicies
}

func getAppliedVirtualMesh(
	virtualMeshes networkingv1alpha2.VirtualMeshSlice,
	mesh *discoveryv1alpha2.Mesh,
) *discoveryv1alpha2.MeshStatus_AppliedVirtualMesh {
	for _, vMesh := range virtualMeshes {
		vMesh := vMesh // pike
		if vMesh.Status.State != networkingv1alpha2.ApprovalState_ACCEPTED {
			continue
		}
		for _, meshRef := range vMesh.Spec.Meshes {
			if ezkube.RefsMatch(mesh, meshRef) {
				return &discoveryv1alpha2.MeshStatus_AppliedVirtualMesh{
					Ref:                ezkube.MakeObjectRef(vMesh),
					Spec:               &vMesh.Spec,
					ObservedGeneration: vMesh.Generation,
				}
			}
		}
	}
	return nil
}

// Fetch all FailoverServices applicable to the given Mesh.
func getAppliedFailoverServices(
	failoverServices networkingv1alpha2.FailoverServiceSlice,
	mesh *discoveryv1alpha2.Mesh,
) []*discoveryv1alpha2.MeshStatus_AppliedFailoverService {
	var appliedFailoverServices []*discoveryv1alpha2.MeshStatus_AppliedFailoverService
	for _, failoverService := range failoverServices {
		failoverService := failoverService // pike
		for _, meshRef := range failoverService.Spec.Meshes {
			if !ezkube.RefsMatch(meshRef, mesh) {
				continue
			}
			appliedFailoverServices = append(appliedFailoverServices, &discoveryv1alpha2.MeshStatus_AppliedFailoverService{
				Ref:                ezkube.MakeObjectRef(failoverService),
				Spec:               &failoverService.Spec,
				ObservedGeneration: failoverService.Generation,
			})
		}
	}
	return appliedFailoverServices
}
