package approval

import (
	"context"
	"sort"

	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation"

	"github.com/solo-io/go-utils/contextutils"
	discoveryv1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/input"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha2"
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

	for _, meshService := range input.MeshServices().List() {
		meshService.Status.AppliedTrafficPolicies = getAppliedTrafficPolicies(input.TrafficPolicies().List(), meshService)

		meshService.Status.AppliedAccessPolicies = getAppliedAccessPolicies(input.AccessPolicies().List(), meshService)
	}
	for _, mesh := range input.Meshes().List() {
		mesh.Status.AppliedVirtualMeshes = getAppliedVirtualMeshes(input.VirtualMeshes().List(), mesh)
	}

	for _, mesh := range input.Meshes().List() {
		appliedFailoverServices := getAppliedFailoverServices(input.FailoverServices().List(), mesh)
		mesh.Status.AppliedFailoverServices = appliedFailoverServices
	}

	_, err := v.translator.Translate(ctx, input, reporter)
	if err != nil {
		// should never happen
		contextutils.LoggerFrom(ctx).Errorf("internal error: failed to run translator: %v", err)
	}

	// initialize traffic policy statuses
	for _, trafficPolicy := range input.TrafficPolicies().List() {
		trafficPolicy.Status = v1alpha2.TrafficPolicyStatus{
			State:              v1alpha2.ApprovalState_ACCEPTED,
			ObservedGeneration: trafficPolicy.Generation,
			MeshServices:       map[string]*v1alpha2.ApprovalStatus{},
		}
	}

	// initialize access policy statuses
	for _, accessPolicy := range input.AccessPolicies().List() {
		accessPolicy.Status = v1alpha2.AccessPolicyStatus{
			State:              v1alpha2.ApprovalState_ACCEPTED,
			ObservedGeneration: accessPolicy.Generation,
			MeshServices:       map[string]*v1alpha2.ApprovalStatus{},
		}
	}

	// write FailoverService statuses
	for _, failoverService := range input.FailoverServices().List() {
		failoverService.Status = v1alpha2.FailoverServiceStatus{
			State:              v1alpha2.ApprovalState_ACCEPTED,
			ObservedGeneration: failoverService.Generation,
			Meshes:             map[string]*v1alpha2.ApprovalStatus{},
		}
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
			trafficPolicy.Status.MeshServices[sets.Key(meshService)] = &v1alpha2.ApprovalStatus{
				AcceptanceOrder: acceptedIndex,
				State:           v1alpha2.ApprovalState_ACCEPTED,
			}
			validatedTrafficPolicies = append(validatedTrafficPolicies, appliedTrafficPolicy)
			acceptedIndex++
		} else {
			var errMsgs []string
			for _, tpErr := range errsForTrafficPolicy {
				errMsgs = append(errMsgs, tpErr.Error())
			}
			trafficPolicy.Status.MeshServices[sets.Key(meshService)] = &v1alpha2.ApprovalStatus{
				State:  v1alpha2.ApprovalState_INVALID,
				Errors: errMsgs,
			}
			trafficPolicy.Status.State = v1alpha2.ApprovalState_INVALID
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
			accessPolicy.Status.MeshServices[sets.Key(meshService)] = &v1alpha2.ApprovalStatus{
				AcceptanceOrder: acceptedIndex,
				State:           v1alpha2.ApprovalState_ACCEPTED,
			}
			validatedAccessPolicies = append(validatedAccessPolicies, appliedAccessPolicy)
			acceptedIndex++
		} else {
			var errMsgs []string
			for _, apErr := range errsForAccessPolicy {
				errMsgs = append(errMsgs, apErr.Error())
			}
			accessPolicy.Status.MeshServices[sets.Key(meshService)] = &v1alpha2.ApprovalStatus{
				State:  v1alpha2.ApprovalState_INVALID,
				Errors: errMsgs,
			}
			accessPolicy.Status.State = v1alpha2.ApprovalState_INVALID
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
			failoverService.Status.Meshes[sets.Key(mesh)] = &v1alpha2.ApprovalStatus{
				AcceptanceOrder: acceptedIndex,
				State:           v1alpha2.ApprovalState_ACCEPTED,
			}
			validatedFailoverServices = append(validatedFailoverServices, appliedFailoverService)
			acceptedIndex++
		} else {
			var errMsgs []string
			for _, fsErr := range errsForFailoverService {
				errMsgs = append(errMsgs, fsErr.Error())
			}
			failoverService.Status.Meshes[sets.Key(mesh)] = &v1alpha2.ApprovalStatus{
				State:  v1alpha2.ApprovalState_INVALID,
				Errors: errMsgs,
			}
			failoverService.Status.State = v1alpha2.ApprovalState_INVALID
		}
	}

	return validatedFailoverServices
}

// the approval reporter uses istioTranslator reports to
// validate individual policies
type approvalReporter struct {
	// NOTE(ilackarms): map access should be synchronous (called in a single context),
	// so locking should not be necessary.
	unapprovedTrafficPolicies  map[*discoveryv1alpha2.MeshService]map[string][]error
	unapprovedAccessPolicies   map[*discoveryv1alpha2.MeshService]map[string][]error
	unapprovedFailoverServices map[*discoveryv1alpha2.Mesh]map[string][]error
	invalidFailoverServices    map[string][]error
}

func newApprovalReporter() *approvalReporter {
	return &approvalReporter{
		unapprovedTrafficPolicies:  map[*discoveryv1alpha2.MeshService]map[string][]error{},
		unapprovedAccessPolicies:   map[*discoveryv1alpha2.MeshService]map[string][]error{},
		unapprovedFailoverServices: map[*discoveryv1alpha2.Mesh]map[string][]error{},
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
	// TODO(ilackarms):
	panic("implement me")
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

func getAppliedTrafficPolicies(
	trafficPolicies v1alpha2.TrafficPolicySlice,
	meshService *discoveryv1alpha2.MeshService,
) []*discoveryv1alpha2.MeshServiceStatus_AppliedTrafficPolicy {
	var matchingTrafficPolicies v1alpha2.TrafficPolicySlice
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
func sortTrafficPoliciesByAcceptedDate(meshService *discoveryv1alpha2.MeshService, trafficPolicies v1alpha2.TrafficPolicySlice) {
	isUpToDate := func(tp *v1alpha2.TrafficPolicy) bool {
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
		}

		switch {
		case status1.State == v1alpha2.ApprovalState_ACCEPTED:
			if status2.State != v1alpha2.ApprovalState_ACCEPTED {
				// accepted comes before non accepted
				return true
			}

			if tp1UpToDate := isUpToDate(tp1); tp1UpToDate != isUpToDate(tp2) {
				// up to date is validated before modified
				return tp1UpToDate
			}

			// sort by the previous acceptance order
			return status1.AcceptanceOrder < status2.AcceptanceOrder
		case status2.State == v1alpha2.ApprovalState_ACCEPTED:
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
	accessPolicies v1alpha2.AccessPolicySlice,
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

func getAppliedVirtualMeshes(
	virtualMeshes v1alpha2.VirtualMeshSlice,
	mesh *discoveryv1alpha2.Mesh,
) []*discoveryv1alpha2.MeshStatus_AppliedVirtualMesh {
	var matchingVirtualMeshes v1alpha2.VirtualMeshSlice
	for _, vMesh := range virtualMeshes {
		for _, meshRef := range vMesh.Spec.Meshes {
			if ezkube.RefsMatch(mesh, meshRef) {
				matchingVirtualMeshes = append(matchingVirtualMeshes, vMesh)
				break
			}
		}
	}

	sortVirtualMeshesByAcceptedDate(mesh, matchingVirtualMeshes)

	var appliedVirtualMeshes []*discoveryv1alpha2.MeshStatus_AppliedVirtualMesh
	for _, vMesh := range matchingVirtualMeshes {
		vMesh := vMesh // pike
		appliedVirtualMeshes = append(appliedVirtualMeshes, &discoveryv1alpha2.MeshStatus_AppliedVirtualMesh{
			Ref:                ezkube.MakeObjectRef(vMesh),
			Spec:               &vMesh.Spec,
			ObservedGeneration: vMesh.Generation,
		})
	}
	return appliedVirtualMeshes
}

// Fetch all FailoverServices applicable to the given Mesh.
func getAppliedFailoverServices(
	failoverServices v1alpha2.FailoverServiceSlice,
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

// sort the set of virtual meshes in the order in which they were accepted.
// VMeshes which were accepted first and have not changed (i.e. their observedGeneration is up-to-date) take precedence.
// Next are vMeshes that were previously accepted but whose observedGeneration is out of date. This permits vmeshes which were modified but formerly correct to maintain
// their acceptance status ahead of vmeshes which were unomdified and previously rejected.
// Next will be the vmeshes which have been modified and rejected.
// Finally, vmeshes which are rejected and modified
func sortVirtualMeshesByAcceptedDate(mesh *discoveryv1alpha2.Mesh, virtualMeshes v1alpha2.VirtualMeshSlice) {
	isUpToDate := func(vm *v1alpha2.VirtualMesh) bool {
		return vm.Status.ObservedGeneration == vm.Generation
	}

	sort.SliceStable(virtualMeshes, func(i, j int) bool {
		vMesh1, vMesh2 := virtualMeshes[i], virtualMeshes[j]

		status1 := vMesh1.Status.Meshes[sets.Key(mesh)]
		status2 := vMesh2.Status.Meshes[sets.Key(mesh)]

		if status2 == nil {
			// if status is not set, the vMesh is "pending" for this mesh
			// and should get sorted after an accepted status.
			return status1 != nil
		}

		switch {
		case status1.State == v1alpha2.ApprovalState_ACCEPTED:
			if status2.State != v1alpha2.ApprovalState_ACCEPTED {
				// accepted comes before non accepted
				return true
			}

			if vMesh1UpToDate := isUpToDate(vMesh1); vMesh1UpToDate != isUpToDate(vMesh2) {
				// up to date is validated before modified
				return vMesh1UpToDate
			}

			// sort by the previous acceptance order
			return status1.AcceptanceOrder < status2.AcceptanceOrder
		case status2.State == v1alpha2.ApprovalState_ACCEPTED:
			// accepted comes before non accepted
			return false
		default:
			// neither policy has been accepted, we can simply sort by unique key
			return sets.Key(vMesh1) < sets.Key(vMesh2)
		}
	})
}
