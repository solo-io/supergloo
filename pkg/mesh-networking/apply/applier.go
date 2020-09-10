package apply

import (
	"context"
	"sort"

	discoveryv1alpha2sets "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2/sets"
	v1alpha2sets "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha2/sets"

	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation"

	"github.com/solo-io/go-utils/contextutils"
	discoveryv1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/input"
	networkingv1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/utils/selectorutils"
	"github.com/solo-io/skv2/contrib/pkg/sets"
	"github.com/solo-io/skv2/pkg/ezkube"
)

// the Applier validates user-applied configuration
// and produces a snapshot that is ready for translation (i.e. with accepted policies applied to all the Status of all targeted TrafficTargets)
// Note that the Applier also updates the statuses of objects contained in the input Snapshot.
// The Input Snapshot's SyncStatuses method should usually be called after running the Applier.
type Applier interface {
	Apply(ctx context.Context, input input.Snapshot)
}

type applier struct {
	// the applier runs the networking translator in order to detect & report translation errors
	translator translation.Translator
}

func NewApplier(
	translator translation.Translator,
) Applier {
	return &applier{
		translator: translator,
	}
}

func (v *applier) Apply(ctx context.Context, input input.Snapshot) {
	ctx = contextutils.WithLogger(ctx, "validation")
	reporter := newApplyReporter()

	virtualMeshes := input.VirtualMeshes().List()
	validateOneVirtualMeshPerMesh(virtualMeshes)

	for _, trafficTarget := range input.TrafficTargets().List() {
		trafficTarget.Status.AppliedTrafficPolicies = getAppliedTrafficPolicies(input.TrafficPolicies().List(), trafficTarget)
		trafficTarget.Status.AppliedAccessPolicies = getAppliedAccessPolicies(input.AccessPolicies().List(), trafficTarget)
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
			TrafficTargets:     map[string]*networkingv1alpha2.ApprovalStatus{},
		}
	}
	setWorkloadsForTrafficPolicies(input.TrafficPolicies(), input.Workloads())

	// initialize access policy statuses
	for _, accessPolicy := range input.AccessPolicies().List() {
		accessPolicy.Status = networkingv1alpha2.AccessPolicyStatus{
			State:              networkingv1alpha2.ApprovalState_ACCEPTED,
			ObservedGeneration: accessPolicy.Generation,
			TrafficTargets:     map[string]*networkingv1alpha2.ApprovalStatus{},
		}
	}
	setWorkloadsForAccessPolicies(input.AccessPolicies(), input.Workloads())

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

	// update traffictarget, trafficpolicy, and accesspolicy statuses
	for _, trafficTarget := range input.TrafficTargets().List() {
		trafficTarget.Status.ObservedGeneration = trafficTarget.Generation
		trafficTarget.Status.AppliedTrafficPolicies = validateAndReturnApprovedTrafficPolicies(ctx, input, reporter, trafficTarget)
		trafficTarget.Status.AppliedAccessPolicies = validateAndReturnApprovedAccessPolicies(ctx, input, reporter, trafficTarget)
	}

	for _, mesh := range input.Meshes().List() {
		mesh.Status.ObservedGeneration = mesh.Generation
		mesh.Status.AppliedFailoverServices = validateAndReturnApprovedFailoverServices(ctx, input, reporter, mesh)
		mesh.Status.AppliedVirtualMesh = validateAndReturnVirtualMesh(ctx, input, reporter, mesh)
	}

	// TODO(ilackarms): validate workloads
}

func setWorkloadsForTrafficPolicies(
	trafficPolicies v1alpha2sets.TrafficPolicySet,
	workloads discoveryv1alpha2sets.WorkloadSet) {
	for _, trafficPolicy := range trafficPolicies.List() {
		var matchingWorkloads []string
		// TODO(awang) optimize if the returned workloads list gets too large
		//if len(trafficPolicy.Spec.GetSourceSelector()) == 0 {
		//	trafficPolicy.Status.Workloads = []string{"*"}
		//	return
		//}
		for _, workload := range workloads.List() {
			if selectorutils.SelectorMatchesWorkload(trafficPolicy.Spec.GetSourceSelector(), workload) {
				matchingWorkloads = append(matchingWorkloads, sets.Key(workload))
			}
		}
		trafficPolicy.Status.Workloads = matchingWorkloads
	}
}

func setWorkloadsForAccessPolicies(
	accessPolicies v1alpha2sets.AccessPolicySet,
	workloads discoveryv1alpha2sets.WorkloadSet) {
	for _, accessPolicy := range accessPolicies.List() {
		var matchingWorkloads []string
		// TODO(awang) optimize if the returned workloads list gets too large
		for _, workload := range workloads.List() {
			if selectorutils.IdentityMatchesWorkload(accessPolicy.Spec.GetSourceSelector(), workload) {
				matchingWorkloads = append(matchingWorkloads, sets.Key(workload))
			}
		}
		accessPolicy.Status.Workloads = matchingWorkloads
	}
}

// this function both validates the status of TrafficPolicies (sets error or accepted state)
// as well as returns a list of accepted traffic policies for the traffic target status
func validateAndReturnApprovedTrafficPolicies(ctx context.Context, input input.Snapshot, reporter *applyReporter, trafficTarget *discoveryv1alpha2.TrafficTarget) []*discoveryv1alpha2.TrafficTargetStatus_AppliedTrafficPolicy {
	var validatedTrafficPolicies []*discoveryv1alpha2.TrafficTargetStatus_AppliedTrafficPolicy

	// track accepted index
	var acceptedIndex uint32
	for _, appliedTrafficPolicy := range trafficTarget.Status.AppliedTrafficPolicies {
		errsForTrafficPolicy := reporter.getTrafficPolicyErrors(trafficTarget, appliedTrafficPolicy.Ref)

		trafficPolicy, err := input.TrafficPolicies().Find(appliedTrafficPolicy.Ref)
		if err != nil {
			// should never happen
			contextutils.LoggerFrom(ctx).Errorf("internal error: failed to look up applied traffic policy %v: %v", appliedTrafficPolicy.Ref, err)
			continue
		}

		if len(errsForTrafficPolicy) == 0 {
			trafficPolicy.Status.TrafficTargets[sets.Key(trafficTarget)] = &networkingv1alpha2.ApprovalStatus{
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
			trafficPolicy.Status.TrafficTargets[sets.Key(trafficTarget)] = &networkingv1alpha2.ApprovalStatus{
				State:  networkingv1alpha2.ApprovalState_INVALID,
				Errors: errMsgs,
			}
			trafficPolicy.Status.State = networkingv1alpha2.ApprovalState_INVALID
		}
	}

	return validatedTrafficPolicies
}

// this function both validates the status of AccessPolicies (sets error or accepted state)
// as well as returns a list of accepted AccessPolicies for the traffic target status
func validateAndReturnApprovedAccessPolicies(
	ctx context.Context,
	input input.Snapshot,
	reporter *applyReporter,
	trafficTarget *discoveryv1alpha2.TrafficTarget,
) []*discoveryv1alpha2.TrafficTargetStatus_AppliedAccessPolicy {
	var validatedAccessPolicies []*discoveryv1alpha2.TrafficTargetStatus_AppliedAccessPolicy

	// track accepted index
	var acceptedIndex uint32
	for _, appliedAccessPolicy := range trafficTarget.Status.AppliedAccessPolicies {
		errsForAccessPolicy := reporter.getAccessPolicyErrors(trafficTarget, appliedAccessPolicy.Ref)

		accessPolicy, err := input.AccessPolicies().Find(appliedAccessPolicy.Ref)
		if err != nil {
			// should never happen
			contextutils.LoggerFrom(ctx).Errorf("internal error: failed to look up applied AccessPolicy %v: %v", appliedAccessPolicy.Ref, err)
			continue
		}

		if len(errsForAccessPolicy) == 0 {
			accessPolicy.Status.TrafficTargets[sets.Key(trafficTarget)] = &networkingv1alpha2.ApprovalStatus{
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
			accessPolicy.Status.TrafficTargets[sets.Key(trafficTarget)] = &networkingv1alpha2.ApprovalStatus{
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
	reporter *applyReporter,
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
	reporter *applyReporter,
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

// the applyReporter validates individual policies and reports any encountered errors
type applyReporter struct {
	// NOTE(ilackarms): map access should be synchronous (called in a single context),
	// so locking should not be necessary.
	unappliedTrafficPolicies  map[*discoveryv1alpha2.TrafficTarget]map[string][]error
	unappliedAccessPolicies   map[*discoveryv1alpha2.TrafficTarget]map[string][]error
	unappliedFailoverServices map[*discoveryv1alpha2.Mesh]map[string][]error
	unappliedVirtualMeshes    map[*discoveryv1alpha2.Mesh]map[string][]error
	invalidFailoverServices   map[string][]error
}

func newApplyReporter() *applyReporter {
	return &applyReporter{
		unappliedTrafficPolicies:  map[*discoveryv1alpha2.TrafficTarget]map[string][]error{},
		unappliedAccessPolicies:   map[*discoveryv1alpha2.TrafficTarget]map[string][]error{},
		unappliedFailoverServices: map[*discoveryv1alpha2.Mesh]map[string][]error{},
		unappliedVirtualMeshes:    map[*discoveryv1alpha2.Mesh]map[string][]error{},
		invalidFailoverServices:   map[string][]error{},
	}
}

// mark the policy with an error; will be used to filter the policy out of
// the accepted status later
func (v *applyReporter) ReportTrafficPolicyToTrafficTarget(trafficTarget *discoveryv1alpha2.TrafficTarget, trafficPolicy ezkube.ResourceId, err error) {
	invalidTrafficPoliciesForTrafficTarget := v.unappliedTrafficPolicies[trafficTarget]
	if invalidTrafficPoliciesForTrafficTarget == nil {
		invalidTrafficPoliciesForTrafficTarget = map[string][]error{}
	}
	key := sets.Key(trafficPolicy)
	errs := invalidTrafficPoliciesForTrafficTarget[key]
	errs = append(errs, err)
	invalidTrafficPoliciesForTrafficTarget[key] = errs
	v.unappliedTrafficPolicies[trafficTarget] = invalidTrafficPoliciesForTrafficTarget
}

func (v *applyReporter) ReportAccessPolicyToTrafficTarget(trafficTarget *discoveryv1alpha2.TrafficTarget, accessPolicy ezkube.ResourceId, err error) {
	invalidAccessPoliciesForTrafficTarget := v.unappliedAccessPolicies[trafficTarget]
	if invalidAccessPoliciesForTrafficTarget == nil {
		invalidAccessPoliciesForTrafficTarget = map[string][]error{}
	}
	key := sets.Key(accessPolicy)
	errs := invalidAccessPoliciesForTrafficTarget[key]
	errs = append(errs, err)
	invalidAccessPoliciesForTrafficTarget[key] = errs
	v.unappliedAccessPolicies[trafficTarget] = invalidAccessPoliciesForTrafficTarget
}

func (v *applyReporter) ReportVirtualMeshToMesh(mesh *discoveryv1alpha2.Mesh, virtualMesh ezkube.ResourceId, err error) {
	invalidVirtualMeshesForMesh := v.unappliedVirtualMeshes[mesh]
	if invalidVirtualMeshesForMesh == nil {
		invalidVirtualMeshesForMesh = map[string][]error{}
	}
	key := sets.Key(virtualMesh)
	errs := invalidVirtualMeshesForMesh[key]
	errs = append(errs, err)
	invalidVirtualMeshesForMesh[key] = errs
	v.unappliedVirtualMeshes[mesh] = invalidVirtualMeshesForMesh
}

func (v *applyReporter) ReportFailoverServiceToMesh(mesh *discoveryv1alpha2.Mesh, failoverService ezkube.ResourceId, err error) {
	invalidFailoverServicesForMesh := v.unappliedFailoverServices[mesh]
	if invalidFailoverServicesForMesh == nil {
		invalidFailoverServicesForMesh = map[string][]error{}
	}
	key := sets.Key(failoverService)
	errs := invalidFailoverServicesForMesh[key]
	errs = append(errs, err)
	invalidFailoverServicesForMesh[key] = errs
	v.unappliedFailoverServices[mesh] = invalidFailoverServicesForMesh
}

func (v *applyReporter) ReportFailoverService(failoverService ezkube.ResourceId, newErrs []error) {
	key := sets.Key(failoverService)
	errs := v.invalidFailoverServices[key]
	if errs == nil {
		errs = []error{}
	}
	errs = append(errs, newErrs...)
	v.invalidFailoverServices[key] = errs
}

func (v *applyReporter) getTrafficPolicyErrors(trafficTarget *discoveryv1alpha2.TrafficTarget, trafficPolicy ezkube.ResourceId) []error {
	invalidTrafficPoliciesForTrafficTarget, ok := v.unappliedTrafficPolicies[trafficTarget]
	if !ok {
		return nil
	}
	tpErrors, ok := invalidTrafficPoliciesForTrafficTarget[sets.Key(trafficPolicy)]
	if !ok {
		return nil
	}
	return tpErrors
}

func (v *applyReporter) getAccessPolicyErrors(trafficTarget *discoveryv1alpha2.TrafficTarget, accessPolicy ezkube.ResourceId) []error {
	invalidAccessPoliciesForTrafficTarget, ok := v.unappliedAccessPolicies[trafficTarget]
	if !ok {
		return nil
	}
	apErrors, ok := invalidAccessPoliciesForTrafficTarget[sets.Key(accessPolicy)]
	if !ok {
		return nil
	}
	return apErrors
}

func (v *applyReporter) getFailoverServiceErrors(mesh *discoveryv1alpha2.Mesh, failoverService ezkube.ResourceId) []error {
	var errs []error
	// Mesh-dependent errors
	invalidAccessPoliciesForTrafficTarget, ok := v.unappliedFailoverServices[mesh]
	if ok {
		fsErrors, ok := invalidAccessPoliciesForTrafficTarget[sets.Key(failoverService)]
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

func (v *applyReporter) getVirtualMeshErrors(mesh *discoveryv1alpha2.Mesh, virtualMesh ezkube.ResourceId) []error {
	var errs []error
	// Mesh-dependent errors
	invalidAccessPoliciesForTrafficTarget, ok := v.unappliedVirtualMeshes[mesh]
	if ok {
		fsErrors, ok := invalidAccessPoliciesForTrafficTarget[sets.Key(virtualMesh)]
		if ok {
			errs = append(errs, fsErrors...)
		}
	}

	return errs
}

func getAppliedTrafficPolicies(
	trafficPolicies networkingv1alpha2.TrafficPolicySlice,
	trafficTarget *discoveryv1alpha2.TrafficTarget,
) []*discoveryv1alpha2.TrafficTargetStatus_AppliedTrafficPolicy {
	var matchingTrafficPolicies networkingv1alpha2.TrafficPolicySlice
	for _, policy := range trafficPolicies {
		if selectorutils.SelectorMatchesService(policy.Spec.DestinationSelector, trafficTarget) {
			matchingTrafficPolicies = append(matchingTrafficPolicies, policy)
		}
	}

	sortTrafficPoliciesByAcceptedDate(trafficTarget, matchingTrafficPolicies)

	var appliedPolicies []*discoveryv1alpha2.TrafficTargetStatus_AppliedTrafficPolicy
	for _, policy := range matchingTrafficPolicies {
		policy := policy // pike
		appliedPolicies = append(appliedPolicies, &discoveryv1alpha2.TrafficTargetStatus_AppliedTrafficPolicy{
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
func sortTrafficPoliciesByAcceptedDate(trafficTarget *discoveryv1alpha2.TrafficTarget, trafficPolicies networkingv1alpha2.TrafficPolicySlice) {
	isUpToDate := func(tp *networkingv1alpha2.TrafficPolicy) bool {
		return tp.Status.ObservedGeneration == tp.Generation
	}

	sort.SliceStable(trafficPolicies, func(i, j int) bool {
		tp1, tp2 := trafficPolicies[i], trafficPolicies[j]

		status1 := tp1.Status.TrafficTargets[sets.Key(trafficTarget)]
		status2 := tp2.Status.TrafficTargets[sets.Key(trafficTarget)]

		if status2 == nil {
			// if status is not set, the traffic policy is "pending" for this traffic target
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

// Fetch all AccessPolicies applicable to the given TrafficTarget.
// Sorting is not needed because the additive semantics of AccessPolicies does not allow for conflicts.
func getAppliedAccessPolicies(
	accessPolicies networkingv1alpha2.AccessPolicySlice,
	trafficTarget *discoveryv1alpha2.TrafficTarget,
) []*discoveryv1alpha2.TrafficTargetStatus_AppliedAccessPolicy {
	var appliedPolicies []*discoveryv1alpha2.TrafficTargetStatus_AppliedAccessPolicy
	for _, policy := range accessPolicies {
		policy := policy // pike
		if !selectorutils.SelectorMatchesService(policy.Spec.DestinationSelector, trafficTarget) {
			continue
		}
		appliedPolicies = append(appliedPolicies, &discoveryv1alpha2.TrafficTargetStatus_AppliedAccessPolicy{
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
